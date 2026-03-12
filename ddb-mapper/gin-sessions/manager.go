package sessions

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/csrf"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/gbac"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
	"github.com/nguyengg/go-aws-commons/opaque-token/hmac"
	ini "github.com/nguyengg/init-once"
)

// Manager manages sessions of struct type T.
//
// The zero-value is ready for use.
type Manager[T any] struct {
	// Client is the client for making DynamoDB service calls.
	Client *dynamodb.Client

	mapper *mapper.Mapper[T]

	sessionIdCookieName  string
	sessionCookieOptions func(c *http.Cookie)
	newSessionId         func() string

	csrfSignVerifier hmac.Engine
	csrfOpts         csrf.Options

	getGroupsFn gbac.GetGroupsFunc
	groupsOpts  gbac.Options

	// once guards init.
	once ini.SuccessOnce
}

// manager is used by package-level methods which are not type-aware.
//
// [Manager] implements this interface; its exported methods delegate to these unexpored methods.
type manager interface {
	get(c *gin.Context) (any, error)
	tryGet(c *gin.Context) (any, error)
	regenerate(c *gin.Context) (any, error)
	Save(c *gin.Context) error
	Destroy(c *gin.Context) error
}

// New creates a new [Manager] of type T.
//
// Struct T must have these struct tags:
//
//	// Hash key is required, and its type must be a string since only string session Ids are supported.
//	Field string `dynamodbav:"sessionId,hashkey" tableName:"my-table"`
//
// See [github.com/nguyengg/go-aws-commons/ddb-mapper] for more information on how the struct tags are parsed.
// If type T does not implement the required tags or the tags fail validation, a non-nil error is returned.
//
// If you have multiple sessions from different cookies, [Config.SessionIdCookieName] differentiates them.
//
// To enable CSRF generation (and validation), pass [WithCSRF] as an option.
//
// [github.com/nguyengg/go-aws-commons/ddb-mapper]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb-mapper
func New[T any](optFns ...func(cfg *Config)) (*Manager[T], error) {
	cfg := Config{
		SessionIdCookieName: DefaultSessionIdCookieName,
		NewSessionId:        DefaultNewSessionId,
	}
	for _, fn := range optFns {
		fn(&cfg)
	}
	if cfg.SessionIdCookieName == "" {
		cfg.SessionIdCookieName = DefaultSessionIdCookieName
	}
	if cfg.NewSessionId == nil {
		cfg.NewSessionId = DefaultNewSessionId
	}

	m, err := mapper.New[T](cfg.mapperOpts...)
	if err != nil {
		return nil, err
	}

	if t := m.HashKey.Type; !t.ConvertibleTo(reflect.TypeFor[string]()) {
		return nil, fmt.Errorf("only string hashkeys are supported at the moment; struct type %T has %s hashkey instead", m.StructType, t)
	}

	return &Manager[T]{
		Client:               cfg.Client,
		mapper:               m,
		sessionIdCookieName:  cfg.SessionIdCookieName,
		sessionCookieOptions: cfg.SessionCookieOptions,
		newSessionId:         cfg.NewSessionId,
	}, nil
}

// Get returns the current session.
//
// If there are no sessions attached, [Get] implicitly creates a new one. Use [TryGet] if you don't want this behaviour;
// for example, an authentication middleware will want to verify that existing session exists with a valid user.
//
// [Get] and [Regenerate] do not automatically save new session to DynamoDB. You must explicitly call [Save] to do so.
// If [Get] creates a new session, it will automatically issue Set-Cookie response header for the new session Id and
// (optionally) the new CSRF token.
func (m *Manager[T]) Get(c *gin.Context) (*T, error) {
	v, err := m.get(c)
	if err != nil {
		return nil, err
	}

	return v.(*T), nil
}

func (m *Manager[T]) get(c *gin.Context) (any, error) {
	if err := m.init(); err != nil {
		return nil, err
	}

	// load from ddb or context first to see if session exists. if it does, return it. if it does not (which can be due
	// to absence of session cookie, or session does not exist in ddb), then regenerate to create a new session.
	s, err := m.load(c)
	if err != nil {
		return nil, err
	}
	if s != nil {
		return s.v.(*T), nil
	}
	if s, err = m.doRegenerate(c, nil); err != nil {
		return nil, err
	}

	return s.v.(*T), nil
}

