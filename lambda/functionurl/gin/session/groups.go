package session

// Groups is a string list, preferably a string set.
type Groups []string

// Test verifies that the user's groups satisfy the membership rules.
//
// Use AllOf and/or OneOf to describe how to authorise the user's groups.
//
// Usage:
//
//	// user must be able to read both payments and inventory.
//	Groups([]string{...}).Test(AllOf("can_read_payment", "can_read_inventory"))
//
//	// user must be able to read both payments and inventory, but write permissions implies read as well.
//	Groups([]string{...}).Test(OneOf("can_read_payment", "can_write_payment"), OneOf("can_read_inventory", "can_write_inventory"))
func (groups Groups) Test(rule func(*rules), more ...Rule) bool {
	opts := &rules{
		allOf: make(map[string]bool),
		oneOf: nil,
	}
	rule(opts)
	for _, f := range more {
		f(opts)
	}

	// the one-pass algorithm goes through the groups that user belongs to then checks off the list.
	// if any rules remain then user is not authorised.
	for _, group := range groups {
		delete(opts.allOf, group)

		var p, n *node
		for n = opts.oneOf; n != nil; p, n = n, n.next {
			if n.groups[group] {
				if p == nil {
					opts.oneOf = n.next
					continue
				}

				p.next = n.next
			}
		}
	}

	return len(opts.allOf) == 0 && opts.oneOf == nil
}

// mini linked list implementation of oneOf rules.
type node struct {
	groups map[string]bool
	next   *node
}

type rules struct {
	allOf map[string]bool
	oneOf *node
}

type Rule func(*rules)

// AllOf adds a rule that the user must belong to all the groups specified here.
func AllOf(group string, more ...string) Rule {
	return func(opts *rules) {
		opts.allOf[group] = true
		for _, g := range more {
			opts.allOf[g] = true
		}
	}
}

// OneOf adds a rule that the user must belong to at least one of the groups specified here.
func OneOf(first, second string, more ...string) Rule {
	return func(opts *rules) {
		groups := map[string]bool{first: true, second: true}
		for _, group := range more {
			groups[group] = true
		}

		opts.oneOf = &node{
			groups: groups,
			next:   opts.oneOf,
		}
	}
}
