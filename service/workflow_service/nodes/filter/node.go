package filternode

import (
	"context"
	_ "embed"
	"errors"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (

	// Category is the category of FilterNode.
	Category = structs.CategoryExecutor

	// Name is the name of FilterNode.
	Name = "n8n-nodes-base.filter"
)

var (
	//go:embed node.json
	rawJson []byte
)

type (
	FilterExecutor struct {
		spec *structs.WorkflowNodeSpec
	}

	ParameterOptions struct {
		IgnoreCase          bool `json:"ignoreCase"`
		LooseTypeValidation bool `json:"looseTypeValidation"`
	}
)

func init() {
	fe := &FilterExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	fe.spec.JsonConfig = rawJson
	fe.spec.GenerateSpec()

	core.Register(fe)
}

func (fe *FilterExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (fe *FilterExecutor) Name() string {
	return Name
}

func (fe *FilterExecutor) DefaultSpec() interface{} {
	return fe.spec
}

func (fe *FilterExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	var keptItems, discardedItems structs.NodeData
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
				return core.GenerateFailedResponse(Name, errors.New("condition is not a boolean"))
			}
		}

		if pass {
			keptItems = append(keptItems, item)
		} else {
			discardedItems = append(discardedItems, item)
		}

	}

	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{keptItems, discardedItems})
}
