package stringset

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStringSet_IsSubset(t *testing.T) {
	type args struct {
		other StringSet
	}
	tests := []struct {
		name string
		m    StringSet
		args args
		want bool
	}{
		{
			name: "is subset",
			m:    []string{"a", "b"},
			args: args{other: []string{"a", "b", "c"}},
			want: true,
		},
		{
			name: "is not subset",
			m:    []string{"a", "b", "c"},
			args: args{other: []string{"a", "b", "d"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsSubset(tt.args.other); got != tt.want {
				t.Errorf("IsSubset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSet_IsSuperset(t *testing.T) {
	type args struct {
		other StringSet
	}
	tests := []struct {
		name string
		m    StringSet
		args args
		want bool
	}{
		{
			name: "is superset",
			m:    []string{"a", "b", "c"},
			args: args{other: []string{"a", "b"}},
			want: true,
		},
		{
			name: "is not superset",
			m:    []string{"a", "b", "c"},
			args: args{other: []string{"a", "b", "d"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsSuperset(tt.args.other); got != tt.want {
				t.Errorf("IsSuperset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSet_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       StringSet
		want    []byte
		wantErr bool
	}{
		{
			name: "marshal",
			m:    []string{"a", "b", "c"},
			want: []byte(`["a","b","c"]`),
		},
		{
			name: "marshal empty array",
			m:    []string{},
			want: []byte(`[]`),
		},
		{
			name: "marshal empty stringset",
			m:    make(StringSet, 0),
			want: []byte(`[]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestStringSet_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       StringSet
		args    args
		wantErr bool
	}{
		{
			name: "unmarshal",
			m:    []string{"a", "b", "c"},
			args: args{data: []byte(`["a","b","c"]`)},
		},

		{
			name: "unmarshal empty array",
			m:    []string{},
			args: args{data: []byte(`[]`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := json.Unmarshal(tt.args.data, &tt.m); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringSet_Delete(t *testing.T) {
	type args struct {
		value     string
		lenBefore int
		lenAfter  int
	}
	tests := []struct {
		name   string
		m      StringSet
		args   args
		wantOk bool
	}{
		// TODO: Add test cases.
		{
			name: "delete OK",
			m:    []string{"a", "b", "c"},
			args: args{
				value:     "a",
				lenBefore: 3,
				lenAfter:  2,
			},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if lenBefore := len(tt.m); lenBefore != tt.args.lenBefore {
				t.Errorf("lenBefore = %v, want %v", lenBefore, tt.args.lenBefore)
			}

			if gotOk := tt.m.Delete(tt.args.value); gotOk != tt.wantOk {
				t.Errorf("Delete() = %v, want %v", gotOk, tt.wantOk)
			}

			if tt.wantOk && tt.m.Has(tt.args.value) {
				t.Errorf("Has() = %t, want %t", true, false)
			}

			if lenAfter := len(tt.m); lenAfter != tt.args.lenAfter {
				t.Errorf("lenAfter = %v, want %v", lenAfter, tt.args.lenAfter)
			}
		})
	}
}

func TestStringSet_Add(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		m    StringSet
		args args
		want bool
	}{
		{
			name: "add OK",
			m:    []string{"a", "b"},
			args: args{value: "c"},
			want: true,
		},
		{
			name: "add duplicate",
			m:    []string{"a", "b"},
			args: args{value: "b"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Add(tt.args.value); got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		values []string
	}
	tests := []struct {
		name string
		args args
		want StringSet
	}{
		// TODO: Add test cases.
		{
			name: "empty array",
			args: args{values: []string{}},
			want: StringSet{},
		},
		{
			name: "array",
			args: args{values: []string{"a", "b"}},
			want: StringSet{"a", "b"},
		},
		{
			name: "duplicates removed",
			args: args{values: []string{"a", "b", "a"}},
			want: StringSet{"a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSet_Equals(t *testing.T) {
	type args struct {
		other StringSet
	}
	tests := []struct {
		name string
		m    StringSet
		args args
		want bool
	}{
		{
			name: "equals",
			m:    []string{"a", "b", "c"},
			args: args{other: []string{"c", "b", "a"}},
			want: true,
		},
		{
			name: "empty sets are equal",
			m:    []string{},
			args: args{other: StringSet{}},
			want: true,
		},
		{
			name: "not equal",
			m:    []string{"a", "b", "c"},
			args: args{other: []string{"a", "b"}},
			want: false,
		},
		{
			name: "not equal",
			m:    []string{"a", "b"},
			args: args{other: []string{"a", "b", "c"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Equal(tt.args.other); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
