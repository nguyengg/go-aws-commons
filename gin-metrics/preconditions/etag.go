package preconditions

import "strings"

// ETag is either strong or weak, and can be either strongly compared or weakly compared against other ETag.
//
// See https://www.rfc-editor.org/rfc/rfc9110.html#entity.tag.comparison.
type ETag interface {
	// Strong returns true if this is a strong ETag.
	Strong() bool
	// Compare compares this vs. the given ETag, and whether the comparison is strong or weak.
	Compare(tag ETag, strong bool) bool
	// String returns the ETag string representation, including W/ prefix if it's a weak ETag.
	String() string

	// value returns the value of the ETag (e.g. "xyzzy" with the quotes and without the W/ prefix).
	value() string
}

// StrongETag is an ETag whose ETag.Strong always returns true.
//
// A few methods require a strong ETag to be passed in, hence this interface.
type StrongETag interface {
	ETag

	// strong is a sentinel impl to differentiate it from weak etag.
	strong()
}

// NewStrongETag returns a strong ETag.
//
// Value should be surrounded by quotes (e.g. "xyzzy").
func NewStrongETag(value string) StrongETag {
	if !strings.HasPrefix(value, `"`) {
		value = `"` + value
	}
	if !strings.HasSuffix(value, `"`) {
		value = value + `"`
	}
	return strongETag{v: value}
}

type strongETag struct {
	v string
}

func (s strongETag) value() string {
	return s.v
}

func (s strongETag) Strong() bool {
	return true
}

func (s strongETag) Compare(tag ETag, strong bool) bool {
	switch t := tag.(type) {
	case weakETag:
		if strong {
			return false
		}

		return s.v == t.v

	case strongETag:
		return s.v == t.v

	default:
		panic("are you strong or weak?")
	}
}

func (s strongETag) strong() {
	// should never be called.
}

func (s strongETag) String() string {
	return s.v
}

// NewWeakETag returns a weak ETag.
//
// Value should be surrounded by quotes (e.g. "xyzzy") without the "W/" prefix.
func NewWeakETag(value string) ETag {
	if !strings.HasPrefix(value, `"`) {
		value = `"` + value
	}
	if !strings.HasSuffix(value, `"`) {
		value = value + `"`
	}
	return weakETag{v: value}
}

type weakETag struct {
	v string
}

func (w weakETag) value() string {
	return w.v
}

func (w weakETag) Strong() bool {
	return false
}

func (w weakETag) Compare(tag ETag, strong bool) bool {
	switch t := tag.(type) {
	case weakETag:
		return w.v == t.v

	case strongETag:
		if strong {
			return false
		}

		return w.v == t.v

	default:
		panic("are you strong or weak?")
	}
}

func (w weakETag) String() string {
	return "W/" + w.v
}
