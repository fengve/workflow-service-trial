package temporal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	gomock "github.com/golang/mock/gomock"
	"go.temporal.io/api/serviceerror"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"golang.org/x/exp/slices"
)

type MyTestReporter struct {
	t gomock.TestReporter
}

func (m *MyTestReporter) Errorf(format string, args ...interface{}) {
	if m.t == nil {
		panic(fmt.Sprintf(format, args...))
	}
	m.t.Errorf(format, args...)
}

func (m *MyTestReporter) Fatalf(format string, args ...interface{}) {
	if m.t == nil {
		panic(fmt.Sprintf(format, args...))
	}
	m.t.Fatalf(format, args...)
}

func CreateMockTemporalClient() (*MockClient, *gomock.Controller) {
	t := &MyTestReporter{}
	mockCtrl := gomock.NewController(t)
	// defer mockCtrl.Finish()
	temporalClient := NewMockClient(mockCtrl)
	temporalClient.
		EXPECT().
		Close().
		Times(1)
	temporalClient.
		EXPECT().
		CheckHealth(gomock.Any(), gomock.Any()).
		Return(nil, nil).
		AnyTimes()
	temporalClient.
		EXPECT().
		ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).
		AnyTimes()
	temporalClient.
		EXPECT().
		DescribeWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, serviceerror.NewNotFound("")).
		AnyTimes()
	temporalClient.
		EXPECT().
		ListOpenWorkflow(gomock.Any(), gomock.Any()).
		Return(&workflowservice.ListOpenWorkflowExecutionsResponse{}, nil).
		AnyTimes()

	return temporalClient, mockCtrl
}

func TerminateAllOpenWorkflows(ctx context.Context, client client.Client, namespace string, taskQueue string) error {
	listOpenWorkflowrequest := &workflowservice.ListOpenWorkflowExecutionsRequest{
		Namespace:       namespace,
		MaximumPageSize: *aws.Int32(200),
	}
	listOpenWorkflowResponse, err := client.ListOpenWorkflow(ctx, listOpenWorkflowrequest)
	if err != nil {
		return err
	}

	for _, execution := range listOpenWorkflowResponse.Executions {
		if execution.TaskQueue == taskQueue && slices.Contains(WorkflowOpeningStatus, execution.Status) {
			err = client.TerminateWorkflow(ctx, execution.Execution.WorkflowId, execution.Execution.RunId, "terminate before test")
			if err != nil {
				fmt.Println("failed to TerminateWorkflow", err)
			}
		}
	}
	return nil
}
