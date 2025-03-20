package core

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

// JsonStr returns the JSON string of the input.
func JsonStr(input interface{}) string {
	// FIXME: error should NOT be ignored
	resultByte, _ := json.Marshal(input)
	return string(resultByte)
}

func ConvertNullTimeToStandardTimePointer(nullTime sql.NullTime) *time.Time {
	if nullTime.Valid {
		return &nullTime.Time
	} else {
		return nil
	}
}

func UnmarshalOmitEmpty(from []byte, to interface{}) error {
	if from == nil || len(from) == 0 {
		return nil
	}
	return json.Unmarshal(from, to)
}

// GetValueFromMap Get value from map[string]interface{} by key.
func GetValueFromMap[T any](data map[string]interface{}, key string) (T, bool) {
	// val, ok := data[key] will not panic if data is nil
	// val := data[key] will panic if data is nil
	if val, ok := data[key]; ok {
		if v, ok := val.(T); ok {
			return v, true
		}
	}
	// zero value of T
	var zero T
	return zero, false
}

// GetValueFromMapWithDefault Get value from map[string]interface{} by key with default value.
func GetValueFromMapWithDefault[T any](data map[string]interface{}, key string, defaultValue T) T {
	if val, ok := data[key]; ok {
		if v, ok := val.(T); ok {
			return v
		}
	}
	return defaultValue
}

// is array or slice
func IsArray(target interface{}) bool {
	if target == nil {
		return false
	}
	kind := reflect.TypeOf(target).Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

// Erase type
func ConvertToInterfaceArray(target interface{}) ([]interface{}, error) {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("provided interface is not a slice/array - %v", val.Kind())
	}

	res := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		res[i] = val.Index(i).Interface()
	}
	return res, nil
}

// is map
func IsMap(target interface{}) bool {
	if target == nil {
		return false
	}
	kind := reflect.TypeOf(target).Kind()
	return kind == reflect.Map
}

func ConvertToInterfaceMap(target interface{}) (map[string]interface{}, error) {
	if target == nil {
		return nil, nil
	}
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("provided interface is not a map - %v", val.Kind())
	}

	retMap := make(map[string]interface{})
	for _, key := range val.MapKeys() {
		retMap[key.String()] = val.MapIndex(key).Interface()
	}
	return retMap, nil
}

func ConvertToFloat64(target interface{}) (float64, error) {
	if target == nil {
		return 0, fmt.Errorf("provided interface is nil")
	}
	val := reflect.ValueOf(target)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil

	default:
		return 0, fmt.Errorf("provided interface is not a number - %v", val.Kind())
	}
}

func isStructZero(target interface{}) bool {
	if target == nil {
		return true
	}

	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < val.NumField(); i++ {
		if !reflect.DeepEqual(val.Field(i).Interface(), reflect.Zero(val.Field(i).Type()).Interface()) {
			return false
		}
	}
	return true
}

// path can be a.b.c
func GetMapValueByPath(m map[string]interface{}, path string) (interface{}, bool) {
	segments := strings.Split(path, ".")

	var ok bool
	var val interface{}
	for _, segment := range segments {
		val, ok = m[segment]
		m, _ = val.(map[string]interface{})
	}

	return val, ok
}

