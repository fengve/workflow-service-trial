package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sugerio/workflow-service-trial/shared/structs"
)

type ExpressionEvaluator struct {
	Sandbox *Sandbox
}

func (ee *ExpressionEvaluator) EvaluateExpression(expression string, itemIndex int) (interface{}, error) {
	// checks if the expression is empty or fixed
	if len(expression) == 0 {
		return expression, nil
	}
	if expression[0] == '=' {
		expression = expression[1:]
	} else {
		return expression, nil
	}
	if !strings.Contains(expression, "{{") {
		return expression, nil
	}
	// ---------------- Expression ----------------
	chunks, err := ee.splitExpression(expression)

	if err != nil {
		return nil, err
	}

	chunks = removeEmptyTextChunk(chunks)

	res, err := ee.renderExpression(chunks, itemIndex)

	return res, err
}

type ChunkType string

const (
	TextType ChunkType = "text"
	CodeType ChunkType = "code"
)

type ExpressionChunk struct {
	Text string
	Type ChunkType
}

// escapeCode removes the escape character before the "}}" in the string.
func escapeCode(text string) string {
	return strings.ReplaceAll(text, `\\}}`, `}}`)
}

const (
	OpenBracket  = "{"
	CloseBracket = "}"
)

var RegOpenBracket = regexp.MustCompile(`(\\|)\{\{`)
var RegCloseBracket = regexp.MustCompile(`(\\|)\}\}`)

func getActiveRegex(searchingFor string) *regexp.Regexp {
	if searchingFor == OpenBracket {
		return RegOpenBracket
	}
	return RegCloseBracket
}

func (ee *ExpressionEvaluator) splitExpression(expression string) ([]ExpressionChunk, error) {
	// Split the expression into chunks

	var chunks []ExpressionChunk

	searchingFor := OpenBracket

	buffer := ""
	index := 0

	for index < len(expression) {
		expr := expression[index:]

		activeRegex := getActiveRegex(searchingFor)
		loc := activeRegex.FindStringSubmatchIndex(expr)

		if loc == nil { // no more brackets
			buffer += expr
			if searchingFor == OpenBracket {
				chunks = append(chunks, ExpressionChunk{
					Type: TextType,
					Text: buffer,
				})
			} else {
				chunks = append(chunks, ExpressionChunk{
					Type: CodeType,
					Text: escapeCode(buffer),
				})
			}
			break
		}
		esc := loc[2] != -1 && expr[loc[2]:loc[2]+1] == "\\"
		if esc {
			buffer += expr[:loc[1]]
			index += loc[1]
		} else {
			buffer += expr[:loc[0]]
			if searchingFor == OpenBracket {
				chunks = append(chunks, ExpressionChunk{
					Type: TextType,
					Text: buffer,
				})
				searchingFor = CloseBracket
			} else {
				chunks = append(chunks, ExpressionChunk{
					Type: CodeType,
					Text: escapeCode(buffer),
				})
				searchingFor = OpenBracket
			}
			buffer = ""
			index += loc[1]
		}
	}
	return chunks, nil
}

func removeEmptyTextChunk(chunks []ExpressionChunk) []ExpressionChunk {
	var result []ExpressionChunk
	for _, chunk := range chunks {
		if chunk.Text != "" || chunk.Type == CodeType {
			result = append(result, chunk)
		}
	}
	return result

}

