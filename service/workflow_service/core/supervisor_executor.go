package core

import "github.com/sugerio/workflow-service-trial/shared/structs"

// Executor is the struct for node executor.
type Executor struct {
	nodeObj NodeObject
	spec    *structs.WorkflowNodeSpec
	params  *structs.WorkflowNode
}

// NewExecutor creates a new executor.
func NewExecutor(nodeName string) *Executor {
	e := &Executor{
		spec:   &structs.WorkflowNodeSpec{},
		params: &structs.WorkflowNode{},
	}
	e.nodeObj = MustNewNode(nodeName)
	return e
}

func MustNewNode(name string) NodeObject {
	nodeObject, ok := nodeObjectRegistry[name]
	if !ok {
		Warnf("Failed to found node %s", name)
		return nil
	}
	return nodeObject
}

func (e *Executor) GetNode() NodeObject {
	return e.nodeObj
}
