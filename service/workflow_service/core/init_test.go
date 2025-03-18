package core_test

// Command to run all tests under this package
// go test -v service/workflow_service/core/*_test.go

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Netflix/go-env"

	fiberAdapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/go-kit/log"
	"go.temporal.io/sdk/client"
	temporalWorkflow "go.temporal.io/sdk/workflow"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	workflowTemporal "github.com/sugerio/workflow-service-trial/service/workflow_service/temporal"
	"github.com/sugerio/workflow-service-trial/shared"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
)

var (
	logger           log.Logger
	workflowService  *api.WorkflowService
	environment      structs.Environment
	rdsDbQueries     *rdsDbLib.Queries
	readRdsDbQueries *rdsDbLib.Queries
	awsSdkClients    *awsLib.AwsSdkClients
	temporalClient   client.Client
	testFiberLambda  *fiberAdapter.FiberLambda
)

// There can be only ONE TestMain for each package.
// It has only one function named Run(), which runs all the tests within the package.
func TestMain(m *testing.M) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// Set up Logger.
	logger = log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "app", "workflow-service-core-test")

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