func (ee *ExpressionEvaluator) renderExpression(chunks []ExpressionChunk, itemIndex int) (interface{}, error) {

	// single item
	if len(chunks) == 1 && chunks[0].Type == CodeType {
		// return original type
		return ee.parseCode(chunks[0].Text, itemIndex)
	}

	// parse the expression from chunks
	var resultData []string
	for _, chunk := range chunks {
		if chunk.Type == CodeType {
			res, err := ee.parseCode(chunk.Text, itemIndex)
			if err != nil {
				returnErrStr := fmt.Sprintf("Error parsing code: %v", err)
				resultData = append(resultData, returnErrStr)
				Errorf("Error parsing code: %v\n%s", err, chunk.Text)
				// we don't return error here, continue to evaluate the rest of the expression
			}
			resStr, err := convertAnyValueToString(res)
			if err != nil {
				returnErrStr := fmt.Sprintf("Error converting value to string: %v", err)
				resultData = append(resultData, returnErrStr)
				Errorf("Error converting value to string: %v\n%s", err, chunk.Text)
				// we don't return error here, continue to evaluate the rest of the expression
			}

			resultData = append(resultData, resStr)
		} else {
			resultData = append(resultData, chunk.Text)
		}
	}
	return strings.Join(resultData, ""), nil
}

func (ee *ExpressionEvaluator) parseCode(code string, itemIndex int) (interface{}, error) {
	// Evaluate the code chunk
	return ee.Sandbox.RunCode(code, itemIndex)
}

func convertAnyValueToString(val interface{}) (string, error) {
	// convert value to string
	switch v := val.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
		// date
	case time.Time:
		return v.UTC().Format(time.RFC3339), nil
	case nil:
		return "", nil
	default:
		// case []interface{}:
		// case map[string]interface{}:
		// case structs.NodeData:
		// case structs.NodeSingleData:

		r, err := json.Marshal(v)
		return string(r), err
	}
}

// NewExpressionEvaluator creates a new ExpressionEvaluator
// sbc can be nil
func NewExpressionEvaluator(sbc *SandboxContext) *ExpressionEvaluator {
	sb := Sandbox{
		Lang:    CodeLanguage,
		Context: sbc,
	}
	sb.Initialize()
	return &ExpressionEvaluator{
		Sandbox: &sb,
	}
}

// Get the parameter value from the input with Javascript expression evaluation.
func GetParameterValue(
	value interface{},
	parameterName string,
	input *structs.NodeExecuteInput,
	itemIndex int,
	returnObjectAsString bool,
) (interface{}, error) {
	// If the input value is nil, return nil
	if value == nil {
		return nil, nil
	}

	sandboxContext := getSandboxContextFromInput(input)
	evaluator := NewExpressionEvaluator(sandboxContext)
	returnData, err := resolveParameterValue(value, evaluator, itemIndex)
	if err != nil {
		return nil, err
	}

	if returnObjectAsString && IsMap(value) {
		return convertAnyValueToString(returnData)
	}

	return returnData, nil
}

func resolveParameterValue(value interface{}, eval *ExpressionEvaluator, itemIndex int) (interface{}, error) {

	if IsMap(value) {
		// Data is an object
		parameterMap, err := ConvertToInterfaceMap(value)
		if err != nil {
			return nil, err
		}
		returnData := make(map[string]interface{})
		for key, value := range parameterMap {
			resolvedValue, err := resolveParameterValue(value, eval, itemIndex)
			if err != nil {
				return nil, err
			}
			returnData[key] = resolvedValue
		}
		return returnData, nil

	}

	if IsArray(value) {
		parameterArray, err := ConvertToInterfaceArray(value)
		// Data is an array
		if err != nil {
			return nil, err
		}
		returnData := make([]interface{}, len(parameterArray))
		for i, item := range parameterArray {
			resolvedValue, err := resolveParameterValue(item, eval, itemIndex)
			if err != nil {
				// TODO continue on error?
				return nil, err
			}
			returnData[i] = resolvedValue
		}

		return returnData, nil
	}

	return resolveSimpleParameterValue(value, eval, itemIndex)
}

func resolveSimpleParameterValue(value interface{}, eval *ExpressionEvaluator, itemIndex int) (interface{}, error) {
	valueStr, ok := value.(string)
	if !ok {
		// return if value is not a string
		return value, nil
	}
	return eval.EvaluateExpression(valueStr, itemIndex)
}
