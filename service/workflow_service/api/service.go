package api

import (
	"context"
	"fmt"
	"os"

	fiberAdapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/go-kit/log"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.temporal.io/sdk/client"

	mw "github.com/sugerio/workflow-service-trial/middleware"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes"
	workflowTemporal "github.com/sugerio/workflow-service-trial/service/workflow_service/temporal"
	"github.com/sugerio/workflow-service-trial/shared"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
)

const (
	WORKFLOW_SERVICE_TYPE      = "WORKFLOW_SERVICE_TYPE"
	WORKFLOW_SERVICE_TYPE_MAIN = "MAIN"
)

type (
	// WorkflowService is the service that provides the ability to create, update, delete, and retrieve workflows.
	WorkflowService struct {
		Ctx              context.Context
		Logger           log.Logger
		environment      *structs.Environment
		awsSdkClients    *awsLib.AwsSdkClients
		rdsDbQueries     *rdsDbLib.Queries
		readRdsDbQueries *rdsDbLib.Queries
		fiberApp         *fiber.App
		temporalClient   client.Client
	}
)

// NewWorkflowService creates a new WorkflowService.
func NewWorkflowService(
	ctx context.Context,
	logger log.Logger,
	environment *structs.Environment,
	awsSdkClients *awsLib.AwsSdkClients,
	rdsDbQueries *rdsDbLib.Queries,
	readRdsDbQueries *rdsDbLib.Queries,
	temporalClient client.Client) *WorkflowService {
	service := &WorkflowService{
		Ctx:              ctx,
		Logger:           logger,
		environment:      environment,
		awsSdkClients:    awsSdkClients,
		rdsDbQueries:     rdsDbQueries,
		readRdsDbQueries: readRdsDbQueries,
		temporalClient:   temporalClient,
	}

	return service
}

func (service *WorkflowService) Start() error {

	fmt.Printf("Current env: %s", service.environment.Env)

	core.SetupGlobals(
		service.environment,
		service.awsSdkClients,
		service.rdsDbQueries,
		service.readRdsDbQueries,
		service.temporalClient)

	// Set up globals for the temporal workflows and activities.
	sharedTemporal.SetupGlobals(
		service.environment,
		service.awsSdkClients,
		service.rdsDbQueries,
		service.readRdsDbQueries)
	// Only the main service should set up the webhooks and temporal workflows for schedule triggers.
	if ifIsMain() {
		// Unregister all the webhooks for active webhook workflows. Ignore errors.
		err := core.UnregisterAllWebhooks(service.Ctx)
		if err != nil {
			service.Logger.Log("unregister all webhooks failed when start", err)
		}
		// Register all the webhooks for active webhook workflows.
		err = core.RegisterAllWebhooks(service.Ctx)
		if err != nil {
			return err
		}

		// Clean up the temporal workflows for active schedule trigger workflows. Ignore errors.
		err = workflowTemporal.TerminateAllTemporalWorkflows_ScheduleTrigger(service.Ctx)
		if err != nil {
			service.Logger.Log("terminate all temporal workflows for schedule trigger failed when start", err)
		}
		// Set up the temporal workflows for active schedule trigger workflows.
		err = workflowTemporal.SetupAllTemporalWorkflows_ScheduleTrigger(service.Ctx)
		if err != nil {
			return err
		}
	}

	// Set up fiber app.
	service.fiberApp = fiber.New(
		fiber.Config{
			BodyLimit: 50 * 1024 * 1024, // max 50MB request body to process
		},
	)

	service.fiberApp.Use(otelfiber.Middleware())
	// Use panic recover middleware, must be the first middleware
	service.fiberApp.Use(mw.NewPanicRecover())
	// Apply CORS middleware for adding CORS headers in response.
	service.fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     service.environment.AllowOrigins,
		AllowHeaders:     "Origin, Authorization, Content-Type, X-Amz-Date, X-Amz-Security-Token, X-Api-Key, X-Suger-Entity-Type, X-Suger-Entity-Id, X-Suger-Email, sessionid",
		AllowCredentials: true,
	}))
	service.fiberApp.Use(mw.OrgIdMiddleware)
	service.fiberApp.Use(mw.LoggerMiddleware)
	// Register routes after the above middleware
	service.RegisterAllRouteMethods()

	// If it is a test, do not start the rest service listener.
	if shared.IsTestEnv() || shared.IsLocalTestEnv() {
		fmt.Printf("Test env, skip starting rest service listener")
		// add panic entry point for testing
		service.fiberApp.Get("/panic", func(c *fiber.Ctx) error {
			panic("Panic for testing")
		})
		return nil
	}

	// start rest service via fiber app.
	return service.fiberApp.Listen(service.environment.ServicePort.WorkflowService)
}

func (service *WorkflowService) Close() {
	// Shutdown all active executions.
	core.ShutdownActiveExecutions()
	// shutdown rest service via fiber app.
	service.fiberApp.Shutdown()
}

func (service *WorkflowService) RegisterAllRouteMethods() {
	service.RegisterRouteMethods_Execution()
	service.RegisterRouteMethods_Node()
	service.RegisterRouteMethods_Webhook()
	service.RegisterRouteMethods_Workflow()
	service.RegisterRouteMethods_DynamicParameter()
}

func (service *WorkflowService) GetTestFiberAdapter() *fiberAdapter.FiberLambda {
	return fiberAdapter.New(service.fiberApp)
}

func GetContextUserId(ctx *fiber.Ctx) string {
	return "anonymous"
}

// Check if the current workflow service is the main service.
func ifIsMain() bool {
	serviceType := os.Getenv(WORKFLOW_SERVICE_TYPE)
	if serviceType == WORKFLOW_SERVICE_TYPE_MAIN {
		return true
	}
	return false
}
