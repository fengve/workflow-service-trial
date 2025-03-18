package structs

type SqlCondition struct {
	// The column name.
	Column string `json:"column"`
	// The operator.
	Operator SqlOperator `json:"operator"`
	// The value.
	Value interface{} `json:"value"`
	// The value type.
	ValueType SqlValueType `json:"valueType"`
} //@name SqlCondition

type SqlOperator string //@name SqlOperator

const (
	SqlOperator_EQ       SqlOperator = "EQ"       // Equal
	SqlOperator_NOT_EQ   SqlOperator = "NOT_EQ"   // Not equal
	SqlOperator_GT       SqlOperator = "GT"       // Greater than
	SqlOperator_GTE      SqlOperator = "GTE"      // Greater than or equal
	SqlOperator_LT       SqlOperator = "LT"       // Less than
	SqlOperator_LTE      SqlOperator = "LTE"      // Less than or equal
	SqlOperator_IS       SqlOperator = "IS"       // Is
	SqlOperator_IS_NOT   SqlOperator = "IS_NOT"   // Is not
	SqlOperator_IN       SqlOperator = "IN"       // In
	SqlOperator_NOT_IN   SqlOperator = "NOT_IN"   // Not in
	SqlOperator_LIKE     SqlOperator = "LIKE"     // Like
	SqlOperator_ILIKE    SqlOperator = "ILIKE"    // Ilike
	SqlOperator_NOT_LIKE SqlOperator = "NOT_LIKE" // Not like
)

type SqlValueType string //@name SqlValueType

const (
	SqlValueType_STRING SqlValueType = "STRING" // String
	SqlValueType_INT    SqlValueType = "INT"    // Int
	SqlValueType_FLOAT  SqlValueType = "FLOAT"  // Float
	SqlValueType_BOOL   SqlValueType = "BOOL"   // Boolean

	SqlValueType_STRING_ARRAY SqlValueType = "STRING_ARRAY" // Array of strings
	SqlValueType_INT_ARRAY    SqlValueType = "INT_ARRAY"    // Array of integers
	SqlValueType_FLOAT_ARRAY  SqlValueType = "FLOAT_ARRAY"  // Array of floats
	SqlValueType_BOOL_ARRAY   SqlValueType = "BOOL_ARRAY"   // Array of booleans

	SqlValueType_NULL SqlValueType = "NULL" // Null
)
