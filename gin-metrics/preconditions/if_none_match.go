package preconditions

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// IfNoneMatch parses request header "If-None-Match" and (weakly) compares it against the specified etag.
//
// The method returns two boolean variables: exists is true only if the request header "If-None-Match" is present, and
// noneMatches is true only if the header "If-None-Match" is valid and fails weak comparison against the given etag
// argument.
//
// If the request header "If-None-Match" is present but invalid, a non-nil error is returned. You should abort request
// with a 400 Bad Request in that case. If the header is missing, (false, true, nil) will be returned (see
// https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.2-10.3).
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
		return nil, false, fmt.Errorf("If-None-Match header has empty value")
	case n == 1 && values[0] == "*":
		return anyETagNoneMatcher{}, true, nil
	default:
		var etags []ETag

		for _, t := range values {
			switch matches := eTagPattern.FindStringSubmatch(t); {

			case len(matches) == 3:
				if matches[1] == "W/" {
					etags = append(etags, weakETag{value: matches[2]})
				} else {
					etags = append(etags, strongETag{value: matches[2]})
				}

			default:
				return nil, false, fmt.Errorf("If-None-Match header has invalid value: %q", t)
			}
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

func (m anyETagNoneMatcher) NoneMatch(_ ETag) bool {
	return false
}
