package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Netflix/go-env"
	//"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/go-kit/log"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	workflow_service "github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	workflowTemporal "github.com/sugerio/workflow-service-trial/service/workflow_service/temporal"
	"github.com/sugerio/workflow-service-trial/shared"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

const (
	// Time in milliseconds to wait before gracefully shutting down server.
	GRACEFULLY_SHUTDOWN_SERVER_MS = 20
)

func main() {
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()
	logger := log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "app", "workflow-service")
	// set environments for local dev
	if shared.IsLocalDevEnv() {
		structs.SetupLocalEnvironmentVariables()
		defer structs.CleanupLocalEnvironmentVariables()
	}
	var environment structs.Environment
	envSet, err := env.UnmarshalFromEnviron(&environment)
	shared.Check(logger, "Failed to UnmarshalFromEnviron", err)
	// Save remaining environment variables in Environment.Extras
	environment.Extras = envSet
	logger = log.With(logger, "env", environment.Env)
	var awsSdkClients *awsLib.AwsSdkClients
	if shared.IsLocalDevEnv() {
		//awsSdkClients, err = awsLib.NewAwsSdkClients_SSO(ctx, logger, structs.AWS_PROFILE_TEST)
		shared.Check(logger, "Failed to initialize AwsSdkClients", err)
	} else {
		// Initiate AWS SDK Clients from IRSA
		//awsSdkClients, err = awsLib.NewAwsSdkClients_IRSA(ctx, logger)
		shared.Check(logger, "Failed to initialize AwsSdkClients", err)
	}

	fmt.Println(environment)
	fmt.Println("--------")
	// Get the RDS DB password via SecretsManager.
	//secretValue, err := awsSdkClients.GetSecretsManagerClient().GetSecretValue(
	//	ctx,
	//	&secretsmanager.GetSecretValueInput{
	///		SecretId: aws.String(environment.RdsDb.PasswordSecretId),
	//})
	//shared.Check(logger, "Failed to get rds db secret", err)
	//environment.RdsDb.Password = *secretValue.SecretString
	writeDB, readDB, err := structs.ConnectRdsDbWithRead(&environment)
	shared.Check(logger, "Failed to connect to RDS DB", err)
	// Set the maximum number of open connections to the database.
	writeDB.SetMaxOpenConns(50)
	readDB.SetMaxOpenConns(50)
	defer writeDB.Close()
	defer readDB.Close()
	rdsDbQueries := rdsDbLib.New(writeDB)
	readRdsDbQueries := rdsDbLib.New(readDB)
	// Set up the Temporal Client.
	temporalClient, err := client.Dial(
		client.Options{
			HostPort: environment.Temporal.HostPort,
			ContextPropagators: []workflow.ContextPropagator{
				sharedTemporal.NewCommonContextPropagator(),
			},
			Namespace: "default",
		},
	)
	shared.Check(logger, "Failed to initialize Temporal Client", err)
	defer temporalClient.Close()

	temporalWorker, err := workflowTemporal.InitiateTemporalWorker(temporalClient)
	shared.Check(logger, "Failed to initiate Temporal Worker for workflow", err)
	defer temporalWorker.Stop()
	workflowService := workflow_service.NewWorkflowService(
		ctx,
		logger,
		&environment,
		awsSdkClients,
		rdsDbQueries,
		readRdsDbQueries,
		temporalClient,
	)
	// Start Workflow Service.
	go func() {
		err = workflowService.Start()
		if err != nil {
			logger.Log("msg", "workflow service failed to start fiber", "error", err)
			os.Exit(1)
		}
	}()
	// Handle Linux signals to gracefully shutdown server.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exit := func() {
		cancelCtx()
		// Give some time to goroutines for processing inflight events before forcefully shutdown.
		time.Sleep(time.Duration(GRACEFULLY_SHUTDOWN_SERVER_MS) * time.Millisecond)
		logger.Log("msg", "Forcefully shutdown server after grace period expires")
		workflowService.Close()
	}
	select {
	case signalExit := <-done:
		err := fmt.Errorf("%s", signalExit)
		logger.Log("error", fmt.Sprintf("Gracefully shutdown server since of system signals %v", err))
		exit()
	}
	logger.Log("info", "Graceful Exit Successfully!")
}
