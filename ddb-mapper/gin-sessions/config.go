package sessions

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/csrf"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	"github.com/nguyengg/go-aws-commons/opaque-token/hmac"
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

	// Client is the client for making DynamoDB service calls.
	Client *dynamodb.Client

	// these opaque fields must use the various With* methods to configure.

	mapperOpts []func(cfg *config.Config)
	csrf       hmac.Engine
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
