package aggregatenode

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	Category = structs.CategoryExecutor
	Name     = "n8n-nodes-base.aggregate"
)

var (
	//go:embed node.json
	rawJson []byte

	//go:embed aggregate.svg
	icon []byte
)

type (
	AggregateExecutor struct {
		spec *structs.WorkflowNodeSpec
	}

	ParameterOptions struct {
		DisableDotNotation bool `json:"disableDotNotation,omitempty"`
		IncludeBinaries    bool `json:"includeBinaries,omitempty"`
		KeepOnlyUnique     bool `json:"keepOnlyUnique,omitempty"`
		MergeLists         bool `json:"mergeLists,omitempty"`
		KeepMissing        bool `json:"keepMissing,omitempty"`
	}

	FieldsToAggregate struct {
		FieldToAggregate []FieldToAggregate `json:"fieldToAggregate,omitempty"`
	}

	FieldToAggregate struct {
		FieldToAggregate string `json:"fieldToAggregate,omitempty"`
		RenameField      bool   `json:"renameField,omitempty"`
		OutputFieldName  string `json:"outputFieldName,omitempty"`
	}
)

func init() {
	ae := &AggregateExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	ae.spec.JsonConfig = rawJson
	ae.spec.GenerateSpec()

	core.Register(ae)
	core.RegisterEmbedIcons(Name, icon)
}

