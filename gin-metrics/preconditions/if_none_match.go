package preconditions

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// IfNoneMatch parses request header "If-None-Match" and uses weak comparison to compare the request header against the
// specified etag.
//
// For the return value:
//   - exists is true only if the request header is present
//   - noneMatches is true only if exists is true, the request header is valid and passes evaluation described in
//     https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.1-8
//   - a non-nil error implies exists is true, noneMatches is false, and the request header has invalid value.
//
// Usage:
//
//	switch exists, noneMatches, err := IfNoneMatch(c, preconditions.NewWeakEtag(`W/"xyzzy"`)); {
//		case noneMatches && err == nil:
//			// (*, true, nil): your conditional GET can proceed.
//		case !noneMatches && err == nil:
//			// (*, false, nil): your conditional GET should abort with 304 Not Modified; other methods with 412
//			// Precondition Failed.
//		case err != nil:
//			// (true, *, nil): you should abort with 400 Bad Request since If-None-Match is invalid. err.Error() can be
//	 		// used as error message.
//		case !exists:
//			// (false, true, nil) will be returned, see https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.2-10.3.
//	}
//
// In the scenario where "the origin server does not have a current representation for the target resource"
// (https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.2-10.1), use IfNoneMatchNoETag instead which will return
// (false, true, nil) if If-None-Match is "*".
func IfNoneMatch(c *gin.Context, etag ETag) (exists, noneMatches bool, err error) {
	var m ifNoneMatcher
	if v, ok := c.Get(ifNoneMatchKey); ok {
		if err, ok = v.(error); ok {
			return true, false, err
		}

		m = v.(ifNoneMatcher)
	} else {
		if m, exists, err = parseIfNoneMatch(c.Request.Header); !exists {
			return false, true, nil
		} else if err != nil {
			c.Set(ifNoneMatchKey, err)
			return true, false, err
		}

		c.Set(ifNoneMatchKey, m)
	}

	return true, m.NoneMatch(etag), nil
}

// IfNoneMatchNoETag is a variant of IfNoneMatch for the scenario where "the origin server does not have a current
// representation for the target resource" (https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.2-10.1).
//
// IfNoneMatchNoETag will return (false, true, nil) if If-None-Match is "*".
func IfNoneMatchNoETag(c *gin.Context) (exists, noneMatches bool, err error) {
	return IfNoneMatch(c, noETag{})
}

// parseIfNoneMatch can be used as a low-level construct to parse and check the validity of "If-None-Match" request
// header.
//
// A non-nil error is returned only if the "If-None-Match" header is present but invalid. A nil ifNoneMatcher is
// returned if the request did not have the "If-None-Match" header.
func parseIfNoneMatch(header http.Header) (m ifNoneMatcher, exists bool, err error) {
	values, ok := header["If-None-Match"]
	if !ok {
		return nil, false, nil
	}

	switch n := len(values); {
	case n == 0:
		return nil, false, fmt.Errorf("If-None-Match header has empty values")
	case n == 1 && values[0] == "*":
		return anyETagNoneMatcher{}, true, nil
	default:
		var etags []ETag

		for i, t := range values {
			etag, err := ParseETag(t)
			if err != nil {
				return nil, false, fmt.Errorf("If-None-Match header has invalid value at ordinal %d", i+1)
			}

			etags = append(etags, etag)
		}

		return eTagsNoneMatcher(etags), true, nil
	}
}

// ifNoneMatcher has single method NoneMatch to compare weakly "If-None-Match" header against ETag.
type ifNoneMatcher interface {
	// NoneMatch compares weakly against the given ETag.
	NoneMatch(ETag) bool
}

type eTagsNoneMatcher []ETag

func (m eTagsNoneMatcher) NoneMatch(tag ETag) bool {
	for _, t := range m {
		if t.Compare(tag, false) {
			log.Printf("%s vs. %s == %t", t, tag, true)
			return false
		}
	}

	return true
}

type anyETagNoneMatcher struct {
}

func (m anyETagNoneMatcher) NoneMatch(tag ETag) bool {
	if _, ok := tag.(noETag); ok {
		return true
	}

	return false
}
