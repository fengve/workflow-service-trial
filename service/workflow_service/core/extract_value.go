package core

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/sugerio/workflow-service-trial/shared/structs"
)

func ExtractValue(
	value interface{},
	parameterName string,
	ctx *structs.NodeExecuteInput,
	propertySpec *structs.DescriptionProperties,
	itemIndex int,
) (interface{}, error) {

	// findPropertyFromParameterName

	if len(propertySpec.Type) == 0 {
		return value, nil
	}

	switch propertySpec.Type {

	case "resourceLocator":
		return extractValueRLC(value, propertySpec, parameterName)

	case "filter":
		return extractValueFilter(value, propertySpec, parameterName, itemIndex)

	default:
		return extractValueOther(value, propertySpec, parameterName)
	}

}

func extractValueRLC(
	value interface{},
	propertySpec *structs.DescriptionProperties,
	parameterName string,
) (interface{}, error) {

	// get modeProp & check
	var modeProp structs.DescriptionModes
	valMap, ok := value.(map[string]interface{})
	if !ok {
		return value, nil
	}

	if mode, ok := valMap["mode"].(string); ok {
		for _, modeProp = range propertySpec.Modes {
			if modeProp.Name == mode {
				break // found
			}
		}
	}

	if isStructZero(modeProp.ExtractValue) {
		return valMap["value"], nil
	}

	valStr, ok := valMap["value"].(string)
	if !ok {
		typeName := fmt.Sprintf("%T", valStr)
		return nil, fmt.Errorf(`only strings can be passed to extractValue. Parameter "%s" passed "%s"`, parameterName, typeName)
	}

	if modeProp.ExtractValue.Type != "regex" {
		return nil, fmt.Errorf("property with unknown `extractValue`: parameter=%s, extractValueType=%s", parameterName, modeProp.ExtractValue.Type)
	}

	regex, err := regexp.Compile(modeProp.ExtractValue.Regex)
	if err != nil {
		return nil, err
	}

	return executeRegexExtractValue(valStr, regex, parameterName, propertySpec.DisplayName)

}

func isFilterValue(value interface{}) bool {

	return value != nil && reflect.TypeOf(value).Kind() == reflect.Map &&
		reflect.ValueOf(value).MapIndex(reflect.ValueOf("conditions")).IsValid() &&
		reflect.ValueOf(value).MapIndex(reflect.ValueOf("combinator")).IsValid()
}

func extractValueFilter(
	value interface{},
	propertySpec *structs.DescriptionProperties,
	parameterName string,
	itemIndex int,
) (interface{}, error) {
	if !isFilterValue(value) {
		return value, nil
	}

	filterValues, err := toFilterValues(value)

	if err != nil {
		return nil, fmt.Errorf(`property "%s" has an invalid filter value. %v`, parameterName, err)
	}

	return ExecuteFilter(filterValues, itemIndex, false)
}

func extractValueOther(
	value interface{},
	propertySpec *structs.DescriptionProperties,
	parameterName string,
) (interface{}, error) {

	// check
	if isStructZero(propertySpec.ExtractValue) {
		return value, nil
	}

	valStr, ok := value.(string)
	if !ok {
		typeName := fmt.Sprintf("%T", valStr)
		return nil, fmt.Errorf(`only strings can be passed to extractValue. Parameter "%s" passed "%s"`, parameterName, typeName)
	}

	if propertySpec.ExtractValue.Type != "regex" {
		return nil, fmt.Errorf("property with unknown `extractValue`: parameter=%s, extractValueType=%s", parameterName, propertySpec.ExtractValue.Type)
	}

	regex, err := regexp.Compile(propertySpec.ExtractValue.Regex)
	if err != nil {
		return nil, err
	}

	return executeRegexExtractValue(valStr, regex, parameterName, propertySpec.DisplayName)
}

func isParameterDisplay(parameterName string, ctx *structs.NodeExecuteInput) bool {
	// TODO
	return true
}

func executeRegexExtractValue(
	value string,
	regex *regexp.Regexp,
	parameterName string,
	displayName string,
) (interface{}, error) {
	match := regex.FindString(value)
	if match == "" {
		return nil, fmt.Errorf("no match found for regex in parameter %s (%s)", parameterName, displayName)
	}
	// TODO check more than one match
	return match, nil
}
