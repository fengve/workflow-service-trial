package switcher

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	// Category is the category of ManualTriggerNode.
	Category = structs.CategoryExecutor

	// Name is the name of ManualTriggerNode.
	Name = "n8n-nodes-base.switch"

	ModeExpression = "expression"
	ModeRules      = "rules"
)

var (
	//go:embed node.json
	rawJson []byte
)

type (
	SwitchExecutor struct {
		spec *structs.WorkflowNodeSpec
	}

	ParameterRule struct {
		Operation string      `json:"operation"`
		Value2    interface{} `json:"value2"`
		OutputKey string      `json:"outputkey"`
	}
)

func init() {
	se := &SwitchExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	se.spec.JsonConfig = rawJson
	se.spec.GenerateSpec()

	core.Register(se)
}

func (se *SwitchExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (se *SwitchExecutor) Name() string {
	return Name
}

func (se *SwitchExecutor) DefaultSpec() interface{} {
	return se.spec
}

func (se *SwitchExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {

	returnData := []structs.NodeData{}
	items := core.GetInputData(input.Data)

ItemLoop:
	for itemIndex, item := range items {
		mode, err := core.GetNodeParameterAsBasicType(Name, "mode", ModeRules, input, itemIndex)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		if mode == ModeExpression {
			outputsAmount, err := core.GetNodeParameterAsType(Name, "outputsAmount", 4, input, itemIndex)
			if err != nil {
				if core.ContinueOnFail(input.Params) {
					returnData = append(returnData, structs.NodeData{
						core.NewNodeSingleDataError(err, itemIndex),
					})
					continue ItemLoop
				} else {
					return core.GenerateFailedResponse(Name, err)
				}
			}
			// init returnData if it is the first item
			if itemIndex == 0 {
				returnData = make([]structs.NodeData, *outputsAmount)
			}
			outputIndex, err := core.GetNodeParameterAsType(Name, "output", -1, input, itemIndex)
			if err != nil {
				if core.ContinueOnFail(input.Params) {
					returnData = append(returnData, structs.NodeData{
						core.NewNodeSingleDataError(err, itemIndex),
					})
					continue ItemLoop
				} else {
					return core.GenerateFailedResponse(Name, err)
				}
			}

			// checkIndexRange
			if *outputIndex < 0 || *outputIndex >= *outputsAmount {
				err := fmt.Errorf(`the output index value %d is not allowed, it has to be between 0 and %d [item %d]`,
					*outputIndex, *outputsAmount-1, itemIndex,
				)
				if core.ContinueOnFail(input.Params) {
					returnData[0] = append(returnData[0],
						structs.NodeSingleData{"json": map[string]interface{}{"error": err.Error()}},
					)
					continue ItemLoop
				} else {
					return core.GenerateFailedResponse(Name, err)
				}
			}

			returnData[*outputIndex] = append(returnData[*outputIndex], item)

		} else if mode == ModeRules {
			rules, err := core.GetNodeParameterAsType(Name, "rules.rules", []ParameterRule{}, input, itemIndex)
			if err != nil {
				if core.ContinueOnFail(input.Params) {
					returnData = append(returnData, structs.NodeData{
						core.NewNodeSingleDataError(err, itemIndex),
					})
					continue ItemLoop
				} else {
					return core.GenerateFailedResponse(Name, err)
				}
			}
			for i, rule := range *rules {
				if rule.Operation == "" {
					(*rules)[i].Operation = "equal"
				}
			}
			// init returnData if it is the first item
			if itemIndex == 0 {
				returnData = make([]structs.NodeData, len(*rules))
			}

			value1, err := core.GetNodeParameter(Name, "value1", nil, input, itemIndex)

			if err != nil {
				if core.ContinueOnFail(input.Params) {
					returnData = append(returnData, structs.NodeData{
						core.NewNodeSingleDataError(err, itemIndex),
					})
					continue ItemLoop
				} else {
					return core.GenerateFailedResponse(Name, err)
				}
			}

			for ruleIndex, rule := range *rules {

				compareOperationResult := compareOperation(rule.Operation, value1, rule.Value2)
				if compareOperationResult {
					// If the rule matched the item
					// add to the output
					returnData[ruleIndex] = append(returnData[ruleIndex], item)
					// continue with the next item
					continue ItemLoop
				}

			}
			// fallback output
			outputIndex, err := core.GetNodeParameterAsType(Name, "fallbackOutput", -1, input, itemIndex)
			if err != nil {
				core.Debugf("Error getting fallbackOutput: %v  [item %d]", err, itemIndex)
			}

			if *outputIndex != -1 {
				// checkIndexRange
				if *outputIndex < 0 || *outputIndex >= len(*rules) {
					err := fmt.Errorf(`the output index value %d is not allowed, it has to be between 0 and %d [item %d]`,
						*outputIndex, len(*rules)-1, itemIndex,
					)

					if core.ContinueOnFail(input.Params) {
						returnData[0] = append(returnData[0],
							structs.NodeSingleData{"json": map[string]interface{}{"error": err.Error()}},
						)
						continue
					} else {
						return core.GenerateFailedResponse(Name, err)
					}

				}
				returnData[*outputIndex] = append(returnData[*outputIndex], item)
			}
		}
	}

	return core.GenerateSuccessResponse(structs.NodeData{}, returnData)
}

func compareOperation(operation string, value1 interface{}, value2 interface{}) bool {
	switch operation {
	case "after":
		v1, err1 := core.ConvertToDate(value1)
		v2, err2 := core.ConvertToDate(value2)
		if err1 != nil || err2 != nil {
			return false
		}

		return v1.After(v2)

	case "before":
		v1, err1 := core.ConvertToDate(value1)
		v2, err2 := core.ConvertToDate(value2)
		if err1 != nil || err2 != nil {
			return false
		}
		return v1.Before(v2)

	case "contains":
		return strings.Contains(value1.(string), value2.(string))

	case "notContains":
		return !strings.Contains(value1.(string), value2.(string))

	case "endsWith":
		return strings.HasSuffix(value1.(string), value2.(string))

	case "notEndsWith":
		return !strings.HasSuffix(value1.(string), value2.(string))

	case "equal":
		v1, err1 := json.Marshal(value1)
		v2, err2 := json.Marshal(value2)
		if err1 != nil || err2 != nil {
			return false
		}
		return string(v1) == string(v2)

	case "notEqual":
		v1, err1 := json.Marshal(value1)
		v2, err2 := json.Marshal(value2)
		if err1 != nil || err2 != nil {
			return false
		}
		return string(v1) != string(v2)

	case "larger", "largerEqual", "smaller", "smallerEqual":
		v1, err1 := core.ConvertToFloat(value1)
		v2, err2 := core.ConvertToFloat(value2)
		if err1 != nil || err2 != nil {
			return false
		}
		switch operation {
		case "larger":
			return v1 > v2

		case "largerEqual":
			return v1 >= v2

		case "smaller":
			return v1 < v2

		case "smallerEqual":
			return v1 <= v2
		}

	case "startsWith":
		return strings.HasPrefix(value1.(string), value2.(string))

	case "notStartsWith":
		return !strings.HasPrefix(value1.(string), value2.(string))

	case "regex":
		regexPattern := value2.(string)
		regexOptions := ""

		// Detect if the pattern includes regex options
		regexMatch := regexp.MustCompile(`^/(.*?)/([gimusy]*)$`).FindStringSubmatch(regexPattern)

		var regex *regexp.Regexp

		if regexMatch != nil {
			// Extract the pattern and options
			regexPattern = regexMatch[1]
			regexOptions = regexMatch[2]

			// Apply 'i' and 's' flags if present
			if strings.Contains(regexOptions, "i") {
				regexPattern = "(?i)" + regexPattern
			}
			if strings.Contains(regexOptions, "s") {
				regexPattern = "(?s)" + regexPattern
			}
		}

		regex = regexp.MustCompile(regexPattern)

		return regex.MatchString(value1.(string))

	case "notRegex":
		regexPattern := value2.(string)
		regexOptions := ""

		// Detect if the pattern includes regex options
		regexMatch := regexp.MustCompile(`^/(.*?)/([gimusy]*)$`).FindStringSubmatch(regexPattern)

		var regex *regexp.Regexp

		if regexMatch != nil {
			// Extract the pattern and options
			regexPattern = regexMatch[1]
			regexOptions = regexMatch[2]

			// Apply 'i' and 's' flags if present
			if strings.Contains(regexOptions, "i") {
				regexPattern = "(?i)" + regexPattern
			}
			if strings.Contains(regexOptions, "s") {
				regexPattern = "(?s)" + regexPattern
			}
		}

		regex = regexp.MustCompile(regexPattern)
		return !regex.MatchString(value1.(string))
	}

	return false
}
