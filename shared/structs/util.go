package structs

import (
	"database/sql"
	"encoding/json"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"reflect"
)

func IntegrationInfoFromJson(data []byte) (IntegrationInfo, error) {
	result := IntegrationInfo{}
	err := json.Unmarshal(data, &result)
	return result, err
}

// CopyCommonFields copies src fields into dest fields. A src field is copied
// to a dest field if they have the same field name & field type.
// Dest and src must be pointers to structs.
func CopyCommonFields(src, dest interface{}) {
	srcType := reflect.TypeOf(src).Elem()
	destType := reflect.TypeOf(dest).Elem()
	destFieldsMap := map[string]int{}

	for i := 0; i < destType.NumField(); i++ {
		destFieldsMap[destType.Field(i).Name] = i
	}

	for i := 0; i < srcType.NumField(); i++ {
		if j, ok := destFieldsMap[srcType.Field(i).Name]; ok {
			if srcType.Field(i).Type == destType.Field(j).Type {
				reflect.ValueOf(dest).Elem().Field(j).Set(
					reflect.ValueOf(src).Elem().Field(i),
				)
			}
		}
	}
}

// Connect the RDS database with two connections (one for read+write, one for read only)
// using the Environment info and return the (writeDB, readDB, error/nil).
// If the read endpoint is not set, the readDB will be the same as the writeDB.
func ConnectRdsDbWithRead(environment *Environment) (*sql.DB, *sql.DB, error) {
	rdsDbConfig := rdsDbLib.SqlDbConfig{
		HostName: environment.RdsDb.Endpoint,
		HostPort: environment.RdsDb.Port,
		UserName: environment.RdsDb.User,
		Password: environment.RdsDb.Password,
		DbName:   environment.RdsDb.Name,
	}
	writeDB, err := rdsDbLib.ConnectSqlDb("postgres", rdsDbConfig)
	if err != nil {
		return nil, nil, err
	}
	// If the read endpoint is not set, return the writeDB as the readDB.
	if environment.RdsDb.ReadEndpoint == "" {
		return writeDB, writeDB, nil
	}
	// Use the read endpoint as the host name for the readDB, and connect to the readDB.
	rdsDbConfig.HostName = environment.RdsDb.ReadEndpoint
	readDB, err := rdsDbLib.ConnectSqlDb("postgres", rdsDbConfig)
	if err != nil {
		return nil, nil, err
	}
	return writeDB, readDB, nil
}
