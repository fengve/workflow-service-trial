package core

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sugerio/workflow-service-trial/shared/structs"

	"github.com/valyala/fasthttp"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core/internalhooks"
)

type WorkflowExecute struct {
	ctx              context.Context
	needDelete       bool
	WorkflowId       string
	ExecutionId      int32
	AdditionalData   *structs.WorkflowExecuteAdditionalData
	Mode             structs.WorkflowExecutionMode
	RunExecutionData *structs.WorkflowRunExecutionData
	Status           atomic.Pointer[structs.WorkflowExecutionStatus]
}

type NodeToAdd struct {
	Node        *structs.WorkflowNode
	Data        structs.NodeData
	OutputIndex int
}

func NewWorkflowExecute(ctx context.Context, additionalData *structs.WorkflowExecuteAdditionalData,
	mode structs.WorkflowExecutionMode) *WorkflowExecute {
	// TODO: add a parameter runExecutionData if needed, currently use a empty input

	runExecutionData := &structs.WorkflowRunExecutionData{
		StartData: &structs.WorkflowRunExecutionStartData{},
		ResultData: &structs.WorkflowRunExecutionResultData{
			RunData: make(map[string][]*structs.WorkflowExecutionTaskData),
		},
		ExecutionData: &structs.WorkflowRunExecutionExecutionData{},
	}
	wfe := &WorkflowExecute{
		ctx:              ctx,
		AdditionalData:   additionalData,
		Mode:             mode,
		RunExecutionData: runExecutionData,
	}
	return wfe
}

func GetBaseAdditionalData() *structs.WorkflowExecuteAdditionalData {
	// TODO: need userId parameter
	return &structs.WorkflowExecuteAdditionalData{}
}

func GetAdditionalDataWithHooks(
	ctx context.Context,
	mode structs.WorkflowExecutionMode,
	workflowEntity *structs.WorkflowEntity,
	userId string) (*structs.WorkflowExecuteAdditionalData, int, error) {
	executionData := structs.WorkflowExecutionDataProcess{
		ExecutionMode: mode,
		ExecutionData: nil, // empty by default
		SessionId:     "",  // empty by default
		WorkflowData:  workflowEntity,
		UserId:        userId,
	}

	executionId, err := GetActiveExecutions().AddExecution(ctx, &executionData, 0)
	if err != nil {
		return nil, 0, err
	}

	additionalData := GetBaseAdditionalData()
	additionalData.Hooks = GetWorkflowHooksMain(strconv.Itoa(executionId))
	additionalData.Hooks.Mode = executionData.ExecutionMode
	additionalData.Hooks.RetryOf = executionData.RetryOf
	additionalData.Hooks.WorkflowData = workflowEntity

	return additionalData, executionId, nil
}

func GetAdditionalDataWithTestWebHooks(mode structs.WorkflowExecutionMode,
	workflowEntity *structs.WorkflowEntity) (*structs.WorkflowExecuteAdditionalData, string, error) {
	executionData := structs.WorkflowExecutionDataProcess{
		ExecutionMode: mode,
		ExecutionData: nil, // empty by default
		SessionId:     "",  // empty by default
		WorkflowData:  workflowEntity,
	}

	executionId, err := GetActiveExecutions().AddTestWebhookExecution(&executionData)
	if err != nil {
		return nil, "", err
	}
	return GetBaseAdditionalData(), executionId, nil
}

