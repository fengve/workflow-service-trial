package api_test

// Command to run all tests under this package
// go test -v service/workflow_service/api/*_test.go

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/Netflix/go-env"

	"github.com/aws/aws-lambda-go/events"
	fiberAdapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	workflowTemporal "github.com/sugerio/workflow-service-trial/service/workflow_service/temporal"
	"github.com/sugerio/workflow-service-trial/shared"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	sharedRdsDb "github.com/sugerio/workflow-service-trial/shared/rds_db"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
	"github.com/teris-io/shortid"
	"go.temporal.io/sdk/client"
	temporalWorkflow "go.temporal.io/sdk/workflow"
)

var (
	logger             log.Logger
	sid                *shortid.Shortid
	workflowService    *api.WorkflowService
	environment        structs.Environment
	rdsDbQueries       *rdsDbLib.Queries
	readRdsDbQueries   *rdsDbLib.Queries
	sharedRdsDbQueries *sharedRdsDb.Queries
	awsSdkClients      *awsLib.AwsSdkClients
	temporalClient     client.Client
	testFiberLambda    *fiberAdapter.FiberLambda
)

// There can be only ONE TestMain for each package. It has only one function named Run(), which runs all the tests within the package.
func TestMain(m *testing.M) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// Set up Logger.
	logger = log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "app", "workflow-service-api-test")

	// set env
	structs.SetupEnvironmentVariables()
	defer structs.CleanupEnvironmentVariables()

	// Fetch Environment Variables
	environment = structs.Environment{}
	envSet, err := env.UnmarshalFromEnviron(&environment)
	if err != nil {
		panic(fmt.Sprintf("Failed to Fetch Environment Variables : %v", err))
	}
	environment.Extras = envSet

	testRdsDb, err := structs.CreateRdsDbClientForLocalTest(ctx)
	shared.Check(logger, "Failed to create test container for Postgres DB", err)
	rdsDbQueries = rdsDbLib.New(testRdsDb)

	temporalClient, err = client.Dial(
		client.Options{
			HostPort: client.DefaultHostPort,
			ContextPropagators: []temporalWorkflow.ContextPropagator{
				sharedTemporal.NewCommonContextPropagator(),
			},
			Namespace: "default",
		},
	)
	shared.Check(logger, "Failed to connect temporal client", err)
	defer temporalClient.Close()

	sid, err = shortid.New(2, shortid.DefaultABC, 2342)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize short ID generator: %v", err))
	}
	// Set up shared rdsDbQueries.
	sharedRdsDbQueries = sharedRdsDb.New(sid, rdsDbQueries, readRdsDbQueries)

	// Initiate AWS SDK Clients.
	awsSdkClients, err = awsLib.CreateAwsSdkClientsForTesting(ctx, logger, environment.AwsAuthMethod)
	shared.Check(logger, "Failed to initialize AwsSdkClients", err)

	temporalWorker, err := workflowTemporal.InitiateTemporalWorker(temporalClient)
	shared.Check(logger, "Failed to initiate Temporal Worker for workflow", err)
	defer temporalWorker.Stop()

	workflowService := api.NewWorkflowService(
		ctx,
		logger,
		&environment,
		awsSdkClients,
		rdsDbQueries,
		readRdsDbQueries,
		temporalClient,
	)
	err = workflowService.Start()
	if err != nil {
		panic(err)
	}
	defer workflowService.Close()

	// Register route bounds to HTTP methods with API handlers.
	testFiberLambda = workflowService.GetTestFiberAdapter()

	// Run all tests within the package.
	exitVal := m.Run()
	os.Exit(exitVal)
}

// Copied from service/marketplace_service/service_test.go
// Provides admin role to the built request
func GetAPIGatewayProxyRequest_CreateOrganization() (events.APIGatewayProxyRequest, error) {
	request := events.APIGatewayProxyRequest{}
	testRequestFile, err := os.ReadFile("../../marketplace_service/api/test_files/request_create_organization.json")
	if err != nil {
		return request, err
	}
	err = json.Unmarshal(testRequestFile, &request)
	return request, err
}

func createWorkflow(
	assert *require.Assertions, filePath string, orgId string,
) (workflowId string, createReq *structs.WorkflowManualRunRequest) {
	createReq = &structs.WorkflowManualRunRequest{}
	testFile, err := os.ReadFile(filePath)
	assert.Nil(err)
	err = json.Unmarshal(testFile, createReq)
	createReq.WorkflowData.SugerOrgId = orgId

	createReqBytes, _ := json.Marshal(createReq.WorkflowData)
	request := events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodPost,
		Path:           fmt.Sprintf("/workflow/org/%s/workflow", orgId),
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           string(createReqBytes),
		RequestContext: api.AuthorizerRequestContext,
	}

	response, err := testFiberLambda.Proxy(request)
	assert.Nil(err)
	var createResp structs.GetWorkflowResponse
	err = json.Unmarshal([]byte(response.Body), &createResp)
	assert.Nil(err, fmt.Sprint("response body:", response.Body))
	assert.NotEmpty(createResp.Data.VersionId)
	assert.NotEmpty(createResp.Data.ID)
	workflowId = createResp.Data.ID
	createReq.WorkflowData.ID = workflowId
	createReq.WorkflowData.VersionId = createResp.Data.VersionId
	return
}

func createWorkflowExecutionAndData_Testing(assert *require.Assertions, workflowEntity *structs.WorkflowEntity) int32 {
	entity, err := rdsDbQueries.CreateWorkflowExecutionEntity(
		context.Background(),
		rdsDbLib.CreateWorkflowExecutionEntityParams{WorkflowId: workflowEntity.ID})
	assert.Nil(err)
	workflowJson, err := json.Marshal(workflowEntity)
	assert.Nil(err)
	_, err = rdsDbQueries.CreateWorkflowExecutionData(
		context.Background(),
		rdsDbLib.CreateWorkflowExecutionDataParams{
			ExecutionId:  entity.ID,
			WorkflowData: workflowJson,
			Data:         "{}",
		})
	assert.Nil(err)
	return entity.ID
}

func addActiveExecution_Testing(assert *require.Assertions, ctx context.Context, workflowId string) int {
	workflowEntity, err := core.GetWorkflowEntityById(ctx, workflowId)
	assert.Nil(err)
	data := structs.WorkflowExecutionDataProcess{WorkflowData: workflowEntity}
	executionId, err := core.GetActiveExecutions().AddExecution(context.Background(), &data, 0)
	assert.Nil(err)
	return executionId
}
