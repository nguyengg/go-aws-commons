// Package sessions provides sessions management middleware and helper methods with DynamoDB integration via
// [github.com/nguyengg/go-aws-commons/ddb] and CSRF generation with [github.com/nguyengg/go-aws-commons/opaque-token/hmac].
//
// There are two ways to use this package. First is by constructing a manager with New, then accessing the session
// struct by way of the Manager interface.
//
// TODO add usage.
//
// Second is by using the package-level methods Get, Regenerate, Save, and Destroy, counterparts to the Manager methods
// with same name. The second usage pattern is useful if you're writing middlewares or handlers that do not have access
// to the Manager instance. To customise session settings in this mode, you should still create a manager, then add
// Manager.Middleware prior to your handler in the chain. Without this, the package-level methods will assume all
// default settings without the ability to generate CSRF tokens.
//
// [github.com/nguyengg/go-aws-commons/ddb]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb
// [github.com/nguyengg/go-aws-commons/opaque-token/hmac]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/opaque-token/hmac
package sessions

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
	"github.com/nguyengg/go-aws-commons/ddb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
	"github.com/nguyengg/go-aws-commons/opaque-token/hmac"
	ini "github.com/nguyengg/init-once"
)

// Manager manages sessions of type T.
//
// The zero-value is ready for use without CSRF generation.
type Manager[T any] struct {
	// SessionIdCookieName is the name of the cookie that contains session Id.
	//
	// Default to DefaultSessionIdCookieName.
	SessionIdCookieName string

	// NewSessionId is used to create the Id for a new session.
	//
	// Defaults to DefaultNewSessionId. You can replace it with uuid.NewString for example.
	NewSessionId func() string

	// CSRFCookieName is the name of the cookie that contains CSRF token.
	//
	// Defaults to DefaultCSRFCookieName. This cookie may be used as a source to validate CSRF token. If NewWithCSRF was
	// used to create manager then the cookie is also used to send generated CSRF tokens to user.
	CSRFCookieName string

	// SessionCookieOptions can be used to modify the session cookie prior to setting the Set-Cookie response header.
	//
	// Invalid settings will cause the cookie to be silent dropped so be very careful with this. Most likely you just
	// want to change the [http.Cookie.MaxAge] to something more reasonable.
	SessionCookieOptions func(c *http.Cookie)

	// CSRFCookieOptions can be used to modify the session cookie prior to setting the Set-Cookie response header.
	//
	// Invalid settings will cause the cookie to be silent dropped so be very careful with this. Most likely you just
	// want to change the [http.Cookie.MaxAge] to something more reasonable.
	CSRFCookieOptions func(c *http.Cookie)

	// Client is the DynamoDB client for saving session data.
	//
	// By default, `configcache.Get` will be used to provide an instance.
	Client ddb.ManagerAPIClient

	mapper           *mapper.Mapper[T]
	csrfSignVerifier hmac.Hasher
	csrfValue        string

	// once guards init.
	once ini.SuccessOnce
}

// New creates a new Manager of type T and given cookie name.
//
// Struct of type T must have these struct tags:
//
//	// Hash key is required, and its type must be a string since only string session Ids are supported.
//	Field string `dynamodbav:"sessionId,hashkey" tableName:"my-table"`
//
// See ddb.Table from [github.com/nguyengg/go-aws-commons/ddb] for more information on how the struct tags are parsed.
// If type T does not implement the required tags or the tags fail validation, a non-nil error is returned.
//
// If you have multiple sessions from different cookies, the name argument differentiates them. Pass
// DefaultSessionIdCookieName for the name argument if you don't have a better value.
//
// To enable CSRF generation, pass WithCSRF as an option.
//
// [github.com/nguyengg/go-aws-commons/ddb]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb
func New[T any](name string, optFns ...func(cfg *Config)) (Manager[T], error) {
	cfg := Config{SessionIdCookieName: name}
	for _, fn := range optFns {
		fn(&cfg)
	}

	return newManager[T](cfg)
}

// newManager returns the specific manager instead of Manager interface.
func newManager[T any](cfg Config) (*manager[T], error) {
	t := reflect.TypeFor[T]()
	table, err := ddb.NewTable(t, func(options *ddb.TableOptions) {
		options.Validator = func(attr *ddb.Attribute) error {
			// hashkey must be string.
			tags := strings.Split(attr.Tag.Get("dynamodbav"), ",")
			if len(tags) > 1 {
				switch tags[1] {
				case "hashkey":
					if attr.Type.Kind() != reflect.String {
						return fmt.Errorf("hashkey of type %s is not supported; must be string", attr.Type)
					}
				}
			}

			return nil
		}
	})
	if err != nil {
		return nil, fmt.Errorf("sessions: create sessions for type %s error: %w", t, err)
	}

	m := &manager[T]{
		config: cfg,
		table:  table,
		t:      t,
	}
	if cfg.Client != nil {
		m.ddb = ddb.NewManager(cfg.Client)
	}
	return m, nil
}

