package delete_execution

import (
	"context"
	_ "embed"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	// Category is the category of ManualTriggerNode.
	Category = structs.CategoryExecutor

	// Name is the name of ManualTriggerNode.
	Name = "n8n-nodes-base.deleteExecution"
)

var (
	//go:embed node.json
	rawJson []byte
)

type DeleteExecutionExecutor struct {
	spec *structs.WorkflowNodeSpec
}

func init() {
	executor := &DeleteExecutionExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	executor.spec.JsonConfig = rawJson
	executor.spec.GenerateSpec()

	core.Register(executor)
}

func (executor *DeleteExecutionExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (executor *DeleteExecutionExecutor) Name() string {
	return Name
}

func (executor *DeleteExecutionExecutor) DefaultSpec() interface{} {
	return executor.spec
}

func (executor *DeleteExecutionExecutor) Execute(ctx context.Context, _ *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{})
}
