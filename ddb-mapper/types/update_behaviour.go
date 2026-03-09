package types

// UpdateBehaviour is used with [ddb.UpdateOptions.WithUpdateBehaviour] to populate the update expression using the
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
