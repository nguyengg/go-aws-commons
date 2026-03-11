package sessions

import (
	"context"
	"crypto/rand"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/csrf"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/gbac"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
	ini "github.com/nguyengg/init-once"
)

const (
	// DefaultSessionIdCookieName is the default cookie name storing session Id.
	//
	// From the request, it is used to retrieve the DynamoDB item storing full session metadata. It is also set on
	// the response as part of session generation so that user can attach it with subsequent requests.
	DefaultSessionIdCookieName = "sid"

	// DefaultCSRFCookieName is the default cookie name storing CSRF token.
	//
	// From the request, it is optionally used as a CSRF source for validation. It is also set on the response as
	// part of CSRF generation workflow so that user can attach the CSRF token with subsequent requests.
	DefaultCSRFCookieName = csrf.DefaultCookieName
)

// Config customises [New].
type Config struct {
	// SessionIdCookieName is the name of the cookie that contains session Id.
	//
	// Default to DefaultSessionIdCookieName.
	SessionIdCookieName string
	// SessionCookieOptions can be used to modify the session cookie prior to setting the Set-Cookie response header.
	//
	// Invalid settings will cause the cookie to be silent dropped so be very careful with this. Most likely you just
	// want to change the [http.Cookie.MaxAge] to something more reasonable.
	SessionCookieOptions func(c *http.Cookie)
	// NewSessionId is used to create the Id for a new session.
	//
	// Defaults to DefaultNewSessionId. You can replace it with uuid.NewString for example.
	NewSessionId func() string

	// CSRFCookieName is the name of the cookie that contains CSRF token.
	//
	// Defaults to DefaultCSRFCookieName.
	CSRFCookieName string
	// CSRFCookieOptions can be used to modify the session cookie prior to setting the Set-Cookie response header.
	//
	// Invalid settings will cause the cookie to be silent dropped so be very careful with this. Most likely you just
	// want to change the [http.Cookie.MaxAge] to something more reasonable.
	CSRFCookieOptions func(c *http.Cookie)
	// CSRFKeyProvider returns the secret that is used to create and validate CSRF tokens.
	//
	// Defaults to a 32-byte key that is randomly generated at first use which is not suitable for production. To
	// disable CSRF generation, pass DisableCSRF.
	CSRFKeyProvider keys.Provider

	// Client is the client for making DynamoDB service calls.
	Client *dynamodb.Client

	// these opaque fields must use the various With* methods to configure.

	mapperOpts    []func(cfg *config.Config)
	csrfDisabled  bool
	extractGroups func(c *gin.Context) (authenticated bool, groups gbac.Groups)
	groupsOptions gbac.Options
}

// WithMapperOptions customises the internal [mapper.Mapper] that [Manager] uses.
//
// Subsequent WithMapperOptions will replace settings made by previous invocations.
func WithMapperOptions(optFns ...func(cfg *config.Config)) func(cfg *Config) {
	return func(cfg *Config) {
		cfg.mapperOpts = optFns
	}
}

var _ mapper.Mapper[any]

type defaultKeyProvider struct {
	key  []byte
	once ini.SuccessOnce
}

func (d *defaultKeyProvider) Provide(_ context.Context, _ *string) ([]byte, *string, error) {
	err := d.once.Do(func() error {
		d.key = make([]byte, 32)
		_, err := rand.Read(d.key)
		return err
	})
	return d.key, nil, err
}
