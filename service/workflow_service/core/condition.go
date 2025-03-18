package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TypeValidationMode string
type FilterTypeCombinator string

const (
	StrictMode TypeValidationMode = "strict"
	LooseMode  TypeValidationMode = "loose"

	And FilterTypeCombinator = "and"
	Or  FilterTypeCombinator = "or"
)

type (
	FilterConditionValue struct {
		Id         string            `json:"id"`
		LeftValue  interface{}       `json:"leftValue"`
		RightValue interface{}       `json:"rightValue"`
		Operator   ConditionOperator `json:"operator"`
	}

	ConditionOperator struct {
		Type        string `json:"type"`
		Operation   string `json:"operation"`
		Name        string `json:"name"`
		SingleValue bool   `json:"singleValue,omitempty"`
	}

	FilterOptionsValue struct {
		CaseSensitive  bool
		LeftValue      string
		TypeValidation TypeValidationMode
	}

	FilterValue struct {
		Options    FilterOptionsValue
		Conditions []FilterConditionValue
		Combinator FilterTypeCombinator
	}
)

var (
	TypeError = fmt.Errorf("value type is invalid")
)

// TODO:
// type FilterConditionMetadata = {
// 	index: number;
// 	unresolvedExpressions: boolean;
// 	itemIndex: number;
// 	errorFormat: 'full' | 'inline';
//   };

func ExecuteFilter(value *FilterValue, itemIndex int, continueOnFail bool) (bool, error) {

	options := value.Options
	ignoreCase := !options.CaseSensitive

	conditionRes := make([]bool, len(value.Conditions))
	var err error
	for i := range value.Conditions {
		conditionRes[i], err = ExecuteFilterCondition(ignoreCase, i, itemIndex, value.Conditions[i])
		if err != nil {
			if !continueOnFail {
				return false, err
			}
		}
	}

	if value.Combinator == And {
		return combineConditions("and", conditionRes), nil
	}

	if value.Combinator == Or {
		return combineConditions("or", conditionRes), nil
	}

	return false, fmt.Errorf("invalid filter combinator: %s", value.Combinator)
}

