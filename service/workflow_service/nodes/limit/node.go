package limitnode

import (
	"context"
	_ "embed"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	Category = structs.CategoryExecutor
	Name     = "n8n-nodes-base.limit"
)

var (
	//go:embed node.json
	rawJson []byte

	//go:embed limit.svg
	icon []byte
)

type (
	LimitExecutor struct {
		spec *structs.WorkflowNodeSpec
	}
)

func init() {
	le := &LimitExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	le.spec.JsonConfig = rawJson
	le.spec.GenerateSpec()

	core.Register(le)
	core.RegisterEmbedIcons(Name, icon)
}

func (le *LimitExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (le *LimitExecutor) Name() string {
	return Name
}

func (le *LimitExecutor) DefaultSpec() interface{} {
	return le.spec
}

func (le *LimitExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	items := core.GetInputData(input.Data)
	result := items

	// default type from expression is int64
	maxItemsVal, err := core.GetNodeParameter(Name, "maxItems", 1, input, 0)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	maxItemInt64, err := core.ConvertToInt(maxItemsVal)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	maxItems := int(maxItemInt64)
	keep, err := core.GetNodeParameterAsBasicType(Name, "keep", "firstItems", input, 0)
	if err != nil {
		keep = "firstItems"
	}

	if maxItems > len(result) {
		return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
	}

	if keep == "firstItems" {
		result = result[:maxItems]
	} else {
		result = result[len(result)-maxItems:]
	}
	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
}
