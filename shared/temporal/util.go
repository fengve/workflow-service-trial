package temporal

import (
	"context"
	"fmt"
	temporalEnums "go.temporal.io/api/enums/v1"
	filterpb "go.temporal.io/api/filter/v1"
	"go.temporal.io/api/serviceerror"
	workflowv1 "go.temporal.io/api/workflow/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"golang.org/x/exp/slices"
	"strings"
)

var WorkflowClosedStatus = []temporalEnums.WorkflowExecutionStatus{
	temporalEnums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
	temporalEnums.WORKFLOW_EXECUTION_STATUS_CANCELED,
	temporalEnums.WORKFLOW_EXECUTION_STATUS_FAILED,
	temporalEnums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
	temporalEnums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
}

var WorkflowOpeningStatus = []temporalEnums.WorkflowExecutionStatus{
	temporalEnums.WORKFLOW_EXECUTION_STATUS_RUNNING,
	temporalEnums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
}

func TerminateWorkflowIfOpen(
	ctx context.Context,
	temporalClient client.Client,
	workflowId string,
	reason string) error {
	response, err := temporalClient.DescribeWorkflowExecution(ctx, workflowId, "")
	if err != nil {
		if _, ok := err.(*serviceerror.NotFound); ok { // not found. no need to terminate.
			return nil
		}
		return err
	}

	if response.WorkflowExecutionInfo != nil {
		if slices.Contains(WorkflowClosedStatus, response.WorkflowExecutionInfo.Status) {
			// The workflow is closed already, no need to terminate.
			return nil
		}

		return temporalClient.TerminateWorkflow(ctx, workflowId, "", reason)
	} else {
		return fmt.Errorf("failed to get the WorkflowExecutionInfo for workflow %s", workflowId)
	}
}

// Start a new workflow no matter the existing workflow status.
// If no existing workflow, start a new workflow.
// If the latest existing workflow is in status of RUNNING or CONTINUED_AS_NEW, terminate it and start a new workflow.
// If the latest existing workflow is in status of COMPLETED, FAILED, CANCELED, TERMINATED or TIMED_OUT, start a new workflow.
func StartWorkflow_Override(
	ctx context.Context,
	temporalClient client.Client,
	options *client.StartWorkflowOptions,
	workflow interface{},
	args ...interface{}) (client.WorkflowRun, error) {
	response, err := temporalClient.DescribeWorkflowExecution(ctx, options.ID, "")
	if err != nil {
		if _, ok := err.(*serviceerror.NotFound); ok { // not found. Start a new one.
			return temporalClient.ExecuteWorkflow(ctx, *options, workflow, args...)
		}
		return nil, err
	}

	if response.WorkflowExecutionInfo != nil {
		if slices.Contains(WorkflowClosedStatus, response.WorkflowExecutionInfo.Status) {
			// The existing workflow is closed, create a new one.
			return temporalClient.ExecuteWorkflow(ctx, *options, workflow, args...)
		} else {
			// The existing workflow is still opening (RUNNING or CONTINUED_AS_NEW), terminate it and create a new one.
			err = temporalClient.TerminateWorkflow(ctx, options.ID, "", "terminate for StartWorkflow_Override")
			if err != nil {
				return nil, err
			}
			return temporalClient.ExecuteWorkflow(ctx, *options, workflow, args...)
		}
	}
	return nil, fmt.Errorf("failed to get the WorkflowExecutionInfo for workflow %s", options.ID)
}

// List all open WorkflowExecutions of the given workflowType.
func ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	temporalClient client.Client,
	workflowType string) ([]workflowv1.WorkflowExecutionInfo, error) {
	results := []workflowv1.WorkflowExecutionInfo{}

	request := &workflowservice.ListOpenWorkflowExecutionsRequest{
		Namespace:       "default",
		MaximumPageSize: int32(100),
		Filters: &workflowservice.ListOpenWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: &filterpb.WorkflowTypeFilter{
				Name: workflowType,
			},
		},
	}
	for {
		response, err := temporalClient.ListOpenWorkflow(ctx, request)
		if err != nil {
			return nil, err
		}
		// if no more executions, then break.
		if len(response.Executions) == 0 {
			break
		}

		for _, execution := range response.Executions {
			results = append(results, *execution)
		}
		// if no more pages, then break.
		if len(response.NextPageToken) == 0 {
			break
		}
		// if there is more pages, then set the next page token and continue.
		request.NextPageToken = response.NextPageToken
	}

	return results, nil
}

// List all open WorkflowExecutions of the given orgId and workflowType.
func ListOpenWorkflowExecutionsByOrgAndType(
	ctx context.Context,
	temporalClient client.Client,
	orgId string,
	workflowType string) ([]workflowv1.WorkflowExecutionInfo, error) {
	openWorkflowExecutions, err := ListOpenWorkflowExecutionsByType(ctx, temporalClient, workflowType)
	if err != nil {
		return nil, err
	}

	// Filter the open workflow executions by containing the orgId in the workflowId.
	orgStr := fmt.Sprintf("orgId/%s", orgId)
	results := []workflowv1.WorkflowExecutionInfo{}
	for _, execution := range openWorkflowExecutions {
		if strings.Contains(execution.GetExecution().GetWorkflowId(), orgStr) {
			results = append(results, execution)
		}
	}

	return results, nil
}
