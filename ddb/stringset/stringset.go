package stringset

// StringSet adds convenient methods to manipulate the items in the set.
//
// It is imperative that tag `dynamodbav :",stringset"` is used to allow the field to be marshaled as a string set. If
// you forget to do so, the array will be marshalled as a list instead.
// See https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute#Marshal.
type StringSet []string

// New creates a new StringSet with duplicate values removed.
func New(values []string) StringSet {
	m := make(map[string]bool)
	s := make([]string, 0)
	for _, v := range values {
		if _, ok := m[v]; !ok {
			m[v] = true
			s = append(s, v)
		}
	}

	return s
}

// Add return true only if the value hasn't existed in the set before the invocation.
func (m *StringSet) Add(value string) bool {
	for _, v := range *m {
		if v == value {
			return false
		}
	}
	*m = append(*m, value)
	return true
}

func (m *StringSet) Has(value string) (ok bool) {
	for _, v := range *m {
		if v == value {
			return true
		}
	}

	return false
}

// Delete returns true only if the value existed in the set before the invocation.
func (m *StringSet) Delete(value string) (ok bool) {
	n := len(*m) - 1
	for i, v := range *m {
		if v == value {
			if i < n {
				(*m)[i], (*m)[n] = (*m)[n], (*m)[i]
			}
			*m = (*m)[:n]
			return true
		}
	}

	return false
}

// Clear resets the array to an empty one.
func (m *StringSet) Clear() {
	*m = make(StringSet, 0)
}

// IsSubset returns true if every element in this set is in the specified set (other).
func (m *StringSet) IsSubset(other StringSet) bool {
	s := make(map[string]bool, len(*m))
	for _, v := range other {
		s[v] = true
	}
	for _, v := range *m {
		if _, ok := s[v]; !ok {
			return false
		}
	}

	return true
}

// IsSuperset returns true if every element in the specified set (other) is in this set.
func (m *StringSet) IsSuperset(other StringSet) bool {
	s := make(map[string]bool, len(*m))
	for _, v := range *m {
		s[v] = true
	}
	for _, v := range other {
		if _, ok := s[v]; !ok {
			return false
		}
	}

	return true
}

// Equal returns true only if every element in this set is in the other set and vice versa.
func (m *StringSet) Equal(other StringSet) bool {
	s := make(map[string]bool, len(*m))
	for _, v := range *m {
		s[v] = true
	}
	for _, v := range other {
		if _, ok := s[v]; !ok {
			return false
		}
		delete(s, v)
	}

	return len(s) == 0
}