func (m *Manager[T]) getSid(c *gin.Context) (string, error) {
	if err := m.init(); err != nil {
		return "", err
	}

	s, err := m.load(c)
	if err != nil {
		return "", err
	}
	if s == nil {
		if s, err = m.doRegenerate(c, nil); err != nil {
			return "", err
		}
	}

	return s.sid, nil
}

// TryGet is a variant of [Get] that does not automatically create a new session.
func (m *Manager[T]) TryGet(c *gin.Context) (*T, error) {
	v, err := m.tryGet(c)
	if err != nil || v == nil {
		return nil, err
	}

	return v.(*T), nil
}

func (m *Manager[T]) tryGet(c *gin.Context) (any, error) {
	if err := m.init(); err != nil {
		return nil, err
	}

	s, err := m.load(c)
	if s != nil && err == nil {
		return s.v.(*T), nil
	}

	return nil, err
}

// Regenerate will always create a new session Id, but it may reuse existing session metadata.
//
// [Get] and [Regenerate] do not automatically save new session to DynamoDB. You must explicitly call [Save] to do so.
// [Regenerate] always issues Set-Cookie response header for the new session Id and (optionally) the new CSRF token.
func (m *Manager[T]) Regenerate(c *gin.Context) (*T, error) {
	v, err := m.regenerate(c)
	if err != nil {
		return nil, err
	}

	return v.(*T), nil
}

func (m *Manager[T]) regenerate(c *gin.Context) (any, error) {
	if err := m.init(); err != nil {
		return nil, err
	}

	// load from ddb or context first to see if there exists a session to copy; regenerate can reuse the found *session.
	s, err := m.load(c)
	if err != nil {
		return nil, err
	}
	if s, err = m.doRegenerate(c, s); err != nil {
		return nil, err
	}

	return s.v.(*T), nil
}

// Save saves session metadata to DynamoDB.
//
// Use this if you have made changes to the session metadata and need to commit them. If there are no sessions
// attached, [Save] will create and store a "zero-value" session.
func (m *Manager[T]) Save(c *gin.Context) error {
	if err := m.init(); err != nil {
		return err
	}

	// load from ddb or context first so that if no session exists, regenerate can be used to create one.
	s, err := m.load(c)
	if err != nil {
		return err
	}
	if s == nil {
		if s, err = m.doRegenerate(c, nil); err != nil {
			return err
		}
	}

	if _, err = m.mapper.Put(c, s.v.(*T), func(opts *config.PutOptions) {
		if m.Client != nil {
			opts.Client = m.Client
		}
	}); err != nil {
		return fmt.Errorf("sessions: save session to DynamoDB error: %w", err)
	}

	return nil
}

// Destroy removes the session item from DynamoDB.
//
// The response will have Set-Cookie headers to delete the session and CSRF cookies as well.
func (m *Manager[T]) Destroy(c *gin.Context) error {
	if err := m.init(); err != nil {
		return err
	}

	s, err := m.load(c)
	if s == nil || err != nil {
		return err
	}

	if _, err = m.mapper.Delete(c, s.v.(*T), func(opts *config.DeleteOptions) {
		if m.Client != nil {
			opts.Client = m.Client
		}
	}); err != nil {
		return fmt.Errorf("sessions: delete session from DynamoDB error: %w", err)
	}

	// nullify and unset just in case there are dangling references.
	s.sid = ""
	s.v = nil
	s.unset(c, m.sessionIdCookieName)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.sessionIdCookieName,
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	})

	if m.csrfSignVerifier != nil {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     m.csrfOpts.CookieName,
			Value:    "",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: false,
			SameSite: http.SameSiteDefaultMode,
		})
	}

	return nil
}

// Middleware returns the gin middleware that will configure package-level accessors to use this manager's settings.
//
// The middleware returned by this method can be used in multiple handler chains.
func (m *Manager[T]) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(managerKeyPrefix+m.sessionIdCookieName, m)
		c.Next()
	}
}

