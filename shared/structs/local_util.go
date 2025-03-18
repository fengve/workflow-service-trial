package structs

import (
	"fmt"
	"os"

	"github.com/sugerio/workflow-service-trial/shared"

	"go.temporal.io/sdk/client"
)

const (
	LOCAL_POSTGRES_DB_NAME       = "postgres"
	LOCAL_POSTGRES_USERNAME      = "rds_db_admin"
	LOCAL_POSTGRES_PASSWORD      = "password"
	LOCAL_POSTGRES_PORT          = "5432"
	LOCAL_POSTGRES_DB_URL_FORMAT = "postgres://%s:%s@localhost:%s/%s?sslmode=disable"

	WORKFLOW_SERVICE_PORT = ":5679"
)

func SetupLocalEnvironmentVariables() {
	currentEnv, _ := os.LookupEnv("ENV")
	fmt.Printf("Current ENV is %s \n", currentEnv)
	setenvIfAbsent("AWS_AUTH_METHOD", shared.AWS_AUTH_METHOD_SSO)
	setenvIfAbsent("WORKFLOW_SERVICE_PORT", WORKFLOW_SERVICE_PORT)
	setenvIfAbsent("WORKFLOW_SERVICE_TYPE", "MAIN")
	setenvIfAbsent("AWS_REGION", "us-west-2")
	setenvIfAbsent("RDS_DB_ENDPOINT", "localhost")
	setenvIfAbsent("RDS_DB_PORT", LOCAL_POSTGRES_PORT)
	setenvIfAbsent("RDS_DB_NAME", LOCAL_POSTGRES_DB_NAME)
	setenvIfAbsent("RDS_DB_USER", LOCAL_POSTGRES_USERNAME)
	setenvIfAbsent("RDS_DB_PASSWORD", LOCAL_POSTGRES_PASSWORD)
	setenvIfAbsent("RDS_DB_PASSWORD_SECRET_ID", "rds-private-postgres-db-dev-password")
	setenvIfAbsent("PUBLIC_BUCKET_NAME", "suger-public-access-bucket")
	setenvIfAbsent("BUCKET_KEY_AWS_MARKETPLACE_STACK_TEMPLATE", "suger-aws-marketplace-stack-template.yaml")
	setenvIfAbsent("BUCKET_KEY_AWS_STACK_LAMBDA", "suger-aws-stack-lambda.zip")
	setenvIfAbsent("TEMPORAL_HOST_PORT", client.DefaultHostPort)
	setenvIfAbsent("CORS_ALLOW_ORIGINS", "http://localhost:3000")
}

func setenvIfAbsent(key string, value string) {
	_, ok := os.LookupEnv(key)
	if ok {
		return
	}
	err := os.Setenv(key, value)
	if err != nil {
		fmt.Printf("Setenv(%s, %s) failed: %s", key, value, err)
	}
}

func CleanupLocalEnvironmentVariables() {
	os.Unsetenv("ENV")
}
