package manual_trigger

import (
	"context"
	_ "embed"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	// Category is the category of ManualTriggerNode.
	Category = structs.CategoryTrigger

	// Name is the name of ManualTriggerNode.
	Name = "n8n-nodes-base.manualTrigger"
)

var (
	//go:embed node.json
	rawJson []byte
)

type ManualTrigger struct {
	spec *structs.WorkflowNodeSpec
}

func init() {
	trigger := &ManualTrigger{
		spec: &structs.WorkflowNodeSpec{},
	}
	trigger.spec.JsonConfig = rawJson
	trigger.spec.GenerateSpec()

	core.Register(trigger)
}

func (trigger *ManualTrigger) Category() structs.NodeObjectCategory {
	return Category
}

func (trigger *ManualTrigger) Name() string {
	return Name
}

func (trigger *ManualTrigger) DefaultSpec() interface{} {
	return trigger.spec
}

func (trigger *ManualTrigger) Execute(ctx context.Context, _ *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	return core.GenerateSuccessResponse(structs.NodeData{
		structs.NodeSingleData{},
	}, []structs.NodeData{})
}
