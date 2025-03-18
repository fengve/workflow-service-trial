package ifnode

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (

	// Category is the category of IfNode.
	Category = structs.CategoryExecutor

	// Name is the name of IfNode.
	Name = "n8n-nodes-base.if"
)

var (
	//go:embed node.json
	rawJson []byte
)

type (
	IfExecutor struct {
		spec *structs.WorkflowNodeSpec
	}

	ParameterOptions struct {
		IgnoreCase          bool `json:"ignoreCase"`
		LooseTypeValidation bool `json:"looseTypeValidation"`
	}
)

func init() {
	ie := &IfExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	ie.spec.JsonConfig = rawJson
	ie.spec.GenerateSpec()

	core.Register(ie)
}

func (ie *IfExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (ie *IfExecutor) Name() string {
	return Name
}

func (ie *IfExecutor) DefaultSpec() interface{} {
	return ie.spec
}

func (ie *IfExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	var trueResult, falseResult structs.NodeData
	items := core.GetInputData(input.Data)

	for itemIndex, item := range items {

		passRaw, err := core.GetNodeParameter(Name, "conditions", false, input, itemIndex, core.GetNodeParameterOptions{
			ExtractValue: true,
		})
		if err != nil {
			if !core.ContinueOnFail(input.Params) {
				return core.GenerateFailedResponse(Name, err)
			}
		}
		pass, ok := passRaw.(bool)
		if !ok {
			if !core.ContinueOnFail(input.Params) {
				return core.GenerateFailedResponse(Name, fmt.Errorf("conditions is not a boolean [itemIndex: %d]", itemIndex))
			}
		}

		if pass {
			trueResult = append(trueResult, item)
		} else {
			falseResult = append(falseResult, item)
		}

	}

	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{trueResult, falseResult})
}
