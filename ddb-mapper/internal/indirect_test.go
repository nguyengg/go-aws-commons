package internal

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IndirectValueIsStruct(t *testing.T) {
	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	tests := []struct {
		name          string
		t             reflect.Type
		mustBePointer bool
		wantErr       bool
	}{
		{
			name:    "Item",
			t:       reflect.TypeFor[Item](),
			wantErr: false,
		},
		{
			name:    "*Item",
			t:       reflect.TypeFor[**Item](),
			wantErr: false,
		},
		{
			name:    "**Item",
			t:       reflect.TypeFor[**Item](),
			wantErr: false,
		},
		{
			name:    "***Item",
			t:       reflect.TypeFor[***Item](),
			wantErr: false,
		},
		{
			// this is the only time error is returned in this test since Item type is not pointer And
			// mustBePointer is true.
			name:          "Item is not pointer",
			t:             reflect.TypeFor[Item](),
			mustBePointer: true,
			wantErr:       true,
		},
		{
			name:          "*Item is pointer",
			t:             reflect.TypeFor[**Item](),
			mustBePointer: true,
			wantErr:       false,
		},
		{
			name:          "**Item is pointer",
			t:             reflect.TypeFor[**Item](),
			mustBePointer: true,
			wantErr:       false,
		},
		{
			name:          "***Item is pointer",
			t:             reflect.TypeFor[***Item](),
			mustBePointer: true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := IndirectTypeIsStruct(tt.t, tt.mustBePointer)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, reflect.TypeFor[Item](), v)
		})
	}
}

func Test_IndirectTypeIsStruct(t *testing.T) {
	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	var ptr = func(i any) *any {
		return &i
	}

	tests := []struct {
		name          string
		item          any
		mustBePointer bool
		wantErr       bool
	}{
		{
			name:    "Item",
			item:    Item{},
			wantErr: false,
		},
		{
			name:    "*Item",
			item:    &Item{},
			wantErr: false,
		},
		{
			name:    "**Item",
			item:    ptr(&Item{}),
			wantErr: false,
		},
		{
			name:    "***Item",
			item:    ptr(ptr(&Item{})),
			wantErr: false,
		},
		{
			// this is the only time error is returned in this test since Item type is not pointer And
			// mustBePointer is true.
			name:          "Item is not pointer",
			item:          Item{},
			mustBePointer: true,
			wantErr:       true,
		},
		{
			name:          "*Item is pointer",
			item:          &Item{},
			mustBePointer: true,
			wantErr:       false,
		},
		{
			name:          "**Item is pointer",
			item:          ptr(&Item{}),
			mustBePointer: true,
			wantErr:       false,
		},
		{
			name:          "***Item is pointer",
			item:          ptr(ptr(&Item{})),
			mustBePointer: true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ptr, err := IndirectValueIsStruct(tt.item, tt.mustBePointer)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, reflect.TypeFor[Item](), v.Type())
			if ptr.IsValid() {
				assert.Equal(t, reflect.TypeFor[*Item](), ptr.Type())
			}
		})
	}
}
