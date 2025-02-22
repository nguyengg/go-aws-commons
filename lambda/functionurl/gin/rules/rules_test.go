package rules

import (
	"testing"
)

func TestGroups(t *testing.T) {
	type args struct {
		rule func(*rules)
		more []Rule
	}
	tests := []struct {
		name   string
		groups Groups
		args   args
		want   bool
	}{
		{
			name:   "AllOf(a, b, c) = true",
			groups: []string{"a", "b", "c"},
			args: args{
				rule: AllOf("a", "b", "c"),
			},
			want: true,
		},
		{
			name:   "AllOf(a, b, c) = false",
			groups: []string{"a", "b"},
			args: args{
				rule: AllOf("a", "b", "c"),
			},
			want: false,
		},
		{
			name:   "OneOf(a, b) = true",
			groups: []string{"a"},
			args: args{
				rule: OneOf("a", "b"),
			},
			want: true,
		},
		{
			name:   "OneOf(a, b) = false",
			groups: []string{"c"},
			args: args{
				rule: OneOf("a", "b"),
			},
			want: false,
		},
		{
			name:   "OneOf(a, b) & OneOf(c, d) = true",
			groups: []string{"a", "c"},
			args: args{
				rule: OneOf("a", "b"),
				more: []Rule{OneOf("c", "d")},
			},
			want: true,
		},
		{
			name:   "OneOf(a, b) & OneOf(c, d) = false - only a",
			groups: []string{"a"},
			args: args{
				rule: OneOf("a", "b"),
				more: []Rule{OneOf("c", "d")},
			},
			want: false,
		},
		{
			name:   "OneOf(a, b) & OneOf(c, d) = false - only c",
			groups: []string{"c"},
			args: args{
				rule: OneOf("a", "b"),
				more: []Rule{OneOf("c", "d")},
			},
			want: false,
		},
		{
			name:   "OneOf(a, b) & OneOf(c, d) = false - no presence",
			groups: []string{"e"},
			args: args{
				rule: OneOf("a", "b"),
				more: []Rule{OneOf("c", "d")},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.groups.Test(tt.args.rule, tt.args.more...); got != tt.want {
				t.Errorf("Test() = %v, want %v", got, tt.want)
			}
		})
	}
}
