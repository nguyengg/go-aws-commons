package sessions

import "github.com/nguyengg/go-aws-commons/opaque-token/hmac"

// WithCSRF configures New to use CSRF generation with the given signer and verifier.
//
// The same hmac.Hasher will be used for CSRF validation as well. See [github.com/nguyengg/go-aws-commons/opaque-token/hmac]
// for various options on constructing the hmac.Hasher.
//
// [github.com/nguyengg/go-aws-commons/opaque-token/hmac]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/opaque-token/hmac
func WithCSRF(signVerifier hmac.Hasher) func(cfg *Config) {
	return func(cfg *Config) {
		cfg.csrfSignVerifier = signVerifier
	}
}