func saveExecutionProgress(
	ctx context.Context,
	workflowEntity *structs.WorkflowEntity,
	executionId string,
	nodeName string,
	taskData *structs.WorkflowExecutionTaskData,
	executionData *structs.WorkflowRunExecutionData,
	sessionId string) {

	execId, err := strconv.Atoi(executionId)
	if err != nil {
		return
	}
	fullExecutionData, err := GetWorkflowExecution(ctx, int32(execId))
	if err != nil {
		Errorf("failed to get workflow execution entity: %v", err)
		return
	}
	if fullExecutionData == nil {
		return
	}
	if fullExecutionData.Finished {
		return
	}
	if fullExecutionData.Data == nil {
		fullExecutionData.Data = &structs.WorkflowRunExecutionData{
			StartData: &structs.WorkflowRunExecutionStartData{},
			ResultData: &structs.WorkflowRunExecutionResultData{
				RunData: make(map[string][]*structs.WorkflowExecutionTaskData),
			},
			ExecutionData: &structs.WorkflowRunExecutionExecutionData{},
		}
	}

	runExecutionData := fullExecutionData.Data

	if runExecutionData.ResultData == nil {
		runExecutionData.ResultData = &structs.WorkflowRunExecutionResultData{
			RunData: make(map[string][]*structs.WorkflowExecutionTaskData),
		}
	}

	if runData, ok := runExecutionData.ResultData.RunData[nodeName]; ok {
		runExecutionData.ResultData.RunData[nodeName] = append(runData, taskData)
	} else {
		runExecutionData.ResultData.RunData[nodeName] = []*structs.WorkflowExecutionTaskData{taskData}
	}
	runExecutionData.ExecutionData = executionData.ExecutionData
	runExecutionData.ResultData.LastNodeExecuted = nodeName

	fullExecutionData.Status = "running"

	err = UpdateWorkflowExecutionEntityAndData(ctx, int32(execId), fullExecutionData)
	if err != nil {
		Errorf("failed to update workflow execution entity: %v", err)
		return
	}
}

func determineFinalExecutionStatus(runData *structs.Run) structs.WorkflowExecutionStatus {
	workflowHasCrashed := runData.Status == structs.WorkflowExecutionStatus_Crashed
	workflowWasCanceled := runData.Status == structs.WorkflowExecutionStatus_Canceled
	workflowHasFailed := runData.Status == structs.WorkflowExecutionStatus_Failed
	workflowDidSucceed := runData.Data.ResultData.Error == "" && !workflowHasCrashed && !workflowWasCanceled && !workflowHasFailed

	var workflowStatusFinal structs.WorkflowExecutionStatus

	if workflowDidSucceed {
		workflowStatusFinal = structs.WorkflowExecutionStatus_Success
	} else {
		workflowStatusFinal = structs.WorkflowExecutionStatus_Failed
	}

	if workflowHasCrashed {
		workflowStatusFinal = structs.WorkflowExecutionStatus_Crashed
	}
	if workflowWasCanceled {
		workflowStatusFinal = structs.WorkflowExecutionStatus_Canceled
	}
	if runData.WaitTill != nil {
		workflowStatusFinal = structs.WorkflowExecutionStatus_Waiting
	}
	runData.Status = workflowStatusFinal
	return workflowStatusFinal
}

func executeErrorWorkflow(workflowEntity *structs.WorkflowEntity, fullRunData *structs.Run,
	mode structs.WorkflowExecutionMode, executionId string, retryOf string) {
	// TODO
}

func saveWorkflowAfterExecutionData(
	ctx context.Context,
	hooks *structs.WorkflowHooks,
	workflowEntity *structs.WorkflowEntity,
	fullRunData *structs.Run,
) {
	workflowStatusFinal := determineFinalExecutionStatus(fullRunData)
	if workflowStatusFinal != structs.WorkflowExecutionStatus_Success {
		executeErrorWorkflow(workflowEntity, fullRunData, hooks.Mode, hooks.ExecutionId, hooks.RetryOf)
	}
	executionId, err := strconv.Atoi(hooks.ExecutionId)
	if err != nil {
		return
	}
	fullExecutionData, err := GetWorkflowExecution(ctx, int32(executionId))
	fullExecutionData.Status = workflowStatusFinal
	fullExecutionData.Finished = fullRunData.Finished
	fullExecutionData.StoppedAt = fullRunData.StoppedAt
	fullExecutionData.WaitTill = fullRunData.WaitTill
	fullExecutionData.Data.ResultData.Error = fullRunData.Data.ResultData.Error

	if fullRunData.NeedDelete {
		id, err := strconv.Atoi(hooks.ExecutionId)
		if err == nil {
			err = DeleteWorkflowExecutionAndData(ctx, workflowEntity.ID, int32(id))
			if err != nil {
				Errorf("failed to delete workflow execution entity: %v", err)
			}
		}
		return
	}

	err = UpdateWorkflowExecutionEntityAndData(ctx, int32(executionId), fullExecutionData)
	if err != nil {
		Errorf("failed to update workflow execution entity: %v", err)
	}
}

