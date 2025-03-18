package core

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/sugerio/workflow-service-trial/shared/structs"
)

// GetInputData returns the input data.
func GetInputData(data []structs.NodeData) structs.NodeData {
	if len(data) == 0 {
		return structs.NodeData{structs.NodeSingleData{}}
	}
	return data[0]
}

// GetInputDataByIndex get data from multiple input
func GetInputDataByIndex(data []structs.NodeData, index int) structs.NodeData {
	if len(data) <= index {
		return structs.NodeData{}
	}
	return data[index]
}

// GenerateFailedResponse returns a failed response.
func GenerateFailedResponse(nodeName string, err error) *structs.NodeExecutionResult {
	return &structs.NodeExecutionResult{
		ExecutionStatus: structs.WorkflowExecutionStatus_Failed,
		Errors: []structs.WorkflowNodeExecutionError{
			{
				Name:    nodeName,
				Message: err.Error(),
			},
		},
	}
}

// GenerateSuccessResponse returns a success response.
func GenerateSuccessResponse(
	triggerData structs.NodeData, executionData []structs.NodeData) *structs.NodeExecutionResult {
	// Make sure all data has a JSON key wrapper
	triggerData = ReturnJsonArray(triggerData)
	for i, data := range executionData {
		executionData[i] = ReturnJsonArray(data)
	}
	// it is safe to ignore the error here
	return &structs.NodeExecutionResult{
		ExecutionStatus: structs.WorkflowExecutionStatus_Success,
		TriggerData:     triggerData,
		ExecutorData:    executionData,
	}
}

// GenerateEmptyResponse returns an empty response.
func GenerateEmptyResponse() *structs.NodeExecutionResult {
	return &structs.NodeExecutionResult{
		ExecutionStatus: structs.WorkflowExecutionStatus_Success,
		TriggerData:     structs.NodeData{},
		ExecutorData:    []structs.NodeData{},
	}
}

// ContinueOnFail returns true if the workflow should continue on fail.
func ContinueOnFail(params *structs.WorkflowNode) bool {
	if params.OnError == "" {
		return params.ContinueOnFail
	}
	return strings.EqualFold(string(params.OnError), "continueRegularOutput") ||
		strings.EqualFold(string(params.OnError), "continueErrorOutput")
}

// Wrap the data in a JSON key if it is not already wrapped
func ReturnJsonArray(data []map[string]interface{}) structs.NodeData {
	if data == nil {
		return nil
	}
	returnData := structs.NodeData{}
	for _, item := range data {
		if _, ok := item["json"]; ok {
			// We already have the JSON key so avoid double wrapping
			returnData = append(returnData, item)
		} else {
			returnData = append(returnData, structs.NodeSingleData{"json": item})
		}
	}
	return returnData
}

// FilterPropertiesByDisplayOption filter properties by display options
func FilterPropertiesByDisplayOption(
	properties []structs.DescriptionProperties,
	parameters map[string]interface{},
	defaultVersion float64,
) []structs.DescriptionProperties {
	selectedProps := make(map[string]structs.DescriptionProperties)
	propsToConfirm := make([]structs.DescriptionProperties, 0)
	for _, prop := range properties {
		propsToConfirm = append(propsToConfirm, prop)
	}
	// 1. all property's display is unconfirmed from start, put all into unconfirmed list
	// 2. in every iteration, checks all unconfirmed properties,
	//    if it doesn't depend on other property or the properties it depends on are confirmed with a value
	//    we can confirm to display or hide it.
	// 3. then goto next iteration, check if it can display more properties. until on one can display
	for {
		flag := false
		remainProps := make([]structs.DescriptionProperties, 0)
		for index := range propsToConfirm {
			prop := propsToConfirm[index]
			if checkDisplayOptions(prop, selectedProps, parameters, defaultVersion) {
				selectedProps[prop.Name] = prop
				flag = true
			} else {
				remainProps = append(remainProps, prop)
			}
		}
		propsToConfirm = remainProps
		if !flag {
			break
		}
	}

	results := make([]structs.DescriptionProperties, 0)
	for name := range selectedProps {
		results = append(results, selectedProps[name])
	}
	return results
}

// getNodeValueByPath get target sub-field by path from root node
func getNodeValueByPath(root interface{}, subPath []string) interface{} {
	current := root
	for _, p := range subPath {
		if node, ok := current.(map[string]interface{}); ok {
			if _, ok := node[p]; !ok {
				return nil
			}
			current = node[p]
		} else {
			return nil
		}
	}
	return current
}

