package temporal

import (
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	sharedRdsDb "github.com/sugerio/workflow-service-trial/shared/rds_db"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

var (
	environment        *structs.Environment
	awsSdkClients      *awsLib.AwsSdkClients
	rdsDbQueries       *rdsDbLib.Queries
	readRdsDbQueries   *rdsDbLib.Queries
	sharedRdsDbQueries *sharedRdsDb.Queries
)

// Set the global variables and clients for the temporal workflow & activity.
func SetupGlobals(
	_environment *structs.Environment,
	_awsSdkClients *awsLib.AwsSdkClients,
	_rdsDbQueries *rdsDbLib.Queries,
	_readRdsDbQueries *rdsDbLib.Queries, ) {
	environment = _environment
	awsSdkClients = _awsSdkClients
	rdsDbQueries = _rdsDbQueries
	readRdsDbQueries = _readRdsDbQueries
	// Initialize the sharedRdsDbQueries if the rdsDbQueries is provided.
	// The readRdsDbQueries will be set to rdsDbQueries if not provided.
	if rdsDbQueries != nil {
		sharedRdsDbQueries = sharedRdsDb.New(
			nil, rdsDbQueries, readRdsDbQueries)
	}
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
