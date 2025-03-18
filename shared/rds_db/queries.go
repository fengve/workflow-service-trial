package rds_db

import (
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/teris-io/shortid"
)

func New(
	sid *shortid.Shortid,
	rdsDbQueries *rdsDbLib.Queries,
	readRdsDbQueries *rdsDbLib.Queries) *Queries {
	if sid == nil {
		sid, _ = shortid.New(3, shortid.DefaultABC, 3456)
	}

	// Use the same rdsDbQueries for readRdsDbQueries if not provided.
	if readRdsDbQueries == nil {
		readRdsDbQueries = rdsDbQueries
	}

	return &Queries{
		sid:              sid,
		rdsDbQueries:     rdsDbQueries,
		readRdsDbQueries: readRdsDbQueries,
	}
}

type Queries struct {
	sid *shortid.Shortid
	// Read/Write RDS DB connection.
	rdsDbQueries *rdsDbLib.Queries
	// Read only RDS DB connection.
	readRdsDbQueries *rdsDbLib.Queries
}