// Manager manages sessions of type T.
//
// TODO add usage.
type iManager[T any] interface {
	// Get returns the current session.
	//
	// If there are no sessions attached, Get implicitly creates a new one. Use TryGet if you don't want this behaviour;
	// for example, an authentication middleware will want to verify that existing session exists with a valid user.
	//
	// Get and Regenerate do not automatically save new session to DynamoDB. You must explicitly call Save to do so.
	// If Get creates a new session, it will automatically issue Set-Cookie response header for the new session Id and
	// (optionally) the new CSRF token.
	Get(c *gin.Context) (*T, error)

	// TryGet is a variant of Get that does not automatically create a new session.
	TryGet(c *gin.Context) (_ *T, err error)

	// Regenerate will always create a new session Id, but it may reuse existing session metadata.
	//
	// Get and Regenerate do not automatically save new session to DynamoDB. You must explicitly call Save to do so.
	// Regenerate always issues Set-Cookie response header for the new session Id and (optionally) the new CSRF token.
	Regenerate(c *gin.Context) (_ *T, err error)

	// Save saves session metadata to DynamoDB.
	//
	// Use this if you have made changes to the session metadata and need to commit them. If there are no sessions
	// attached, Save will create and store a "zero-value" session.
	Save(c *gin.Context) error

	// Destroy removes the session item from DynamoDB.
	//
	// The response will have Set-Cookie headers to delete the session and CSRF cookies as well.
	Destroy(c *gin.Context) error

	// Middleware returns the gin middleware that will configure package-level accessors to use this manager's settings.
	//
	// The middleware returned by this method can be used in multiple handler chains.
	Middleware() gin.HandlerFunc
}

// manager implements Manager.
type manager[T any] struct {
	config
	ddb       *ddb.Manager
	table     *ddb.Table
	t         reflect.Type
	csrfValue string

	// once guards init.
	once ini.SuccessOnce
}

var _ Manager[any] = &manager[any]{}
var _ Manager[any] = (*manager[any])(nil)

func (m *manager[T]) Get(c *gin.Context) (*T, error) {
	if err := m.init(c); err != nil {
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
	if s, err = m.regenerate(c, nil); err != nil {
		return nil, err
	}

	return s.v.(*T), nil
}

// TryGet is a variant of Get that does not automatically create a new session.
func (m *manager[T]) TryGet(c *gin.Context) (*T, error) {
	if err := m.init(c); err != nil {
		return nil, err
	}

	s, err := m.load(c)
	if s != nil && err == nil {
		return s.v.(*T), nil
	}

	return nil, err
}

func (m *manager[T]) Regenerate(c *gin.Context) (*T, error) {
	if err := m.init(c); err != nil {
		return nil, err
	}

	// load from ddb or context first to see if there exists a session to copy; regenerate can reuse the found *session.
	s, err := m.load(c)
	if err != nil {
		return nil, err
	}
	if s, err = m.regenerate(c, s); err != nil {
		return nil, err
	}

	return s.v.(*T), nil
}

func (m *manager[T]) Save(c *gin.Context) error {
	if err := m.init(c); err != nil {
		return err
	}

	// load from ddb or context first so that if no session exists, regenerate can be used to create one.
	s, err := m.load(c)
	if err != nil {
		return err
	}
	if s == nil {
		if s, err = m.regenerate(c, nil); err != nil {
			return err
		}
	}

	// TODO unfortunately, Put will not update s.v so we don't get the new version here.
	if putItemOutput, err := m.ddb.Put(c, s.v); err != nil {
		return fmt.Errorf("sessions: save session to DynamoDB error: %w", err)
	} else if err = attributevalue.UnmarshalMap(putItemOutput.Attributes, s.v); err != nil {
		return fmt.Errorf("sessions: unmarshal DynamoDB PutItem attributes error: %w", err)
	}

	return nil
}

func (m *manager[T]) Destroy(c *gin.Context) error {
	if err := m.init(c); err != nil {
		return err
	}

	s, err := m.load(c)
	if s == nil || err != nil {
		return err
	}

	if _, err = m.ddb.Delete(c, s.v); err != nil {
		return fmt.Errorf("sessions: delete session from DynamoDB error: %w", err)
	}

	// nullify and unset just in case there are dangling references.
	s.sid = ""
	s.v = nil
	s.unset(c, m.SessionIdCookieName)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.SessionIdCookieName,
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	})

	if m.csrfSignVerifier != nil {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     m.CSRFCookieName,
			Value:    "",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: false,
			SameSite: http.SameSiteDefaultMode,
		})
	}

	return nil
}

