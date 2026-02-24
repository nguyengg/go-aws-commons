package preconditions

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// IfMatch parses request header "If-Match" and (strongly) compares it against the specified strong etag.
//
// The method returns two boolean variables: exists is true only if the request header "If-Match" is present, and
// matches is true only if the header "If-Match" is valid and passes strong comparison against the given etag argument.
//
// If the request header "If-Match" is present but invalid, a non-nil error is returned. You should abort request with a
// 400 Bad Request in that case. If the header is missing, (false, false, nil) will be returned as if it was not a matches
// (see https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.1-9.3).
func IfMatch(c *gin.Context, etag StrongETag) (exists, matches bool, err error) {
	var m ifMatcher
	if v, ok := c.Get(ifMatchKey); ok {
		if err, ok = v.(error); ok {
			return true, false, err
		}

		m = v.(ifMatcher)
	} else {
		if m, exists, err = parseIfMatch(c.Request.Header); !exists {
			return false, false, nil
		} else if err != nil {
			c.Set(ifMatchKey, err)
			return false, false, err
		}

		c.Set(ifMatchKey, m)
	}

	return true, m.Match(etag), nil
}

var eTagPattern = regexp.MustCompile(`^(W/)?(".+")$`)

// parseIfMatch can be used as a low-level method to parse and check the validity of "If-Match" request header.
//
// A non-nil error is returned only if the "If-Match" header is present but invalid. A nil ifMatcher is returned if the
// request did not have the "If-Match" header.
func parseIfMatch(header http.Header) (m ifMatcher, exists bool, err error) {
	values, ok := header["If-Match"]
	if !ok {
		return nil, false, nil
	}

	switch n := len(values); {
	case n == 0:
		return nil, false, fmt.Errorf("If-Match header has empty value")
	case n == 1 && values[0] == "*":
		return anyETagMatcher{}, true, nil
	default:
		var etags []StrongETag

		for _, t := range values {
			switch matches := eTagPattern.FindStringSubmatch(t); {

			case len(matches) == 3:
				if matches[1] == "W/" {
					return nil, false, fmt.Errorf("If-Match header must only contain strong ETags or is *")
				}

				etags = append(etags, strongETag{value: matches[2]})

			default:
				return nil, false, fmt.Errorf("If-Match header has invalid value: %q", t)
			}
		}

		return strongETagsMatcher(etags), true, nil
	}
}

// ifMatcher has single method Match to compare strongly "If-Match" header against strong ETag.
type ifMatcher interface {
	// Match compares strongly against the given strong ETag.
	Match(StrongETag) bool
}

type strongETagsMatcher []StrongETag

func (m strongETagsMatcher) Match(tag StrongETag) bool {
	for _, t := range m {
		if t.Compare(tag, true) {
			return true
		}
	}

	return false
}

type anyETagMatcher struct {
}

func (m anyETagMatcher) Match(tag StrongETag) bool {
	return true
}