// load uses session Id cookie to load session metadata from DynamoDB or from context set therein.
func (m *Manager[T]) load(c *gin.Context) (*session, error) {
	if s, ok := get(c, m.sessionIdCookieName); ok {
		return s, nil
	}

	sid, err := c.Cookie(m.sessionIdCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		return nil, nil
	}

	s := &session{}
	if err = m.setSid(s, sid); err != nil {
		return nil, err
	}

	if output, err := m.mapper.Get(c, s.v.(*T), func(opts *config.GetOptions) {
		if m.Client != nil {
			opts.Client = m.Client
		}
	}); err != nil {
		return nil, fmt.Errorf("sessions: get session from DynamoDB error: %w", err)
	} else if len(output.Item) == 0 {
		// session Id cookie references a non-existent session.
		return nil, nil
	}

	s.set(c, m.sessionIdCookieName)
	return s, nil
}

// doRegenerate implements Manager.Regenerate.
//
// The s argument, if non-nil, means there exists a session that should be copied.
func (m *Manager[T]) doRegenerate(c *gin.Context, s *session) (*session, error) {
	if s == nil {
		s = &session{}
		s.set(c, m.sessionIdCookieName)
	}

	sid := m.newSessionId()
	if err := m.setSid(s, sid); err != nil {
		return nil, err
	} else {
		// we must reset version and timestamps as well.
		for _, attr := range []*model.Attribute{m.mapper.Version, m.mapper.CreatedTime, m.mapper.ModifiedTime} {
			if attr == nil {
				continue
			}

			if err = attr.Reset(s.v); err != nil {
				return nil, fmt.Errorf("setting %s to zero value error: %w", m.mapper.Version, err)
			}
		}
	}

	if m.csrfSignVerifier != nil {
		token, err := m.csrfSignVerifier.Sign(c, []byte(s.sid), 16)
		if err != nil {
			return nil, fmt.Errorf("sessions: create CSRF token error: %w", err)
		}

		m.writeCSRFCookie(c, base64.RawURLEncoding.EncodeToString(token))
	}

	m.writeSessionCookie(c, sid)

	return s, nil
}

func (m *Manager[T]) setSid(s *session, sid string) (err error) {
	s.sid = sid

	if s.v != nil {
		if err = m.mapper.HashKey.Set(s.v, sid); err != nil {
			return fmt.Errorf("setting %s error: %w", m.mapper.HashKey, err)
		}

		return nil
	}

	n := reflect.New(m.mapper.StructType)
	s.v = n.Interface()
	if err = m.mapper.HashKey.Set(s.v, sid); err != nil {
		return fmt.Errorf("setting %s error: %w", m.mapper.HashKey, err)
	}

	return nil
}

func (m *Manager[T]) writeSessionCookie(c *gin.Context, sid string) {
	cookie := &http.Cookie{
		Name:     m.sessionIdCookieName,
		Value:    sid,
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}

	if m.sessionCookieOptions != nil {
		m.sessionCookieOptions(cookie)
	}

	http.SetCookie(c.Writer, cookie)
}

func (m *Manager[T]) writeCSRFCookie(c *gin.Context, value string) {
	cookie := &http.Cookie{
		Name:     m.csrfOpts.CookieName,
		Value:    value,
		MaxAge:   0,
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteDefaultMode,
	}

	if m.csrfOpts.CookieOptions != nil {
		m.csrfOpts.CookieOptions(cookie)
	}

	http.SetCookie(c.Writer, cookie)
}

func (m *Manager[T]) init() error {
	return m.once.Do(func() (err error) {
		if m.mapper == nil {
			if m.mapper, err = mapper.New[T](); err != nil {
				return
			}
		}

		if m.sessionIdCookieName == "" {
			m.sessionIdCookieName = DefaultSessionIdCookieName
		}
		if m.newSessionId == nil {
			m.newSessionId = DefaultNewSessionId
		}

		if m.csrfSignVerifier != nil {
			if m.csrfOpts.CookieName == "" {
				m.csrfOpts.CookieName = csrf.DefaultCookieName
			}
		}

		return nil
	})
}

func getManager(c *gin.Context, name ...string) (manager, error) {
	var n string
	if len(name) != 0 {
		n = name[0]
	} else {
		n = DefaultSessionIdCookieName
	}

	if v, ok := c.Get(managerKeyPrefix + n); ok {
		return v.(manager), nil
	}

	panic(fmt.Sprintf(`no Manager.Middleware found with name="%s"`, n))
}

var _ manager = &Manager[any]{}