func GetWorkflowHooksMain(executionId string) structs.WorkflowHooks {
	return structs.WorkflowHooks{
		ExecutionId: executionId,
		HookFunctions: structs.WorkflowExecuteHooks{
			NodeExecuteAfter: []func(context.Context, *structs.WorkflowHooks, string, *structs.NodeExecutionResult, *structs.WorkflowExecutionTaskData, *structs.WorkflowRunExecutionData){
				func(ctx context.Context, hooks *structs.WorkflowHooks, nodeName string, result *structs.NodeExecutionResult, taskData *structs.WorkflowExecutionTaskData, executionData *structs.WorkflowRunExecutionData) {
					internalhooks.OnNodeAfterExecute()
				},
				func(ctx context.Context, hooks *structs.WorkflowHooks, nodeName string, result *structs.NodeExecutionResult, taskData *structs.WorkflowExecutionTaskData, executionData *structs.WorkflowRunExecutionData) {
					saveExecutionProgress(ctx, hooks.WorkflowData, hooks.ExecutionId, nodeName, taskData, executionData, hooks.SessionId)
				},
			},
			NodeExecuteBefore: []func(context.Context, *structs.WorkflowHooks, string){
				func(ctx context.Context, hooks *structs.WorkflowHooks, nodeName string) {
					internalhooks.OnNodeBeforeExecute()
				},
			},
			WorkflowExecuteAfter: []func(context.Context, *structs.WorkflowHooks, *structs.Run){
				func(ctx context.Context, hooks *structs.WorkflowHooks, fullRunData *structs.Run) {
					internalhooks.OnWorkflowAfterExecute()
				},
				func(ctx context.Context, hooks *structs.WorkflowHooks, fullRunData *structs.Run) {
					saveWorkflowAfterExecutionData(ctx, hooks, hooks.WorkflowData, fullRunData)
				},
			},
			WorkflowExecuteBefore: []func(context.Context, *structs.WorkflowHooks, *structs.WorkflowEntity){
				func(ctx context.Context, hooks *structs.WorkflowHooks, workflowEntity *structs.WorkflowEntity) {
					internalhooks.OnWorkflowBeforeExecute()
				},
			},
			SendResponse: []func(context.Context, *structs.WorkflowHooks, *fasthttp.Response){},
		},
	}
}

func (w *WorkflowExecute) Run(ctx context.Context, workflowEntity *structs.WorkflowEntity) error {
	return w.RunFromNode(ctx, workflowEntity, nil)
}

func (w *WorkflowExecute) RunFromNode(
	ctx context.Context,
	workflowEntity *structs.WorkflowEntity,
	startNode *structs.WorkflowNode,
) error {
	err := w.execute(ctx, workflowEntity, startNode)
	if err != nil {
		return err
	}
	return nil
}

// sort nodes order by position from bottom right to top left
// after put into the execution stack is Last in First out
// will execute nodes from top left to bottom right if there are multiple nodes can be run
func sortNodesByPosition(nodes []NodeToAdd) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Node.Position[1] > nodes[j].Node.Position[1] {
			return true
		} else if nodes[i].Node.Position[1] < nodes[j].Node.Position[1] {
			return false
		} else {
			return nodes[i].Node.Position[0] >= nodes[j].Node.Position[0]
		}
	})
}

