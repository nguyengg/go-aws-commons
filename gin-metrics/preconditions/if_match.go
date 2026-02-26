package preconditions

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// IfMatch parses "If-Match" and uses strong comparison to compare the request header against the specified etag.
//
// For the return values:
//   - exists is true only if the request header is present
//   - matches is true only if exists is true, the request header is valid and passes evaluation described in
//     https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.1-8
//   - a non-nil error implies exists is true, matches is false, and the request header has invalid value.
//
// Usage:
//
//	switch exists, matches, err := IfMatch(c, preconditions.NewStrongEtag(`"xyzzy"`)); {
//		case matches && err == nil:
//			// (*, true, nil): your POST, PUT, or DELETE can proceed.
//		case !matches && err == nil:
//			// (*, false, nil): your POST, PUT, or DELETE should abort with 412 Precondition Failed.
//		case err != nil:
//			// (true, *, nil): your POST, PUT, or DELETE should abort with 400 Bad Request since If-Match is invalid.
//			// err.Error() can be used as error message.
//		case !exists:
//			// (false, false, nil) will be returned as if it was not a match (see https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.1-9.3).
//	}
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
		return nil, false, fmt.Errorf("If-Match header has empty values")
	case n == 1 && values[0] == "*":
		return anyETagMatcher{}, true, nil
	default:
		var etags []StrongETag

		for i, t := range values {
			etag, err := ParseETag(t)
			if err != nil {
				return nil, false, fmt.Errorf("If-Match header has invalid value at ordinal %d", i+1)
			}

			if v, ok := etag.(StrongETag); ok {
				etags = append(etags, v)
			} else {
				return nil, false, fmt.Errorf("If-Match header must only contain strong ETags or is *")
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
