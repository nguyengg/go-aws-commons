package types

// AttributeType is a [Mapper] concept to differentiate the roles of attributes.
//
// AttributeType is unrelated to an attribute's [DynamoDB data type] ("S", "N", "B", etc.) nor its Go type (string, int
// etc.). [Mapper] specifically cares about these types of attributes:
//   - Key attributes, which are either hash/partition key and sort/range keys. [Mapper] does not care about the
//     attribute's Go type, so long as they marshal to "S", N", or "B" data types as required by DynamoDB.
//   - Version attribute for optimistic locking. Go string, int and uint types have out-of-the box support; any other Go
//     types will need their own VersionUpdater implementation.
//   - Created and modified time attributes for auto-generating timestamps. The Go types of those attributes must be
//     assignable to [time.Time].
//
// [DynamoDB data type]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.NamingRulesDataTypes.html#HowItWorks.DataTypes
type AttributeType int

const (
	// AttributeTypeHashKey is a hash key attribute.
	AttributeTypeHashKey AttributeType = 1 << iota
	// AttributeTypeSortKey is a sort key attribute.
	AttributeTypeSortKey
	// AttributeTypeVersion is the version attribute used in optimistic locking.
	AttributeTypeVersion
	// AttributeTypeCreatedTime is an attribute for created time.
	AttributeTypeCreatedTime
	// AttributeTypeModifiedTime is an attribute for modification time.
	AttributeTypeModifiedTime
	// AttributeTypeOther is an attribute that has no special semantics to Mapper.
	AttributeTypeOther
)

func (a AttributeType) String() string {
	switch a {
	case AttributeTypeHashKey:
		return "HashKey"
	case AttributeTypeSortKey:
		return "SortKey"
	case AttributeTypeVersion:
		return "Version"
	case AttributeTypeCreatedTime:
		return "CreatedTime"
	case AttributeTypeModifiedTime:
		return "ModifiedTime"
	case AttributeTypeOther:
		return ""
	default:
		return ""
	}
}

// Attribute models a DynamoDB attribute parsed from a `dynamodbav`-tagged struct field.
type Attribute interface {
	// AttributeName is the first tag value in the `dynamodbav` struct tag.
	//
	// This is also the name of the attribute in DynamoDB when encoded with attributevalue.Encoder. AttributeName is
	// often different from reflect.StructField.Name. For example, given this struct:
	//
	//	type MyStruct struct {
	//		ID string `dynamodbav:"id"`
	//	}
	//
	// The [reflect.StructField.Name] would be "ID" while AttributeName would be "id".
	AttributeName() string

	// AttributeType returns the type of this attribute that has special semantics to Mapper.
	AttributeType() AttributeType

	// Get retrieves the value of the field in the given item.
	//
	// Because Get is read-only, item may be either a struct or struct pointer.
	Get(item any) (any, error)

	// Set updates the value of the field in the given item.
	//
	// Unlike Get which is read-only, Set will modify the item argument; item must be a struct pointer as a result.
	Set(item, value any) error
}

// UpdateBehaviour is used with [mapper.UpdateOptions.WithUpdateBehaviour] to populate the update expression using the
// map[string]AttributeValue marshaled from the item.
type UpdateBehaviour int

const (
	// UpdateBehaviourSkipZeroValues makes it so that attributes that are nil or have zero values in the
	// item struct are skipped from the update expression.
	//
	// Useful if you have a custom struct that doesn't work with "omitempty" tags.
	UpdateBehaviourSkipZeroValues UpdateBehaviour = 1 << iota

	// UpdateBehaviourSkipNULLDataType will omit all attributes that have NULL data type.
	//
	// This takes precedence over UpdateBehaviourSkipNULLDataType.
	UpdateBehaviourSkipNULLDataType

	// UpdateBehaviourRemoveNULLDataType will add a REMOVE clause for all attributes that have NULL data type.
	//
	// Incompatible with UpdateBehaviourSkipNULLDataType so will not take effect if UpdateBehaviourSkipNULLDataType
	// is specified.
	UpdateBehaviourRemoveNULLDataType
)

const (
	// UpdateBehaviourDefault is the safest behaviour to use.
	UpdateBehaviourDefault = UpdateBehaviourSkipZeroValues
	// UpdateBehaviourAsTagged will write everything, including empty/zero values and NULL data type.
	UpdateBehaviourAsTagged = 0
)