func (w *WorkflowExecute) execute(
	ctx context.Context, workflowEntity *structs.WorkflowEntity, startNode *structs.WorkflowNode) error {

	nodeNameDict := make(map[string]*structs.WorkflowNode, len(workflowEntity.Nodes))
	triggerNodeList := make([]*structs.WorkflowNode, 0)
	for idx := range workflowEntity.Nodes {
		node := workflowEntity.Nodes[idx]
		nodeNameDict[node.Name] = &node
		nodeTypeSuffix := strings.ToLower(node.Type)
		// If the node is disabled, skip it.
		if node.Disabled {
			continue
		}
		if strings.Contains(nodeTypeSuffix, "manual") ||
			strings.Contains(nodeTypeSuffix, "webhook") ||
			strings.Contains(nodeTypeSuffix, "trigger") {
			triggerNodeList = append(triggerNodeList, &node)
		}
	}

	startAt := time.Now()
	status := structs.WorkflowExecutionStatus_Running
	w.Status.Store(&status)
	w.AdditionalData.Hooks.ExecutionHookFunctionsWorkflowExecuteBefore(ctx, workflowEntity)

	if w.RunExecutionData.ExecutionData.WaitingExecution == nil {
		w.RunExecutionData.ExecutionData.WaitingExecution = make(map[string][]structs.NodeData)
	}

	if w.RunExecutionData.ExecutionData.WaitingExecutionSource == nil {
		w.RunExecutionData.ExecutionData.WaitingExecutionSource = make(map[string][]structs.ExecutionSourceData)
	}

	if w.RunExecutionData.ExecutionData.NodeExecutionStack == nil {
		if startNode != nil {
			w.RunExecutionData.ExecutionData.NodeExecutionStack = structs.NewNodeExecStack([]*structs.WorkflowNode{startNode})
		} else {
			w.RunExecutionData.ExecutionData.NodeExecutionStack = structs.NewNodeExecStack(triggerNodeList) // loop nodeExecutionStack node
		}
	}

	var taskData *structs.WorkflowExecutionTaskData
	var startTime int64
	finished := true

	for w.RunExecutionData.ExecutionData.NodeExecutionStack.Nodes.Len() != 0 {
		// TODO cancel on timeout
		if statusPtr := w.Status.Load(); statusPtr != nil && *statusPtr == structs.WorkflowExecutionStatus_Canceled {
			canceledStatus := structs.WorkflowExecutionStatus_Canceled
			w.Status.Store(&canceledStatus)
			fullRunData := w.getFullRunData(startAt)
			w.AdditionalData.Hooks.ExecutionHookFunctionsWorkflowExecutionAfter(ctx, fullRunData)
			return nil
		}
		// get head node as current exec node
		curNodeStack := w.RunExecutionData.ExecutionData.NodeExecutionStack.PopFront()

		if curNodeStack.Node.Disabled {
			// list of nodes to add to the stack.
			nodesToAdd := make([]NodeToAdd, 0)
			for outputIndex, connections := range workflowEntity.Connections[curNodeStack.Node.Name]["main"] {
				for _, connectionData := range connections {
					if len(curNodeStack.RunResultList) > outputIndex &&
						curNodeStack.RunResultList[outputIndex] != nil &&
						len(curNodeStack.RunResultList[outputIndex]) != 0 {
						// add node to the list for the next sorting and execution.
						nodesToAdd = append(nodesToAdd, NodeToAdd{
							Node: nodeNameDict[connectionData.Node],
							Data: curNodeStack.RunResultList[outputIndex],
						})
					}
				}
			}
			// Sort the nodes by position from top left to bottom right.
			sortNodesByPosition(nodesToAdd)
			for idx := range nodesToAdd {
				w.addNodeToBeExecuted(curNodeStack.Node.Name, workflowEntity, nodesToAdd[idx])
			}
			continue
		}
		// check delete node
		LowerNodeType := strings.ToLower(curNodeStack.Node.Type)
		if strings.Contains(LowerNodeType, "deleteexecution") {
			w.needDelete = true
		}
		// gen nodeInput
		nodeInput := &structs.NodeExecuteInput{
			WorkflowID:       workflowEntity.ID,
			Params:           curNodeStack.Node,
			Data:             curNodeStack.RunResultList,
			AdditionalData:   w.AdditionalData,
			RunExecutionData: w.RunExecutionData,
		}
		// node execute
		nodeObj := NewExecutor(curNodeStack.Node.Type).GetNode()

		startTime = time.Now().UnixMilli()
		w.AdditionalData.Hooks.ExecutionHookFunctionsNodeExecutionBefore(ctx, curNodeStack.Node.Name)

		result := nodeObj.Execute(ctx, nodeInput)
		// get next node and push to nodeExecutionStack
		resultList := w.getResultData(result, nodeObj.Category())
		// WaitingExecution saved the execution results for each node.
		w.RunExecutionData.ExecutionData.WaitingExecution[curNodeStack.Node.Name] = resultList

		taskData = &structs.WorkflowExecutionTaskData{
			StartTime:       startTime,
			ExecutionTime:   time.Now().UnixMilli() - startTime,
			ExecutionStatus: result.ExecutionStatus,
			Data:            map[string][]structs.NodeData{"main": resultList},
		}

		if result.Errors != nil && len(result.Errors) > 0 {
			executionError := result.Errors[0]
			taskData.Error = &structs.WorkflowExecutionError{
				Message:     executionError.Message,
				Description: executionError.Description,
				Node:        curNodeStack.Node,
				WorkflowId:  workflowEntity.ID,
			}
			w.RunExecutionData.ResultData.Error = executionError.Message
		}

		_, ok := w.RunExecutionData.ResultData.RunData[curNodeStack.Node.Name]
		if !ok {
			w.RunExecutionData.ResultData.RunData[curNodeStack.Node.Name] = make([]*structs.WorkflowExecutionTaskData, 0)
		}
		w.RunExecutionData.ResultData.RunData[curNodeStack.Node.Name] = append(
			w.RunExecutionData.ResultData.RunData[curNodeStack.Node.Name], taskData)
		w.AdditionalData.Hooks.ExecutionHookFunctionsNodeExecutionAfter(
			ctx, curNodeStack.Node.Name, result, taskData, w.RunExecutionData)
		w.RunExecutionData.ResultData.LastNodeExecuted = curNodeStack.Node.Name

		// check node ExecutionStatus
		if result.ExecutionStatus != structs.WorkflowExecutionStatus_Success {
			finished = false
			break // one execute not success, break
		}

		nodesToAddList := make([]NodeToAdd, 0)
		for outputIndex, connections := range workflowEntity.Connections[curNodeStack.Node.Name]["main"] {
			for _, connectionData := range connections {
				if len(resultList) > outputIndex && resultList[outputIndex] != nil {
					nodesToAddList = append(
						nodesToAddList,
						NodeToAdd{
							Node:        nodeNameDict[connectionData.Node],
							Data:        resultList[outputIndex],
							OutputIndex: outputIndex,
						})
				}
			}
		}
		sortNodesByPosition(nodesToAddList)
		for idx := range nodesToAddList {
			w.addNodeToBeExecuted(curNodeStack.Node.Name, workflowEntity, nodesToAddList[idx])
		}
	}

	fullRunData := w.getFullRunData(startAt)
	fullRunData.Finished = finished
	w.AdditionalData.Hooks.ExecutionHookFunctionsWorkflowExecutionAfter(ctx, fullRunData)

	return nil
}

