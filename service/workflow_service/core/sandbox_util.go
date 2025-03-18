package core

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

// export and fmt js error to "item is not defined [line 2]" like error
func HandleJavaScriptError(err error) error {
	Debugf("HandleJavaScriptError %v", err)

	var javaScriptError *goja.Exception
	message := err.Error()
	parts := strings.Split(message, " at ")

	if errors.As(err, &javaScriptError) {

		lineNum := resolveErrorLineNum(javaScriptError.String())
		return formatJavaScriptError(parts[0], lineNum)
	}

	lineNum := resolveErrorLineNum(err.Error())
	return formatJavaScriptError(parts[0], lineNum)

}

func resolveErrorLineNum(stack string) int {

	re := regexp.MustCompile(`\d+:\d+`)
	match := re.FindString(stack)
	items := strings.Split(match, ":")
	if len(items) >= 1 {
		// convert string to int and return line number
		lineNum, err := strconv.Atoi(items[0])
		if err != nil {
			return 0
		}
		return lineNum
	}
	return 0
}

func formatJavaScriptError(err string, lineNum int) error {
	if lineNum > 0 {
		return fmt.Errorf("%s [line %d]", err, lineNum)
	}
	return fmt.Errorf(err)
}

func StandardizeJavaScriptObject(item map[string]interface{}) map[string]interface{} {
	// remove fn, undefined
	for key, value := range item {

		switch v := value.(type) {
		case map[string]interface{}:
			item[key] = StandardizeJavaScriptObject(v)
		case []interface{}:
			for i, val := range v {
				switch val.(type) {
				case map[string]interface{}:
					v[i] = StandardizeJavaScriptObject(val.(map[string]interface{}))
				}
			}
		case func(), goja.FunctionCall, func(goja.FunctionCall) goja.Value: // see test `Sandbox run code with return obj has func inside` in sandbox_test.go
			delete(item, key)

		}

		if value == goja.Undefined() {
			delete(item, key)
		}

		// TODO: stringify Date, RegExp if needed
	}

	return item
}

/**
 * Stringify any non-standard JS objects (e.g. `Date`, `RegExp`) inside output items at any depth.
 * remove fn, undefined, null
 */
func StandardizeOutput(items []map[string]interface{}) structs.NodeData {

	for i, item := range items {
		items[i] = StandardizeJavaScriptObject(item)
	}

	return items
}

// wrap item in `json` key if they are not
func NormalizeItems(items []map[string]interface{}) structs.NodeData {
	for i, item := range items {
		if _, ok := item["json"]; !ok {
			items[i] = map[string]interface{}{
				"json": item,
			}
		}
	}

	return items
}

func getSandboxContextFromInput(input *structs.NodeExecuteInput) *SandboxContext {
	runData := make(map[string][]*structs.WorkflowExecutionTaskData)
	if input.RunExecutionData != nil && input.RunExecutionData.ResultData != nil && input.RunExecutionData.ResultData.RunData != nil {
		runData = input.RunExecutionData.ResultData.RunData
	}
	return &SandboxContext{
		Items:     GetInputData(input.Data),
		Params:    input.Params.Parameters,
		Functions: BuiltInFunctions,
		RunData:   runData,
	}
}
