package nodes_test

// Command to run all tests under this package
// go test -v service/workflow_service/nodes_test/*_test.go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/Netflix/go-env"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aws/aws-lambda-go/events"
	fiberAdapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/go-kit/log"
	"github.com/teris-io/shortid"
	"go.temporal.io/sdk/client"
	temporalWorkflow "go.temporal.io/sdk/workflow"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	workflowTemporal "github.com/sugerio/workflow-service-trial/service/workflow_service/temporal"
	"github.com/sugerio/workflow-service-trial/shared"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	sharedRdsDb "github.com/sugerio/workflow-service-trial/shared/rds_db"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
)

var (
	logger                     log.Logger
	sid                        *shortid.Shortid
	workflowService            *api.WorkflowService
	environment                structs.Environment
	rdsDbQueries               *rdsDbLib.Queries
	readRdsDbQueries           *rdsDbLib.Queries
	sharedRdsDbQueries         *sharedRdsDb.Queries
	awsSdkClients              *awsLib.AwsSdkClients
	temporalClient             client.Client
	testFiberLambda            *fiberAdapter.FiberLambda
	testMarketplaceFiberLambda *fiberAdapter.FiberLambda
)

// There can be only ONE TestMain for each package. It has only one function named Run(), which runs all the tests within the package.
func TestMain(m *testing.M) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// Set up Logger.
	logger = log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "app", "workflow-service-test")

	logger.Log("msg", "Start to setup test environment.")

	// set env
	structs.SetupEnvironmentVariables()
	defer structs.CleanupEnvironmentVariables()

	// Fetch Environment Variables
	environment = structs.Environment{}
	envSet, err := env.UnmarshalFromEnviron(&environment)
	shared.Check(logger, "Failed to Fetch Environment Variables", err)
	environment.Extras = envSet

	// init testdb
	testRdsDb, err := structs.CreateRdsDbClientForLocalTest(ctx)
	shared.Check(logger, "Failed to create test container for Postgres DB", err)
	rdsDbQueries = rdsDbLib.New(testRdsDb)
	readRdsDbQueries = rdsDbLib.NewReadOnlyDBForTest(testRdsDb)

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
	shared.Check(logger, "Failed to initialize Short ID generator", err)
	// Set up shared rdsDbQueries.
	sharedRdsDbQueries = sharedRdsDb.New(sid, rdsDbQueries, readRdsDbQueries)

	// // Initiate AWS SDK Clients.
	// awsSdkClients, err = awsLib.CreateAwsSdkClientsForTesting(ctx, logger, environment.AwsAuthMethod)
	// shared.Check(logger, "Failed to initialize AwsSdkClients", err)

	temporalWorker, err := workflowTemporal.InitiateTemporalWorker(temporalClient)
	shared.Check(logger, "Failed to initiate Temporal Worker for workflow", err)
	defer temporalWorker.Stop()

	workflowService = api.NewWorkflowService(
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
		logger.Log("msg", "workflow service failed to start fiber", "error", err)
		os.Exit(1)
	}
	if err != nil {
		panic(err)
	}
	defer workflowService.Close()
	// Register route bounds to HTTP methods with API handlers.
	testFiberLambda = workflowService.GetTestFiberAdapter()
	// prometheus collector has been registered in workflow service before.
	// in test mode create a new registry to avoid duplicate error.
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	// Run all tests within the package.
	exitVal := m.Run()
	os.Exit(exitVal)
}

func GetAPIGatewayProxyRequest() (events.APIGatewayProxyRequest, error) {
	request := events.APIGatewayProxyRequest{}
	testRequestFile, err := os.ReadFile("./test_files/api-gateway-proxy-request-empty.json")
	if err != nil {
		return request, err
	}
	err = json.Unmarshal(testRequestFile, &request)
	return request, err
}

func GetCreateWorkflowRequest(orgId string, filePath string) (*structs.WorkflowEntity, error) {
	createWorkflowRequest := &structs.WorkflowEntity{}
	testFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(testFile, createWorkflowRequest)
	if err != nil {
		return nil, err
	}
	createWorkflowRequest.SugerOrgId = orgId
	return createWorkflowRequest, nil
}
