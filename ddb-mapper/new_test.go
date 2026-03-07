package mapper

import (
	"testing"

	"github.com/nguyengg/go-aws-commons/ddb-mapper/types"
	"github.com/stretchr/testify/assert"
)

func TestMustHave(t *testing.T) {
	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	tests := []struct {
		name string
		flag types.AttributeType
	}{
		{
			name: "no sortkey attribute found",
			flag: types.AttributeTypeSortKey,
		},
		{
			name: "no version attribute found",
			flag: types.AttributeTypeVersion,
		},
		{
			name: "no created time attribute found",
			flag: types.AttributeTypeCreatedTime,
		},
		{
			name: "no modified time attribute found",
			flag: types.AttributeTypeModifiedTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MustHave[Item](tt.flag)(New[Item]())
			assert.ErrorContains(t, err, tt.name)
		})
	}
}