func (m *manager[T]) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(managerKeyPrefix+m.SessionIdCookieName, m)
		c.Next()
	}
}

// load uses session Id cookie to load session metadata from DynamoDB or from context set therein.
func (m *manager[T]) load(c *gin.Context) (*session, error) {
	if s, ok := get(c, m.SessionIdCookieName); ok {
		return s, nil
	}

	sid, err := c.Cookie(m.SessionIdCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		return nil, nil
	}

	s := &session{}
	if err = m.setSid(s, sid); err != nil {
		return nil, err
	}

	if output, err := m.ddb.Get(c, s.v, s.v); err != nil {
		return nil, fmt.Errorf("sessions: get session from DynamoDB error: %w", err)
	} else if len(output.Item) == 0 {
		// session Id cookie references a non-existent session.
		return nil, nil
	}

	s.set(c, m.SessionIdCookieName)
	return s, nil
}

// regenerate implements Manager.Regenerate.
//
// The s argument, if non-nil, means there exists a session that should be copied.
func (m *manager[T]) regenerate(c *gin.Context, s *session) (*session, error) {
	if s == nil {
		s = &session{}
		s.set(c, m.SessionIdCookieName)
	}

	sid := m.NewSessionId()
	if err := m.setSid(s, sid); err != nil {
		return nil, err
	}

	if m.csrfSignVerifier != nil {
		token, err := m.csrfSignVerifier.Sign(c, []byte(s.sid), 16)
		if err != nil {
			return nil, fmt.Errorf("sessions: create CSRF token error: %w", err)
		}

		m.csrfValue = base64.RawURLEncoding.EncodeToString(token)
		m.writeCSRFCookie(c)
	}

	m.writeSessionCookie(c, sid)

	return s, nil
}

func (m *manager[T]) setSid(s *session, sid string) (err error) {
	s.sid = sid

	var pkValue reflect.Value
	if itemValue := reflect.ValueOf(s.v); itemValue.IsValid() {
		if pkValue, err = m.table.HashKey.GetFieldValue(itemValue.Elem()); err != nil {
			return fmt.Errorf("sessions: get hash key field value error: %w", err)
		}
	} else {
		newItemV := reflect.New(m.t)
		s.v = newItemV.Interface()
		if pkValue, err = m.table.HashKey.GetFieldValue(newItemV.Elem()); err != nil {
			return fmt.Errorf("sessions: get new hash key field value error: %w", err)
		}
	}

	pkValue.SetString(sid)

	return nil
}

func (m *manager[T]) writeSessionCookie(c *gin.Context, sid string) {
	cookie := &http.Cookie{
		Name:     m.SessionIdCookieName,
		Value:    sid,
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}

	if m.SessionCookieOptions != nil {
		m.SessionCookieOptions(cookie)
	}

	http.SetCookie(c.Writer, cookie)
}

func (m *manager[T]) writeCSRFCookie(c *gin.Context) {
	cookie := &http.Cookie{
		Name:     m.CSRFCookieName,
		Value:    m.csrfValue,
		MaxAge:   0,
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteDefaultMode,
	}

	if m.CSRFCookieOptions != nil {
		m.CSRFCookieOptions(cookie)
	}

	http.SetCookie(c.Writer, cookie)
}

// newDDBManager is used by manager.init, and can be overridden in unit tests.
var newDDBManager = func(cfg aws.Config) *ddb.Manager {
	return ddb.NewManager(dynamodb.NewFromConfig(cfg))
}

func (m *manager[T]) init(ctx context.Context) error {
	return m.once.Do(func() error {
		if m.SessionIdCookieName == "" {
			m.SessionIdCookieName = DefaultSessionIdCookieName
		}

		if m.NewSessionId == nil {
			m.NewSessionId = DefaultNewSessionId
		}

		if m.CSRFCookieName == "" {
			m.CSRFCookieName = DefaultCSRFCookieName
		}

		if m.ddb == nil {
			cfg, err := configcache.Get(ctx)
			if err != nil {
				return fmt.Errorf("sessions: get config error: %w", err)
			}

			m.ddb = newDDBManager(cfg)
		}

		return nil
	})
}

// getManager returns the *manager[T] attached to request, or create a default one if none is available.
//
// Package-level accessors will use this to retrieve or create a *manager[T] for use.
func getManager[T any](c *gin.Context, name string) (*manager[T], error) {
	if v, ok := c.Get(managerKeyPrefix + name); ok {
		return v.(*manager[T]), nil
	}

	m, err := newManager[T](Config{SessionIdCookieName: name})
	if err != nil {
		return nil, err
	}

	c.Set(managerKeyPrefix+name, m)
	return m, nil
}
