package structs

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/shared"
	"github.com/teris-io/shortid"
)

const (
	TEST_POSTGRES_DB_NAME       = "postgres"
	TEST_POSTGRES_USERNAME      = "rds_db_admin"
	TEST_POSTGRES_PASSWORD      = "password"
	TEST_POSTGRES_PORT          = "5432"
	TEST_POSTGRES_DB_URL_FORMAT = "postgres://%s:%s@localhost:%s/%s?sslmode=disable"

	AWS_PROFILE_TEST = "workload-dev" // Here we use the workload-dev as our unit testing profile.
)

// If the input orgId is empty, then create a new one.
func CreateOrganization_Testing(
	rdsDbQueries *rdsDbLib.Queries,
	sid *shortid.Shortid,
	orgId string) *rdsDbLib.IdentityOrganization {
	var err error
	// If the orgId is not empty, then try to get the organization.
	if orgId != "" {
		// Check if the organization exists.
		org, err := rdsDbQueries.GetOrganizationById(context.Background(), orgId)
		if err == nil {
			return &org
		}
		// If the organization does not exist, then create a new one.
		if err != sql.ErrNoRows {
			panic(err)
		}
	}

	if orgId == "" {
		orgId, err = sid.Generate()
		if err != nil {
			panic(err)
		}
	}

	authId, err := sid.Generate()
	if err != nil {
		panic(err)
	}
	user, err := rdsDbQueries.GetUserByEmail(context.Background(), "ruiqi@suger.io")
	if err != nil {
		panic(err)
	}
	createOrgParams := rdsDbLib.CreateOrganizationParams{
		ID:                 orgId,
		Name:               "test_suger_org",
		EmailDomain:        "suger.io",
		Website:            "www.suger.io",
		Description:        "test org",
		AllowedAuthMethods: []string{},
		CreatedBy:          user.ID,
		AuthID:             authId,
		Status:             string(OrganizationStatus_ACTIVE),
	}
	// Create a new organization.
	organization, err := rdsDbQueries.CreateOrganization(
		context.Background(), createOrgParams)
	if err != nil {
		panic(err)
	}

	// Set the user as the admin of the new org.
	_, err = rdsDbQueries.AddUserToOrganization(
		context.Background(),
		rdsDbLib.AddUserToOrganizationParams{
			UserID:             user.ID,
			OrganizationID:     organization.ID,
			UserRole:           "ADMIN",
			AllowedAuthMethods: []string{},
		},
	)
	if err != nil {
		panic(err)
	}

	return &organization
}

type TestContainerBaseDBOption struct {
	Env                  map[string]string
	DBURLFormat          string
	DBUsername           string
	DBPassword           string
	DBName               string
	DBPort               string
	ImageName            string
	ExposedPortEnvKey    string
	SchemaFilePathFormat string
}

func CreateRdsDbClientForLocalTest(ctx context.Context) (*sql.DB, error) {
	dbPort := os.Getenv("RDS_DB_PORT")
	dbUser := os.Getenv("RDS_DB_USER")
	dbPassword := os.Getenv("RDS_DB_PASSWORD")
	dbName := os.Getenv("RDS_DB_NAME")
	dbUrl := fmt.Sprintf(TEST_POSTGRES_DB_URL_FORMAT, dbUser, dbPassword, dbPort, dbName)
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}

	// Add two users for testing.
	dbQueries := rdsDbLib.New(db)
	_, err = dbQueries.CreateUser(ctx, rdsDbLib.CreateUserParams{
		ID:        "vQAUJlvfT",
		FirstName: "Ruiqi",
		LastName:  "Chen",
		Email:     "ruiqi@suger.io",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new user")
	}
	_, err = dbQueries.CreateUser(ctx, rdsDbLib.CreateUserParams{
		ID:        "PJrlnwU4T",
		FirstName: "Chengjun",
		LastName:  "Yuan",
		Email:     "test@suger.io",
	})
	if err != nil {
		return db, fmt.Errorf("failed to create new user")
	}
	_, err = dbQueries.CreateUser(ctx, rdsDbLib.CreateUserParams{
		ID:        "5GUsZRVzT",
		FirstName: "Jon",
		LastName:  "Yoo",
		Email:     "jon@suger.io",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new user")
	}

	return db, nil
}

func SetupEnvironmentVariables() {
	// Set the ENV to local-test by default if it is not set.
	os.Setenv("ENV", shared.ENV_LOCAL_TEST)
	// If the AWS_AUTH_METHOD is not set, then set it to SSO by default.
	_, isAwsAuthPresent := os.LookupEnv("AWS_AUTH_METHOD")
	if !isAwsAuthPresent {
		os.Setenv("AWS_AUTH_METHOD", shared.AWS_AUTH_METHOD_SSO)
	}

	os.Setenv("AWS_APP_CONFIG_ENDPOINT", "http://localhost:2772")
	os.Setenv("AWS_APP_CONFIG_APPLICATION", "marketplace")
	os.Setenv("AWS_APP_CONFIG_ENVIRONMENT", "default")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("ORG_FILE_BUCKET", "org-file-bucket-dev")
	os.Setenv("PUBLIC_BUCKET_NAME", "suger-public-access-bucket")
	os.Setenv("RDS_DB_ENDPOINT", "localhost")
	os.Setenv("RDS_DB_PORT", TEST_POSTGRES_PORT)
	os.Setenv("RDS_DB_NAME", TEST_POSTGRES_DB_NAME)
	os.Setenv("RDS_DB_USER", TEST_POSTGRES_USERNAME)
	os.Setenv("RDS_DB_PASSWORD", TEST_POSTGRES_PASSWORD)
	os.Setenv("RDS_DB_PASSWORD_SECRET_ID", "rds-private-postgres-db-dev-password")
}

func CleanupEnvironmentVariables() {
}
