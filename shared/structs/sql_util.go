package structs

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sugerio/workflow-service-trial/shared"
)

// ToSqlQuery Convert the SqlCondition to SQL query string.
func (ob *SqlCondition) ToSqlQuery() (string, error) {
	// 1. Start with column name. Remove all spaces in the column name.
	queryString := strings.ReplaceAll(ob.Column, " ", "") + " "

	if DetectSqlInjection(ob.Column) {
		return "", errors.New("SQL injection detected in column name")
	}

	if DetectSqlInjection(fmt.Sprintf("%v", ob.Value)) {
		return "", errors.New("SQL injection detected in value")
	}

	// 2. Add the operator.
	switch ob.Operator {
	case SqlOperator_EQ:
		queryString += "="
	case SqlOperator_NOT_EQ:
		queryString += "!="
	case SqlOperator_GT:
		queryString += ">"
	case SqlOperator_GTE:
		queryString += ">="
	case SqlOperator_LT:
		queryString += "<"
	case SqlOperator_LTE:
		queryString += "<="
	case SqlOperator_IS:
		queryString += "IS"
	case SqlOperator_IS_NOT:
		queryString += "IS NOT"
	case SqlOperator_IN:
		queryString += "IN"
	case SqlOperator_NOT_IN:
		queryString += "NOT IN"
	case SqlOperator_LIKE:
		queryString += "LIKE"
	case SqlOperator_ILIKE:
		queryString += "ILIKE"
	case SqlOperator_NOT_LIKE:
		queryString += "NOT LIKE"
	default:
		return "", fmt.Errorf("invalid SQL operator: %s", ob.Operator)
	}
	// Add a space after the operator.
	queryString += " "

	// 3. Add the value.
	switch ob.ValueType {
	case SqlValueType_STRING:
		v, ok := ob.Value.(string)
		if !ok {
			return "", fmt.Errorf("invalid value: %v; expecting string", ob.Value)
		}
		queryString += "'" + v + "'"
	case SqlValueType_INT:
		v, ok := ob.Value.(int)
		if !ok {
			return "", fmt.Errorf("invalid value: %v; expecting int", ob.Value)
		}
		queryString += fmt.Sprintf("%d", v)
	case SqlValueType_FLOAT:
		v, ok := ob.Value.(float64)
		if !ok {
			return "", fmt.Errorf("invalid value: %v; expecting float64", ob.Value)
		}
		queryString += fmt.Sprintf("%f", v)
	case SqlValueType_BOOL:
		v, ok := ob.Value.(bool)
		if !ok {
			return "", fmt.Errorf("invalid value: %v; expecting bool", ob.Value)
		}
		queryString += fmt.Sprintf("%t", v)
	case SqlValueType_STRING_ARRAY:
		stringArray, err := shared.ConvertInterfaceToArray[string](ob.Value)
		if err != nil {
			return "", err
		}
		queryString += "('" + strings.Join(stringArray, "','") + "')"
	case SqlValueType_INT_ARRAY:
		intArray, err := shared.ConvertInterfaceToArray[int](ob.Value)
		if err != nil {
			return "", err
		}
		queryString +=
			"(" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(intArray)), ","), "[]") + ")"
	case SqlValueType_FLOAT_ARRAY:
		floatArray, err := shared.ConvertInterfaceToArray[float64](ob.Value)
		if err != nil {
			return "", err
		}
		queryString +=
			"(" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(floatArray)), ","), "[]") + ")"
	case SqlValueType_BOOL_ARRAY:
		boolArray, err := shared.ConvertInterfaceToArray[bool](ob.Value)
		if err != nil {
			return "", err
		}
		queryString +=
			"(" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(boolArray)), ","), "[]") + ")"
	case SqlValueType_NULL:
		queryString += "NULL"
	default:
		return "", fmt.Errorf("invalid SQL value type: %s", ob.ValueType)
	}

	return queryString, nil
}

// DetectSqlInjection Detect SQL injection in the segment.
// Returns true if SQL injection is detected.
// Otherwise, returns false.
func DetectSqlInjection(segment string) bool {

	if segment == "" {
		return false
	}

	seg := strings.ToUpper(segment)

	invalidComments := []string{"--", "/*", "*/", "#"}
	for _, comment := range invalidComments {
		if strings.Contains(seg, comment) {
			return true
		}
	}

	invalidChars := []string{";", "\""}
	for _, char := range invalidChars {
		if strings.Contains(seg, char) {
			return true
		}
	}

	invalidKeywords := []string{"CASE WHEN", "LOAD_FILE", "UNION", "HAVING", "CURRENT_CATALOG",
		"CURRENT_DATABASE", "CURRENT_QUERY", "CURRENT_SCHEMA", "CURRENT_USER",
		"DROP TABLE", "CREATE TABLE", "ALTER TABLE", "DELETE FROM", "INSERT INTO"}
	for _, keyword := range invalidKeywords {
		if strings.Contains(seg, keyword) {
			return true
		}
	}

	// using regex to detect functions, black list functions that can be used for SQL injection
	// functions:
	//  CHR,  LENGTH, CHAR_LENGTH, CHAR, HEX, BIN, OCT, QUOTE, QUOTENAME, QUOTED_IDENTIFIER
	//  ASCII, UNICODE, NCHAR and all functions starting with PG_

	functionBlackList := []string{"CHR", "LENGTH", "CHAR_LENGTH", "CHAR", "HEX", "BIN", "OCT", "QUOTE",
		"QUOTENAME", "QUOTED_IDENTIFIER", "ASCII", "UNICODE", "NCHAR", "PG_.*"}

	// build regex string
	var regexStr string
	for _, keyword := range functionBlackList {
		regexStr += fmt.Sprintf(`\b%s\s*\(|`, keyword)
	}
	// remove the last "|"
	regexStr = regexStr[:len(regexStr)-1]
	if regexp.MustCompile(regexStr).MatchString(seg) {
		return true
	}

	// using regex to detect operators, black list operators that can be used for SQL injection
	// operators:
	// ' AND,  ' OR ,' NOT
	regexStr = `(\s*'\s*AND\s*|\s*'\s*OR\s*|\s*'\s*NOT\s*)`
	if regexp.MustCompile(regexStr).MatchString(seg) {
		return true
	}

	return false
}