// Unflatten the string which is stringfied using github.com/WebReflection/flatted
func UnflattenString(target string) (string, error) {
	// Unmarshal to json array
	var arrayObject []interface{}
	err := json.Unmarshal([]byte(target), &arrayObject)
	if err != nil {
		return "", err
	}
	// Parse string
	result := unflattenStringItem(0, &arrayObject)
	// Marshal to json string
	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func unflattenStringItem(index int, source *[]interface{}) interface{} {
	item := (*source)[index]
	// str, pure value
	if str, ok := item.(string); ok {
		return str
	}
	//array, all items is index
	if IsArray(item) {
		result := []interface{}{}
		for _, val := range item.([]interface{}) {
			isNum, numVal := IsNumericStr(val)
			if isNum {
				result = append(result, unflattenStringItem(numVal, source))
			} else {
				result = append(result, val)
			}
		}
		return result
	}

	//map, field value may contains index
	if itemMap, ok := item.(map[string]interface{}); ok {
		result := map[string]interface{}{}
		for k, v := range itemMap {
			isNum, numVal := IsNumericStr(v)
			if isNum {
				result[k] = unflattenStringItem(numVal, source)
			} else {
				result[k] = v
			}
		}
		return result
	}

	return "error"
}

// Check if a target is a numeric string
func IsNumericStr(target interface{}) (bool, int) {
	if str, ok := target.(string); ok {
		val, err := strconv.Atoi(str)
		if err == nil {
			return true, val
		}
	}
	return false, 0
}

// try to convert to a specific type using JSON marshaling.
func ConvertInterfaceToType[T any](input interface{}) (*T, error) {
	result := new(T)

	// If the input is nil, return the zero value of the type.
	if input == nil {
		return result, nil
	}

	// Marshal the input into JSON, checking for errors.
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Unmarshal the JSON data into the result, checking for errors.
	if err := json.Unmarshal(jsonData, result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to type: %w", err)
	}

	return result, nil
}

// wrap error
// additionalInfos is optional
func NewNodeSingleDataError(err error, itemIndex int, additionalInfos ...map[string]interface{}) structs.NodeSingleData {
	if len(additionalInfos) > 0 {
		return structs.NodeSingleData{
			"json": map[string]interface{}{
				"error":                err.Error(),
				"itemIndex":            itemIndex,
				"success":              false,
				"additionalReturnData": additionalInfos[0],
			},
		}
	}
	return structs.NodeSingleData{
		"json": map[string]interface{}{
			"error":     err.Error(),
			"itemIndex": itemIndex,
			"success":   false,
		},
	}
}

// Render content from html template and fields
func GetHtmlContentFromTemplate(htmlTemplate string, fields map[string]interface{}) (string, error) {
	tmpl, err := template.New("").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	if err := tmpl.Execute(&result, fields); err != nil {
		return "", err
	}
	return result.String(), nil
}

// Get the HTML template fields from the node input.
func GetHtmlTemplateFieldsFromNodeInput(input *structs.NodeExecuteInput, itemIndex int) (map[string]interface{}, error) {
	var data structs.NodeData
	// input.Data might be empty
	if len(input.Data) > 0 {
		if itemIndex < 0 || itemIndex >= len(input.Data) {
			return nil, fmt.Errorf("item index out of range")
		}
		data = input.Data[itemIndex]
	}

	fields := map[string]interface{}{
		"params": input.Params.Parameters,
		"data":   data,
		"index":  itemIndex,
	}
	return fields, nil
}

// Whether the content is fixed content, Expressions are prefixed with "="
func IsFixedContent(content string) bool {
	return !strings.HasPrefix(content, "=")
}

// Get data from structs.WorkflowBinaryData
// If the data is base64 encoded, decode it
func GetDataFromBinaryData(binaryData *structs.WorkflowBinaryData) (string, error) {
	// Empty binary data
	if binaryData.Data == "" {
		return "", fmt.Errorf("binary data is empty")
	}

	// Not base64 encoded, return directly
	if !binaryData.Base64Encoded {
		return binaryData.Data, nil
	}

	// Base64 encoded, decode it
	decoded, err := base64.StdEncoding.DecodeString(binaryData.Data)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func CheckDataType(typ string, val interface{}) bool {
	// TODO:
	switch typ {
	case "email":
	case "number":
		//....
	}

	return true
}
func CheckFormTriggerParam(params map[string]interface{}, node *structs.WorkflowNode) error {
	parameters, err := json.Marshal(params)
	if err != nil {
		return errors.WithStack(err)
	}

	var workflowFrom structs.WorkflowFrom
	err = json.Unmarshal(parameters, &workflowFrom)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, field := range workflowFrom.FormFields.Values {

		val, ok := params[field.FieldLabel]
		if field.RequiredField && !ok {
			return errors.Errorf("%s is empty", field.FieldLabel)
		}

		if !ok {
			continue
		}

		if !CheckDataType(field.FieldLabel, val) {
			return errors.Errorf("%s invalid value", field.FieldLabel)
		}
	}

	return nil
}