func (w *WorkflowExecute) getConnectionByDestination(
	connections map[string]structs.WorkflowNodeConnections) map[string]structs.WorkflowNodeConnections {
	returnConnection := make(map[string]structs.WorkflowNodeConnections)
	for sourceNode := range connections {
		for connType := range connections[sourceNode] {
			for inputIndex := range connections[sourceNode][connType] {
				for _, connectionInfo := range connections[sourceNode][connType][inputIndex] {
					if _, ok := returnConnection[connectionInfo.Node]; !ok {
						returnConnection[connectionInfo.Node] = structs.WorkflowNodeConnections{}
					}
					if _, ok := returnConnection[connectionInfo.Node][connectionInfo.Type]; !ok {
						returnConnection[connectionInfo.Node][connectionInfo.Type] = make([][]structs.WorkflowConnection, 0)
					}
					maxIndex := len(returnConnection[connectionInfo.Node][connectionInfo.Type]) - 1
					for j := maxIndex; j < int(connectionInfo.Index); j++ {
						returnConnection[connectionInfo.Node][connectionInfo.Type] = append(
							returnConnection[connectionInfo.Node][connectionInfo.Type], []structs.WorkflowConnection{})
					}
					returnConnection[connectionInfo.Node][connectionInfo.Type][connectionInfo.Index] = append(
						returnConnection[connectionInfo.Node][connectionInfo.Type][connectionInfo.Index],
						structs.WorkflowConnection{
							Node:  sourceNode,
							Type:  connType,
							Index: int64(inputIndex),
						})
				}
			}
		}
	}

	return returnConnection
}

