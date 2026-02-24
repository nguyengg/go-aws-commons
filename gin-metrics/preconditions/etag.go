package preconditions

import "strings"

// ETag is either strong or weak, and can be either strongly compared or weakly compared against other ETag.
//
// See https://www.rfc-editor.org/rfc/rfc9110.html#entity.tag.comparison.
type ETag interface {
	// Strong returns true if this is a strong ETag.
	Strong() bool
	// Value returns the value of the ETag (e.g. "xyzzy" with the quotes; without the W/ prefix for weak ETag).
	Value() string
	// Compare compares this vs. the given ETag, and whether the comparison is strong or weak.
	Compare(tag ETag, strong bool) bool
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
	return strongETag{value: value}
}

type strongETag struct {
	value string
}

func (s strongETag) Strong() bool {
	return true
}

func (s strongETag) Value() string {
	return s.value
}

func (s strongETag) Compare(tag ETag, strong bool) bool {
	switch t := tag.(type) {
	case weakETag:
		if strong {
			return false
		}

		return s.value == t.value

	case strongETag:
		return s.value == t.value

	default:
		panic("are you strong or weak?")
	}
}

func (s strongETag) strong() {
	// should never be called.
}

func (s strongETag) String() string {
	return s.value
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
	return weakETag{value: value}
}

type weakETag struct {
	value string
}

func (w weakETag) Strong() bool {
	return false
}

func (w weakETag) Value() string {
	return w.value
}

func (w weakETag) Compare(tag ETag, strong bool) bool {
	switch t := tag.(type) {
	case weakETag:
		return w.value == t.value

	case strongETag:
		if strong {
			return false
		}

		return w.value == t.value

	default:
		panic("are you strong or weak?")
	}
}

func (w weakETag) String() string {
	return "W/" + w.value
}
