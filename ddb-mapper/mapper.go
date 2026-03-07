// Package mapper provides [Mapper] to interact with DynamoDB tables in a type-safe way.
//
// See [New] for information regarding what type of attributes have special semantics to [Mapper].
package mapper

import (
	"iter"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/untyped"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/types"
)

// Mapper provides a type-safe way to interact with the DynamoDB table containing items of type T.
//
// See [New] for information regarding what type of attributes have special semantics to [Mapper].
type Mapper[T any] struct {
	*untyped.Mapper

	// Client is the client for making DynamoDB service calls.
	//
	// If nil, a default client will be created using configcache.Get.
	Client *dynamodb.Client
	// Encoder is the attributevalue.Encoder to marshal structs into DynamoDB items.
	Encoder *attributevalue.Encoder
	// Decoder is the attributevalue.Decoder to unmarshal attributes from DynamoDB.
	Decoder *attributevalue.Decoder

	// VersionUpdater is used to generate the next version value by updating the item's version value in-place.
	//
	// If VersionUpdater is given, it is always used to update the version. Otherwise, the version's type determines
	// how its next value is computed:
	//	- For integers, uint and int types work out of the box; the next value is version + 1.
	//	- Floats are not supported; VersionUpdater must be explicitly provided.
	//	- For string and string aliases, uuid.NewString produces the next version.
	//	- For any other types, VersionUpdater must be explicitly provided.
	VersionUpdater func(item *T)
}

// TableName returns the name of the table modeled by this [Mapper].
func (m *Mapper[T]) TableName() string {
	return m.Mapper.TableName
}

// HashKey returns the required hash key attribute.
func (m *Mapper[T]) HashKey() types.Attribute {
	return m.Mapper.HashKey
}

// SortKey returns the optional sort key attribute.
func (m *Mapper[T]) SortKey() types.Attribute {
	return m.Mapper.SortKey
}

// All returns an iterator over all attributes modeled by this [Mapper].
func (m *Mapper[T]) All() iter.Seq[types.Attribute] {
	return func(yield func(types.Attribute) bool) {
		if !yield(m.Mapper.HashKey) {
			return
		}
		if m.Mapper.SortKey != nil && !yield(m.Mapper.SortKey) {
			return
		}
		if m.Mapper.Version != nil && !yield(m.Mapper.Version) {
			return
		}
		if m.CreatedTime != nil && !yield(m.CreatedTime) {
			return
		}
		if m.ModifiedTime != nil && !yield(m.ModifiedTime) {
			return
		}

		for _, attr := range m.Others {
			if !yield(attr) {
				return
			}
		}
	}
}

// wrapVersionUpdater returns the version updater function that [untyped.Mapper] uses.
func (m *Mapper[T]) wrapVersionUpdater() func(item any) {
	if fn := m.VersionUpdater; fn != nil {
		return func(item any) {
			m.VersionUpdater(item.(*T))
		}
	}

	return nil
}