func ExecuteFilterCondition(
	ignoreCase bool,
	index int,
	itemIndex int,
	condition FilterConditionValue) (bool, error) {

	const (
		LeftValue  bool = true
		RightValue bool = false
	)

	newInvalidTypeError := func(isLeftValue bool, invalidTypeDesc string) error {
		var value interface{}
		valueIndex := "1"
		if isLeftValue {
			value = condition.LeftValue
		} else {
			value = condition.RightValue
			valueIndex = "2"
		}
		operationType := condition.Operator.Type
		operation := condition.Operator.Operation

		suffix := fmt.Sprintf("[operation %s %s][actual type %T][item %d]", operationType, operation, value, itemIndex)

		return fmt.Errorf("the provided value %s '%v' in condition %d %s. %s", valueIndex, value, index, invalidTypeDesc, suffix)
	}

	rightValue := condition.RightValue
	leftValue := condition.LeftValue

	exists := leftValue != nil
	if condition.Operator.Operation == "exists" {
		return exists, nil
	} else if condition.Operator.Operation == "notExists" {
		return !exists, nil
	} else if condition.Operator.Operation == "equals" {
		if leftValue == nil && rightValue == nil {
			// return true if both are nil
			return true, nil
		} else if leftValue == nil || rightValue == nil {
			// return false if one of them is nil
			return false, nil
		}

	} else if condition.Operator.Operation == "notEquals" {
		if leftValue == nil && rightValue == nil {
			// return false if both are nil
			return false, nil
		}
		if leftValue == nil || rightValue == nil {
			// return true if one of them is nil
			return true, nil
		}
	}
	switch condition.Operator.Type {
	case "string":
		if leftValue == nil || rightValue == nil {
			return false, nil
		}
		left, ok := leftValue.(string)
		if !ok {
			return false, newInvalidTypeError(LeftValue, "is not string")
		}
		right, ok := rightValue.(string)
		if !ok {
			return false, newInvalidTypeError(RightValue, "is not string")
		}

		if ignoreCase {
			left = strings.ToLower(left)
			if !(condition.Operator.Operation == "regex" || condition.Operator.Operation == "notRegex") {
				right = strings.ToLower(right)
			}
		}

		switch condition.Operator.Operation {
		case "equals":
			return left == right, nil
		case "notEquals":
			return left != right, nil
		case "contains":
			return strings.Contains(left, right), nil
		case "notContains":
			return !strings.Contains(left, right), nil
		case "startsWith":
			return strings.HasPrefix(left, right), nil
		case "notStartsWith":
			return !strings.HasPrefix(left, right), nil
		case "endsWith":
			return strings.HasSuffix(left, right), nil
		case "notEndsWith":
			return !strings.HasSuffix(left, right), nil
		case "regex":
			reg := regexp.MustCompile(right)
			return reg.MatchString(left), nil
		case "notRegex":
			reg := regexp.MustCompile(right)
			return !reg.MatchString(left), nil
		}
	case "number":
		if leftValue == nil || rightValue == nil {
			return false, nil
		}
		left, err := ConvertToFloat(leftValue)
		if err != nil {
			return false, newInvalidTypeError(LeftValue, "is not number")
		}
		right, err := ConvertToFloat(rightValue)
		if err != nil {
			return false, newInvalidTypeError(RightValue, "is not number")
		}

		switch condition.Operator.Operation {
		case "equals":
			return left == right, nil
		case "notEquals":
			return left != right, nil
		case "gt":
			return left > right, nil
		case "lt":
			return left < right, nil
		case "gte":
			return left >= right, nil
		case "lte":
			return left <= right, nil
		}
	case "dateTime":
		if leftValue == nil || rightValue == nil {
			return false, nil
		}
		left, err := ConvertToDate(leftValue)
		if err != nil {
			return false, newInvalidTypeError(LeftValue, "is not date")
		}
		right, err := ConvertToDate(rightValue)
		if err != nil {
			return false, newInvalidTypeError(RightValue, "is not date")
		}

		switch condition.Operator.Operation {
		case "equals":
			return left.Equal(right), nil
		case "notEquals":
			return !left.Equal(right), nil
		case "after":
			return left.After(right), nil
		case "before":
			return left.Before(right), nil
		case "afterEquals":
			return left.Equal(right) || left.After(right), nil
		case "beforeEquals":
			return left.Equal(right) || left.Before(right), nil
		}
	case "boolean":
		left, err := ConvertToBool(leftValue)
		if err != nil {
			return false, newInvalidTypeError(LeftValue, "is not boolean")
		}
		switch condition.Operator.Operation {
		case "true":
			return left, nil
		case "false":
			return !left, nil
		case "equals":
			right, err := ConvertToBool(rightValue)
			if err != nil {
				return false, newInvalidTypeError(RightValue, "is not boolean")
			}
			return left == right, nil
		case "notEquals":
			right, err := ConvertToBool(rightValue)
			if err != nil {
				return false, newInvalidTypeError(RightValue, "is not boolean")
			}
			return left != right, nil
		}
	case "array":
		if leftValue == nil {
			if condition.Operator.Operation == "empty" {
				return true, nil
			}
			return false, nil
		}
		left, err := ConvertToArray(leftValue)
		if err != nil {
			return false, newInvalidTypeError(LeftValue, "is not array")
		}

		switch condition.Operator.Operation {
		case "contains":
			rightJsonBytes, err := json.Marshal(rightValue)
			if err != nil {
				return false, newInvalidTypeError(RightValue, "can not be converted to json")
			}
			for _, v := range left {
				leftJsonBytes, err := json.Marshal(v)
				if err != nil {
					return false, newInvalidTypeError(LeftValue, "can not be converted to json")
				}
				if string(leftJsonBytes) == string(rightJsonBytes) {
					return true, nil
				}
			}
			return false, nil
		case "notContains":
			rightJsonBytes, err := json.Marshal(rightValue)
			if err != nil {
				return false, newInvalidTypeError(RightValue, "can not be converted to json")
			}
			for _, v := range left {
				leftJsonBytes, err := json.Marshal(v)
				if err != nil {
					return false, newInvalidTypeError(LeftValue, "can not be converted to json")
				}
				if string(leftJsonBytes) == string(rightJsonBytes) {
					return false, nil
				}
			}
			return true, nil
		case "empty":
			return len(left) == 0, nil
		case "notEmpty":
			return len(left) != 0, nil
		case "lengthEquals", "lengthNotEquals", "lengthGt", "lengthLt", "lengthGte", "lengthLte":
			rightNumber, err := ConvertToInt(rightValue)
			if err != nil {
				return false, newInvalidTypeError(RightValue, "is not number")
			}
			switch condition.Operator.Operation {
			case "lengthEquals":
				return len(left) == int(rightNumber), nil
			case "lengthNotEquals":
				return len(left) != int(rightNumber), nil
			case "lengthGt":
				return len(left) > int(rightNumber), nil
			case "lengthLt":
				return len(left) < int(rightNumber), nil
			case "lengthGte":
				return len(left) >= int(rightNumber), nil
			case "lengthLte":
				return len(left) <= int(rightNumber), nil
			}
		}
	case "object":
		if leftValue == nil {
			return false, nil // n8n return false if leftValue is nil
		}
		left, err := ConvertToInterfaceMap(leftValue)
		if err != nil {
			return false, newInvalidTypeError(LeftValue, "is not object")
		}

		switch condition.Operator.Operation {
		case "empty":
			return len(left) == 0, nil
		case "notEmpty":
			return len(left) != 0, nil
		}
	}
	return false, nil
}

func combineConditions(combinator string, conditionRes []bool) bool {
	switch combinator {
	case "and":
		for _, res := range conditionRes {
			if !res {
				return false
			}
		}
		return true
	case "or":
		for _, res := range conditionRes {
			if res {
				return true
			}
		}
		return false
	}
	return false
}

func toFilterValues(raw interface{}) (*FilterValue, error) {
	condition := &FilterValue{
		Options:    FilterOptionsValue{},
		Conditions: []FilterConditionValue{},
		Combinator: "",
	}
	if raw == nil {
		return condition, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, condition)
	if err != nil {
		return nil, err
	}
	return condition, nil
}

func ConvertToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	}
	res, err := ConvertToFloat64(value)
	if err == nil {
		return res, nil
	}

	return -1, TypeError
}

func ConvertToInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil

	case string:
		return strconv.ParseInt(v, 10, 64)
	}

	return 0, TypeError
}

// tryParseDate tries to parse a date string against multiple layouts and returns the first success
func tryParseDate(dateStr string) (time.Time, error) {
	var Layouts = []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z0800",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05+08:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range Layouts {
		t, err := time.Parse(layout, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func ConvertToDate(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case string:
		return tryParseDate(v)
	}
	return time.Time{}, TypeError
}

func ConvertToBool(value interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	}
	return false, TypeError
}

func ConvertToArray(value interface{}) ([]interface{}, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case string:
		var arr []interface{}
		err := json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return nil, err
		}
		return arr, nil
	}
	val, err := ConvertToInterfaceArray(value)
	if err == nil {
		return val, nil
	}
	return nil, TypeError
}