// checkDisplayOptions check if a property can display by its show rules and hide rule
func checkDisplayOptions(
	property structs.DescriptionProperties,
	propertiesMap map[string]structs.DescriptionProperties,
	parameters map[string]interface{},
	defaultVersion float64,
) bool {
	// get a parameter value
	// if workflow node contains this parameter (saved in workflow creation) use it.
	// otherwise, we should use the default value in node definition (node.json)
	getDefaultOrParameterValue := func(
		properties structs.DescriptionProperties,
		subPath []string,
		inputParams map[string]interface{},
	) interface{} {
		if inputValue, ok := inputParams[properties.Name]; ok {
			return getNodeValueByPath(inputValue, subPath)
		}
		if properties.Default != nil && properties.Default != "" {
			return getNodeValueByPath(properties.Default, subPath)
		}
		return nil
	}
	// check if slice contains a specific value
	containsValue := func(valueList []interface{}, value interface{}) bool {
		for _, v := range valueList {
			if value == v {
				return true
			}
		}
		return false
	}
	// check Show rules in DisplayOptions, return true if it can display,
	// otherwise return false means should hide or currently unknown (due to it depends on
	// other property value. should check it in later iteration)
	checkShowRule := func() bool {
		if property.DisplayOptions.Show != nil {
			for propName, propValues := range property.DisplayOptions.Show {
				// some property depends on NodeVersion.
				// currently all of our nodes created in workflow is defaultVersion
				if propName == "@version" {
					return containsValue(propValues, defaultVersion)
				}
				// in some show rules, property path starts with '/' means it's an absolute path
				// currently all paths are start from root node, same with it
				if propName[:1] == "/" {
					propName = propName[1:]
				}
				// first item is property name, but may contain sub-fields
				propParts := strings.Split(propName, ".")
				dependProperty, ok := propertiesMap[propParts[0]]
				if !ok {
					return false
				}
				value := getDefaultOrParameterValue(dependProperty, propParts[1:], parameters)
				if value == nil {
					return false
				}
				if !containsValue(propValues, value) {
					return false
				}
			}
		}
		return true
	}
	// check Hide rules in DisplayOptions, return true if it can display,
	// otherwise return false means should hide or currently unknown (due to it depends on
	// other property value. should check it in later iteration)
	checkHideRule := func() bool {
		if property.DisplayOptions.Hide != nil {
			for propName, propValues := range property.DisplayOptions.Hide {
				if propName == "@version" {
					return !containsValue(propValues, defaultVersion)
				}
				if propName[:1] == "/" {
					propName = propName[1:]
				}
				propParts := strings.Split(propName, ".")
				dependProperty, ok := propertiesMap[propParts[0]]
				if !ok {
					return false
				}
				value := getDefaultOrParameterValue(dependProperty, propParts[1:], parameters)
				if value == nil {
					return true
				}
				if !containsValue(propValues, value) {
					return true
				}
			}
			return false
		}
		return true
	}
	// must pass all show and hide rules to display
	return checkShowRule() && checkHideRule()
}

func findProp(props []structs.DescriptionProperties, name string) *structs.DescriptionProperties {
	for _, property := range props {
		if property.Name == name {
			return &property
		}
	}
	return nil
}

func findPropFromMaps(props []interface{}, name string) interface{} {
	for _, property := range props {
		if propertyMap, ok := property.(map[string]interface{}); ok {
			if propertyMap["name"] == name {
				return &property
			}
		}
	}
	return nil
}

func findPropertyFromParameterName(
	properties []structs.DescriptionProperties, parameterName string) *structs.DescriptionProperties {
	paramParts := strings.Split(parameterName, ".")
	currentParamPath := ""
	var property interface{}

	for _, p := range paramParts {
		param := strings.Split(p, "[")[0]
		if property == nil {
			property = findProp(properties, param)
			currentParamPath += param
			continue
		}
		if propertySpec, ok := property.(*structs.DescriptionProperties); ok {
			if propertySpec == nil {
				return nil
			}
			options := propertySpec.Options
			if options != nil {
				property = findPropFromMaps(options, param)
				currentParamPath += "." + param
				continue
			}
		}
		if propertyPtr, ok := property.(*interface{}); ok {
			propertyMap, ok := (*propertyPtr).(map[string]interface{})
			if ok {
				options, ok := propertyMap["options"].([]interface{})
				if options != nil && ok {
					property = findPropFromMaps(options, param)
					currentParamPath += "." + param
					continue
				}
				values, ok := propertyMap["values"].([]interface{})
				if values != nil && ok {
					property = findPropFromMaps(values, param)
					currentParamPath += "." + param
					continue
				}
			}
		}
	}

	if currentParamPath == "" {
		return nil
	}
	if ret, ok := property.(*structs.DescriptionProperties); ok {
		return ret
	}
	return nil
}

type GetNodeParameterOptions struct {
	ExtractValue   bool
	RawExpressions bool
	// contextNode
}