func (w *WorkflowExecute) addNodeToBeExecuted(
	previewNodeName string, workflowEntity *structs.WorkflowEntity, nodeToAdd NodeToAdd) {
	connectionByDestination := w.getConnectionByDestination(workflowEntity.Connections)
	if connectionByDestination[nodeToAdd.Node.Name]["main"] != nil &&
		len(connectionByDestination[nodeToAdd.Node.Name]["main"]) == 1 {
		w.RunExecutionData.ExecutionData.NodeExecutionStack.PushFront(&structs.NodeExecutionStackData{
			Node:          nodeToAdd.Node,
			RunResultList: []structs.NodeData{nodeToAdd.Data},
		})
		// WaitingExecutionSource saved the input data source (PreviewNode and its output index) of a Node
		// these info help to get the input data from WaitingExecution
		w.RunExecutionData.ExecutionData.WaitingExecutionSource[nodeToAdd.Node.Name] = []structs.ExecutionSourceData{
			{
				PreviousNode:       previewNodeName,
				PreviousNodeOutput: nodeToAdd.OutputIndex,
			},
		}
	} else {
		// next node have multiple inputData
		resultDataList := make([]structs.NodeData, 0)
		for _, connectionData := range connectionByDestination[nodeToAdd.Node.Name]["main"] {
			// foreach input connection check if the input data is ready
			if len(connectionData) > 0 {
				_, ok := w.RunExecutionData.ExecutionData.WaitingExecution[connectionData[0].Node]
				if !ok {
					// currently don't have all inputData, just skip. will execute after the last prev node.
					return
				}
				resultData := w.RunExecutionData.ExecutionData.WaitingExecution[connectionData[0].Node][connectionData[0].Index]
				resultDataList = append(resultDataList, resultData)

				_, ok = w.RunExecutionData.ExecutionData.WaitingExecutionSource[nodeToAdd.Node.Name]
				if !ok {
					w.RunExecutionData.ExecutionData.WaitingExecutionSource[nodeToAdd.Node.Name] = []structs.ExecutionSourceData{}
				}
				w.RunExecutionData.ExecutionData.WaitingExecutionSource[nodeToAdd.Node.Name] = append(
					w.RunExecutionData.ExecutionData.WaitingExecutionSource[nodeToAdd.Node.Name],
					structs.ExecutionSourceData{
						PreviousNode:       connectionData[0].Node,
						PreviousNodeOutput: int(connectionData[0].Index),
					},
				)
			}
		}
		w.RunExecutionData.ExecutionData.NodeExecutionStack.PushFront(&structs.NodeExecutionStackData{
			Node:          nodeToAdd.Node,
			RunResultList: resultDataList,
		})
	}
}

// TODO: remove TriggerData
func (w *WorkflowExecute) getResultData(result *structs.NodeExecutionResult,
	nodeCategory structs.NodeObjectCategory) []structs.NodeData {
	retList := make([]structs.NodeData, 0)
	switch nodeCategory {
	case structs.CategoryTrigger:
		retList = append(retList, result.TriggerData)
	case structs.CategoryExecutor:
		retList = result.ExecutorData
	default:
	}
	return retList
}

func (w *WorkflowExecute) getFullRunData(startAt time.Time) *structs.Run {
	stopAt := time.Now()
	// TODO this is not eventual status, n8n updates fullRunData status in processSuccessExecution
	var status structs.WorkflowExecutionStatus
	if statusPtr := w.Status.Load(); statusPtr != nil {
		status = *statusPtr
	}
	return &structs.Run{
		Data:       w.RunExecutionData,
		Mode:       w.Mode,
		StartedAt:  &startAt,
		StoppedAt:  &stopAt,
		Status:     status,
		NeedDelete: w.needDelete,
	}
}