func (ae *AggregateExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (ae *AggregateExecutor) Name() string {
	return Name
}

func (ae *AggregateExecutor) DefaultSpec() interface{} {
	return ae.spec
}

func (ae *AggregateExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	items := core.GetInputData(input.Data)
	resultJson := map[string]interface{}{}

	optionsRaw, err := core.GetNodeParameter(Name, "options",
		map[string]interface{}{},
		input,
		0,
		core.GetNodeParameterOptions{},
	)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	options, err := toOptions(optionsRaw)
	if err != nil {
		return core.GenerateFailedResponse(Name, fmt.Errorf(`property "options" is invalid: %v`, err))
	}

	aggregate, err := core.GetNodeParameterAsBasicType(Name, "aggregate", "aggregateIndividualFields", input, 0)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	if aggregate == "aggregateIndividualFields" {
		disableDotNotation := options.DisableDotNotation
		mergeLists := options.MergeLists
		keepMissing := options.KeepMissing
		fieldsToAggregateRaw, err := core.GetNodeParameter(Name, "fieldsToAggregate", map[string]interface{}{},
			input, 0)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		fieldsToAggregate, err := toFieldsToAggregate(fieldsToAggregateRaw)
		if err != nil {
			return core.GenerateFailedResponse(Name, fmt.Errorf(`property "fieldsToAggregate" is invalid: %v`, err))
		}

		fieldToAggregate := fieldsToAggregate.FieldToAggregate
		if len(fieldToAggregate) == 0 {
			return core.GenerateFailedResponse(Name, fmt.Errorf("please add a field to aggregate"))
		}

		values := map[string]interface{}{}
		handleFields := []string{}
		for _, targetField := range fieldToAggregate {
			fieldToAggregate := targetField.FieldToAggregate

			// Check duplicate field
			handleField := fieldToAggregate
			if targetField.RenameField {
				handleField = targetField.OutputFieldName
			}
			if checkArrayContains(handleFields, handleField) {
				return core.GenerateFailedResponse(Name, fmt.Errorf("please make sure each output field name is unique"))
			} else {
				handleFields = append(handleFields, handleField)
			}

			outputField := targetField.OutputFieldName
			if outputField == "" {
				outputField = fieldToAggregate
				if !disableDotNotation && strings.Contains(fieldToAggregate, ".") {
					parts := strings.Split(fieldToAggregate, ".")
					lastPart := parts[len(parts)-1]
					outputField = lastPart
				}
			}

			if fieldToAggregate == "" {
				continue
			}

			outputVal := []interface{}{}
			for _, item := range items {
				value := item["json"].(map[string]interface{})[fieldToAggregate]
				if !disableDotNotation {
					value, _ = getValFromJsonByPath(item["json"].(map[string]interface{}), fieldToAggregate)
				}

				if !keepMissing {
					if value == nil {
						continue
					} else if isArray(value) {
						arrayOfValue := convertUnknowArray(value)
						value = removeNilFromArray(arrayOfValue)
					}
				}

				if isArray(value) && mergeLists {
					arrayOfValue := convertUnknowArray(value)
					outputVal = append(outputVal, arrayOfValue...)
				} else {
					outputVal = append(outputVal, value)
				}
			}
			values[outputField] = outputVal
		}

		for k, v := range values {
			if !disableDotNotation {
				setValToJsonByPath(resultJson, k, v)
			} else {
				resultJson[k] = v
			}
		}
	} else {
		jsonItems := []map[string]interface{}{}
		for _, item := range items {
			jsonItems = append(jsonItems, item["json"].(map[string]interface{}))
		}
		destinationFieldName, err := core.GetNodeParameterAsBasicType(Name, "destinationFieldName", "data",
			input, 0)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		fieldsToIncludeRaw, err := core.GetNodeParameterAsBasicType(Name, "fieldsToInclude", "",
			input, 0)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		fieldsToExcludeRaw, err := core.GetNodeParameterAsBasicType(Name, "fieldsToExclude", "",
			input, 0)

		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		fieldsToInclude := splitStringToArray(fieldsToIncludeRaw, ",")
		fieldsToExclude := splitStringToArray(fieldsToExcludeRaw, ",")

		if len(fieldsToInclude) > 0 || len(fieldsToExclude) > 0 {
			newJsonItems := []map[string]interface{}{}
			for _, jsonItem := range jsonItems {
				outputFields := []string{}
				if len(fieldsToInclude) > 0 {
					for key := range jsonItem {
						if checkArrayContains(fieldsToInclude, key) {
							outputFields = append(outputFields, key)
						}
					}
				}
				if len(fieldsToExclude) > 0 {
					for key := range jsonItem {
						if !checkArrayContains(fieldsToExclude, key) {
							outputFields = append(outputFields, key)
						}
					}
				}

				newJsonItem := map[string]interface{}{}
				for _, field := range outputFields {
					newJsonItem[field] = jsonItem[field]
				}
				newJsonItems = append(newJsonItems, newJsonItem)
			}
			jsonItems = newJsonItems
		}
		resultJson[destinationFieldName] = jsonItems
	}

	result := structs.NodeData{
		map[string]interface{}{
			"json": resultJson,
		},
	}

	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
}

func toOptions(raw interface{}) (*ParameterOptions, error) {
	options := &ParameterOptions{}
	if raw == nil {
		return options, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, options)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func toFieldsToAggregate(raw interface{}) (*FieldsToAggregate, error) {
	result := &FieldsToAggregate{}
	if raw == nil {
		return result, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func checkArrayContains(arr []string, target string) bool {
	for _, item := range arr {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}

func isArray(obj interface{}) bool {
	if obj == nil {
		return false
	}
	value := reflect.TypeOf(obj)
	return value.Kind() == reflect.Array || value.Kind() == reflect.Slice
}

func convertUnknowArray(obj interface{}) []interface{} {
	interfaceSlice := make([]interface{}, 0)
	valValue := reflect.ValueOf(obj)
	for i := 0; i < valValue.Len(); i++ {
		element := valValue.Index(i).Interface()
		interfaceSlice = append(interfaceSlice, element)
	}
	return interfaceSlice
}

func removeNilFromArray(arr []interface{}) []interface{} {
	arrFiltered := []interface{}{}
	for _, valueChild := range arr {
		if valueChild != nil {
			arrFiltered = append(arrFiltered, valueChild)
		}
	}
	return arrFiltered
}

func getValFromJsonByPath(obj map[string]interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")
	for _, key := range keys {
		value, ok := obj[key]
		if !ok {
			return nil, fmt.Errorf("key not found: %s", key)
		}

		if nestedObj, ok := value.(map[string]interface{}); ok {
			obj = nestedObj
		} else {
			return value, nil
		}
	}

	return nil, fmt.Errorf("key not found: %s", keys[len(keys)-1])
}

func setValToJsonByPath(obj map[string]interface{}, path string, value interface{}) error {
	keys := strings.Split(path, ".")
	for i, key := range keys {
		if i == len(keys)-1 {
			obj[key] = value
			return nil
		}

		nestedObj, ok := obj[key]
		if !ok {
			nestedObj = make(map[string]interface{})
			obj[key] = nestedObj
		}

		if childObj, ok := nestedObj.(map[string]interface{}); ok {
			obj = childObj
		} else {
			return fmt.Errorf("path not found: %s", strings.Join(keys[:i+1], "."))
		}
	}

	return nil
}

func getObjecKeys(obj map[string]interface{}) []string {
	keys := []string{}
	for k := range obj {
		keys = append(keys, k)
	}
	return keys
}

func splitStringToArray(str string, split string) []string {
	items := strings.Split(str, ",")
	resut := []string{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			resut = append(resut, item)
		}
	}
	return resut
}