// GetNodeParameter retrieves a parameter value identified by nodeName and
// parameterName from the provided input. The value is returned as an interface{}
// and if the parameter is not found or the default value is nil, a fallbackValue is returned
func GetNodeParameter(
	nodeName string,
	parameterName string,
	fallbackValue interface{},
	input *structs.NodeExecuteInput,
	itemIndex int,
	optionsArray ...GetNodeParameterOptions) (interface{}, error) {

	wrapError := func(action string, err error) error {
		return fmt.Errorf("GetNodeParameter [name %s] %s error: %s [item %d]", parameterName, action, err.Error(), itemIndex)
	}
	// -------------- get parameter spec & value ----------------
	nodeSpecs := GetAllNodeObjects()
	nodeSpec := nodeSpecs[nodeName]
	properties := nodeSpec.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.Properties
	// can ignored ok, default value 0 means it has an empty DefaultVersion
	defaultVersion, _ := nodeSpec.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.DefaultVersion.(float64)
	properties = FilterPropertiesByDisplayOption(properties, input.Params.Parameters, defaultVersion)
	propertySpec := findPropertyFromParameterName(properties, parameterName)

	value, ok := GetMapValueByPath(input.Params.Parameters, parameterName)
	if !ok {
		value = nil
	}

	if value == nil && propertySpec != nil {
		path := strings.Split(parameterName, ".")
		value = getNodeValueByPath(propertySpec.Default, path[1:])
	}

	// get default value & merge with value
	if value == nil {
		value = fallbackValue
	}

	options := GetNodeParameterOptions{}
	if len(optionsArray) > 0 {
		options = optionsArray[0]
	}

	if options.RawExpressions {
		return value, nil
	}
	// ---------------- evaluate expression ------------------
	returnData, err := GetParameterValue(value, parameterName, input, itemIndex, false)
	if err != nil {
		return nil, wrapError("evaluate expression", err)
	}
	// -------------------- extract value --------------------
	if options.ExtractValue && propertySpec != nil {
		returnData, err = ExtractValue(returnData, parameterName, input, propertySpec, itemIndex)
		if err != nil {
			return nil, wrapError("extract value", err)
		}
	}
	// validate value
	returnData, err = validateValueAgainstSchema(returnData, propertySpec)
	if err != nil {
		return nil, wrapError("validate value", err)
	}
	return returnData, nil
}

// Only support Basic data types like: int, int64, float64, bool, string
// Composite types like slice, map or struct are not supported, you might end up shooting yourself in the foot.
func GetNodeParameterAsBasicType[T any](
	nodeName string,
	parameterName string,
	fallbackValue T,
	input *structs.NodeExecuteInput,
	itemIndex int,
	optionsArray ...GetNodeParameterOptions) (T, error) {
	val, err := GetNodeParameter(nodeName, parameterName, fallbackValue, input, itemIndex, optionsArray...)
	if err != nil {
		return fallbackValue, err
	}

	valWithType, ok := val.(T)
	if !ok {
		return fallbackValue, fmt.Errorf(
			"GetNodeParameter [name %s] convert type failed: expect: %T, actual: %T [item %d]",
			parameterName, fallbackValue, val, itemIndex)
	}
	return valWithType, nil
}

// GetNodeParameterAsType retrieves the parameter value and converts it into a specified type (T)
// using JSON marshalling/unmarshalling. This function is designed to support composite
// data structures like structs, arrays, slices, and maps, as well as basic data types
// including int, int64, float32, float64, bool, and string. The return value is a pointer
// to the converted parameter of type T. A fallbackValue is returned if the conversion fails
// or the parameter is not found or default value is nil.
func GetNodeParameterAsType[T any](
	nodeName string,
	parameterName string,
	fallbackValue T,
	input *structs.NodeExecuteInput,
	itemIndex int,
	optionsArray ...GetNodeParameterOptions) (*T, error) {
	raw, err := GetNodeParameter(nodeName, parameterName, fallbackValue, input, itemIndex, optionsArray...)
	if err != nil {
		return nil, err
	}
	ret, err := ConvertInterfaceToType[T](raw)
	if err != nil {
		return &fallbackValue, fmt.Errorf(
			"GetNodeParameter [name %s] options convert Json failed: %v [item %d]", parameterName, err, itemIndex)
	}
	return ret, nil
}

func validateValueAgainstSchema(
	value interface{},
	propertySpec *structs.DescriptionProperties,
) (interface{}, error) {
	// TODO
	return value, nil
}

// GetItemBinaryData returns the binary data of the item.
func GetItemBinaryData(
	input *structs.NodeExecuteInput, itemIndex int, propertyName string) (*structs.WorkflowBinaryData, error) {
	inputData := GetInputData(input.Data)
	if itemIndex < 0 || itemIndex >= len(inputData) {
		return nil, fmt.Errorf("invalid item index")
	}
	itemMap := inputData[itemIndex]
	binaryMap, err := ConvertInterfaceToType[map[string]structs.WorkflowBinaryData](itemMap["binary"])
	if err != nil {
		return nil, fmt.Errorf("no binary input found:%w", err)
	}
	// Get the binary data.
	binaryData, ok := (*binaryMap)[propertyName]
	if !ok {
		return nil, fmt.Errorf("missing binary data")
	}
	return &binaryData, nil
}
