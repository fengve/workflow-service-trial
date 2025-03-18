package core

import (
	"go.temporal.io/sdk/client"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	sharedRdsDb "github.com/sugerio/workflow-service-trial/shared/rds_db"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

var (
	environment   *structs.Environment
	awsSdkClients *awsLib.AwsSdkClients
	// Read/Write RDS DB connection
	rdsDbQueries *rdsDbLib.Queries
	// Read only RDS DB connection
	readRdsDbQueries   *rdsDbLib.Queries
	sharedRdsDbQueries *sharedRdsDb.Queries
	temporalClient     client.Client
)

func SetupGlobals(
	_environment *structs.Environment,
	_awsSdkClients *awsLib.AwsSdkClients,
	_rdsDbQueries *rdsDbLib.Queries,
	_readRdsDbQueries *rdsDbLib.Queries,
	_temporalClient client.Client) {
	environment = _environment
	awsSdkClients = _awsSdkClients
	rdsDbQueries = _rdsDbQueries
	readRdsDbQueries = _readRdsDbQueries
	temporalClient = _temporalClient
	sharedRdsDbQueries = sharedRdsDb.New(nil, _rdsDbQueries, _readRdsDbQueries)
}

func GetEnvironment() *structs.Environment {
	return environment
}

func GetAwsSdkClients() *awsLib.AwsSdkClients {
	return awsSdkClients
}

func GetRdsDbQueries() *rdsDbLib.Queries {
	return rdsDbQueries
}

func GetReadRdsDbQueries() *rdsDbLib.Queries {
	return readRdsDbQueries
}

func GetSharedRdsDbQueries() *sharedRdsDb.Queries {
	return sharedRdsDbQueries
}

func GetTemporalClient() client.Client {
	return temporalClient
}
