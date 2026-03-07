package mapper

import (
	"fmt"
	"reflect"

	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/untyped"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/types"
)

// New creates a new [Mapper] modeling the DynamoDB table that contains items of type T.
//
// `dynamodbav` struct tags tell [New] what AttributeType to assign the field:
//
//	// Field tagged with "hashkey" or "pk" become AttributeTypeHashKey.
//	// Additionally, the hashkey struct tag must also provide the table name.
//	ID string `dynamodbav:"id,hashkey" tablename:"Items"`
//
//	// Field tagged with "sortkey", "sk", or "rangekey" become AttributeTypeRangeKey.
//	Shard int `dynamodbav:"id,sk"`
//
//	// Field tagged with "version" become AttributeTypeVersion.
//	// string, int, and uint have out-of-the-box support.
//	Version int `dynamodbav:"version,version"`
//
//	// Field tagged with "createdTime" or "modifiedTime" become AttributeTypeCreatedTime or AttributeTypeModifiedTime.
//	Created time.Time `dynamodbav:"created,createdTime,unixtime"`
//	Modified time.Time `dynamodbav:"modified,createdTime,unixtime"`
//
//	// All other unignored fields become AttributeTypeOther.
//
// Duplicate attribute names are not allowed, and special-type attributes cannot appear more than once (can't have two
// version attributes for example). A common usage pattern is to create a global Mapper variable in the same package
// that defines the struct that models the item:
//
//	package app
//
//	import "github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
//
//	type Item struct {
//		ID string `json:"id" dynamodbav:"id,hashkey" tablename:"Items"`
//	}
//
//	var Mapper *mapper.Mapper[Item]
//
//	func init() {
//		var err error
//		Mapper, err = mapper.New[Item]()
//		if err != nil {
//			panic(err)
//		}
//	}
//
// Then the [Mapper] can be used like this:
//
//	item := app.Item{ID: "id"}
//	app.Mapper.Get(context.Background(), &item)
func New[T any](optFns ...func(m *Mapper[T])) (m *Mapper[T], err error) {
	tType := reflect.TypeFor[T]()
	if tType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("New[T] requires T to be a struct kind of type, not %s", tType)
	}

	m = internal.ApplyOpts(&Mapper[T]{}, optFns...)
	if m.Mapper, err = untyped.NewFromType(tType, func(opts *untyped.Options) {
		opts.Client = m.Client
		opts.Encoder = m.Encoder
		opts.Decoder = m.Decoder
		opts.VersionUpdater = m.wrapVersionUpdater()
	}); err != nil {
		return nil, err
	}

	return m, nil
}

// MustHave is a helper method to validate that the struct correctly models the required attribute types.
//
// Usage:
//
//	m, err := MustHave[T](AttributeTypeVersion | AttributeTypeCreatedTime | AttributeTypeModifiedTime)(New[T]())
//	if err != nil {
//		panic(err)
//	}
//
// AttributeTypeHashKey and AttributeTypeOther are ignored.
func MustHave[T any](flag types.AttributeType) func(m *Mapper[T], err error) (*Mapper[T], error) {
	return func(m *Mapper[T], err error) (*Mapper[T], error) {
		if flag&types.AttributeTypeSortKey != 0 && m.Mapper.SortKey == nil {
			return nil, fmt.Errorf("no sortkey attribute found")
		}
		if flag&types.AttributeTypeVersion != 0 && m.Version == nil {
			return nil, fmt.Errorf("no version attribute found")
		}
		if flag&types.AttributeTypeCreatedTime != 0 && m.CreatedTime == nil {
			return nil, fmt.Errorf("no created time attribute found")
		}
		if flag&types.AttributeTypeModifiedTime != 0 && m.ModifiedTime == nil {
			return nil, fmt.Errorf("no modified time attribute found")
		}
		return m, err
	}
}
