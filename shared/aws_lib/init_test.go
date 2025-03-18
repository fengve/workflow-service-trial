package aws_lib_test

// Command to run all tests under this package
// go test -v shared/aws_lib/*_test.go

import (
	"context"
	"os"
	"testing"

	"github.com/Netflix/go-env"

	"github.com/go-kit/log"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/shared"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	sharedRdsDb "github.com/sugerio/workflow-service-trial/shared/rds_db"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
	"github.com/teris-io/shortid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

var (
	logger             log.Logger
	sid                *shortid.Shortid
	temporalClient     client.Client
	environment        structs.Environment
	rdsDbQueries       *rdsDbLib.Queries
	readRdsDbQueries   *rdsDbLib.Queries
	sharedRdsDbQueries *sharedRdsDb.Queries
	awsSdkClients      *awsLib.AwsSdkClients
)

// There can be only ONE TestMain for each package. It has a single defined function named Run(), which runs all the tests within the package.
func TestMain(m *testing.M) {
	exitVal := RunTestMain(m)
	os.Exit(exitVal)
}

func RunTestMain(m *testing.M) int {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// Set up Logger.
	logger = log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "app", "aws_lib-test")

	structs.SetupEnvironmentVariables()
	defer structs.CleanupEnvironmentVariables()

	// Fetch Environment Variables
	environment = structs.Environment{}
	envSet, err := env.UnmarshalFromEnviron(&environment)
	shared.Check(logger, "Failed to Fetch Environment Variables", err)
	environment.Extras = envSet

	testRdsDb, err := structs.CreateRdsDbClientForLocalTest(ctx)
	shared.Check(logger, "Failed to create test container for Postgres DB", err)
	rdsDbQueries = rdsDbLib.New(testRdsDb)
	readRdsDbQueries = rdsDbLib.New(testRdsDb)

	temporalClient, err = client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
		ContextPropagators: []workflow.ContextPropagator{
			sharedTemporal.NewCommonContextPropagator(),
		},
		Namespace: "default",
	})
	shared.Check(logger, "Failed to connect temporal client", err)
	defer temporalClient.Close()

	// Set up short ID generator.
	sid, err = shortid.New(2, shortid.DefaultABC, 2342)
	shared.Check(logger, "Failed to initialize Short ID generator", err)
	// Set up shared rdsDbQueries.
	sharedRdsDbQueries = sharedRdsDb.New(sid, rdsDbQueries, readRdsDbQueries)

	// Initiate AWS SDK Clients.
	awsSdkClients, err = awsLib.CreateAwsSdkClientsForTesting(ctx, logger, environment.AwsAuthMethod)
	shared.Check(logger, "Failed to initialize AwsSdkClients", err)

	// Run all tests within the package.
	exitVal := m.Run()
	return exitVal
}
