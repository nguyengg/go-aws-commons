package mapper

import (
	"testing"

	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMustHave(t *testing.T) {
	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	tests := []struct {
		name string
		flag model.AttributeModelType
	}{
		{
			name: "no sortkey attribute found",
			flag: model.AttributeModelTypeSortKey,
		},
		{
			name: "no version attribute found",
			flag: model.AttributeModelTypeVersion,
		},
		{
			name: "no createdtime attribute found",
			flag: model.AttributeModelTypeCreatedTime,
		},
		{
			name: "no modifiedtime attribute found",
			flag: model.AttributeModelTypeModifiedTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMustHave[Item](tt.flag)
			assert.ErrorContains(t, err, tt.name)
		})
	}
}
