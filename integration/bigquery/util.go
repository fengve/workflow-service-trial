package bigquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/bigquery"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"google.golang.org/api/option"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	awsLib "github.com/sugerio/workflow-service-trial/shared/aws_lib"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	Template_BigqueryIntegrationSecretKey = "org_%s_integration_GCP_BIGQUERY_secretkey"
)

type BigqueryProject struct {
	Kind         string `json:"kind,omitempty"`
	ID           string `json:"id,omitempty"`
	NumericId    string `json:"numericId,omitempty"`
	FriendlyName string `json:"friendlyName,omitempty"`
}

type BigqueryProjectListResponse struct {
	Kind       string            `json:"kind,omitempty"`
	Etag       string            `json:"etag,omitempty"`
	Projects   []BigqueryProject `json:"projects,omitempty"`
	TotalItems int32             `json:"totalItems,omitempty"`
}

// Creates a bigquery integration for testing.
func CreateBigqueryIntegration_Testing(
	orgId string,
	rdsDbQueries *rdsDbLib.Queries,
	awsSdkClients *awsLib.AwsSdkClients) (*structs.GcpIntegration, error) {
	secretKey := "suger-dev-gcp-bigquery-private-key-json"
	secretValue, err := awsSdkClients.GetSecretsManagerClient().GetSecretValue(
		context.Background(),
		&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretKey),
		})

	if err != nil {
		return nil, err
	}

	gcpAuthPrivateKey := structs.GcpAuthPrivateKey{}
	err = json.Unmarshal([]byte(*secretValue.SecretString), &gcpAuthPrivateKey)
	if err != nil {
		return nil, err
	}

	bigqueryIntegration := &structs.GcpIntegration{
		// TODO: Replace here before merge
		GcpProjectId:                 gcpAuthPrivateKey.ProjectId,
		ServiceAccountEmail:          gcpAuthPrivateKey.ClientEmail,
		ServiceAccountPrivateKeyJson: *secretValue.SecretString,
		SecretKey:                    secretKey,
	}
	integrationInfo := structs.IntegrationInfo{
		GcpIntegration: bigqueryIntegration,
	}
	integrationInfoJson, err := json.Marshal(integrationInfo)
	if err != nil {
		return nil, err
	}
	_, err = rdsDbQueries.CreateIntegration(
		context.Background(),
		rdsDbLib.CreateIntegrationParams{
			OrganizationID: orgId,
			Partner:        string(structs.Partner_GCP),
			Service:        string(structs.PartnerService_BIGQUERY),
			Status:         string(structs.IntegrationStatus_VERIFIED),
			Info:           integrationInfoJson,
		})
	if err != nil {
		return nil, err
	}

	return bigqueryIntegration, nil
}

func GetBigqueryIntegration(
	orgId string,
	rdsDbQueries *rdsDbLib.Queries,
	awsSdkClients *awsLib.AwsSdkClients) (*structs.GcpIntegration, error) {
	integration_RdsDbLib, err := rdsDbQueries.GetIntegration(
		context.Background(),
		rdsDbLib.GetIntegrationParams{
			OrganizationID: orgId,
			Partner:        string(structs.Partner_GCP),
			Service:        string(structs.PartnerService_BIGQUERY),
		})
	if err != nil {
		return nil, err
	}
	integration, err := structs.FromIdentityIntegration(&integration_RdsDbLib)
	if err != nil {
		return nil, err
	}
	biqueryIntegration := integration.Info.GcpIntegration
	if biqueryIntegration == nil {
		return nil, fmt.Errorf("null integration of orgId %s", orgId)
	}
	if biqueryIntegration.ServiceAccountPrivateKeyJson == "" {
		err := GetBigqueryIntegrationServiceAccountPrivateKeyJson(
			awsSdkClients, orgId, biqueryIntegration)
		if err != nil {
			return nil, err
		}
	}
	return biqueryIntegration, nil
}

func NewBigqueryClient(integration *structs.GcpIntegration, projectId string) (*bigquery.Client, error) {
	if integration == nil || integration.ServiceAccountPrivateKeyJson == "" {
		return nil, fmt.Errorf("integration is invalid")
	}
	client, err := bigquery.NewClient(
		context.Background(), projectId, option.WithCredentialsJSON([]byte(integration.ServiceAccountPrivateKeyJson)))
	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}
	return client, nil
}

// Validate BigqueryIntegration for creation.
func ValidateBigqueryIntegrationForCreation(integration *structs.GcpIntegration) error {
	if integration == nil {
		return errors.New("the input bigqueryIntegration is nil")
	}
	if integration.ServiceAccountPrivateKeyJson == "" {
		return errors.New("the input ServiceAccountPrivateKeyJson is empty")
	}

	gcpAuthPrivateKey := structs.GcpAuthPrivateKey{}
	err := json.Unmarshal([]byte(integration.ServiceAccountPrivateKeyJson), &gcpAuthPrivateKey)
	if err != nil {
		return err
	}
	if gcpAuthPrivateKey.ProjectId == "" {
		return errors.New("the input ServiceAccountPrivateKeyJson does not contain ProjectId")
	}
	integration.GcpProjectId = gcpAuthPrivateKey.ProjectId
	if gcpAuthPrivateKey.ClientEmail == "" {
		return errors.New("the input ServiceAccountPrivateKeyJson does not contain ClientEmail")
	}
	integration.ServiceAccountEmail = gcpAuthPrivateKey.ClientEmail

	return nil
}

// Get the key to store/query Bigquery Integration service account private key to the AWS Secrets Manager.
func GetBigqueryIntegrationSecretKey(integration *structs.IdentityIntegration) string {
	return GenerateBigqueryIntegrationSecretKey(integration.OrganizationID)
}

// Generate a key to store/query Bigquery Integration service account private key to the AWS Secrets Manager.
func GenerateBigqueryIntegrationSecretKey(orgId string) string {
	return fmt.Sprintf(Template_BigqueryIntegrationSecretKey, orgId)
}

// Update the service account private key if it is available in AWS Secret Manager. Otherwise create a new one to store.
func StoreOrUpdateBigqueryIntegrationPrivateKey(
	awsSdkClients *awsLib.AwsSdkClients,
	orgId string,
	integration *structs.GcpIntegration) error {
	if integration == nil || integration.ServiceAccountPrivateKeyJson == "" {
		return errors.New("the input bigqueryIntegration or bigqueryIntegration.ServiceAccountPrivateKeyJson is not available")
	}

	// Generate secret key(for internal use) if empty.
	secretKey := integration.SecretKey
	if secretKey == "" {
		secretKey = GenerateBigqueryIntegrationSecretKey(orgId)
		integration.SecretKey = secretKey
	}

	// Store/update private key in AWS Secrets Manager.
	err := awsSdkClients.CreateOrUpdateSecretInSecretManager(
		orgId,
		secretKey,
		integration.ServiceAccountPrivateKeyJson,
		"GCP Bigquery Integration Service Account Private Key")
	if err != nil {
		return err
	}

	// Clear private key from the input integration for safety.
	integration.ServiceAccountPrivateKeyJson = ""
	return nil
}

// Get BigqueryIntegration.ServiceAccountPrivateKeyJson from the AWS Secrets Manager.
// Sets field bigqueryIntegration.ServiceAccountPrivateKey in the input.
// Returns error if not found.
func GetBigqueryIntegrationServiceAccountPrivateKeyJson(
	awsSdkClients *awsLib.AwsSdkClients,
	orgId string,
	integration *structs.GcpIntegration) error {
	secretKey := integration.SecretKey
	if secretKey == "" {
		secretKey = GenerateBigqueryIntegrationSecretKey(orgId)
		integration.SecretKey = secretKey
	}

	serviceAccountPrivateKeyJson, err := awsSdkClients.GetSecretFromSecretManager(orgId, secretKey)
	if err != nil {
		return err
	}
	if serviceAccountPrivateKeyJson == nil {
		return errors.New("failed to GetBigqueryIntegrationServiceAccountPrivateKey. the private key is not found")
	}

	integration.ServiceAccountPrivateKeyJson = *serviceAccountPrivateKeyJson
	return nil
}

// Delete BigqueryIntegration.ServiceAccountPrivateKey from the AWS Secrets Manager.
// Returns error if failed to delete.
func DeleteBigqueryIntegrationServiceAccountPrivateKey(
	awsSdkClients *awsLib.AwsSdkClients,
	orgId string,
	integration *structs.GcpIntegration) error {
	secretKey := integration.SecretKey
	if secretKey == "" {
		secretKey = GenerateBigqueryIntegrationSecretKey(orgId)
		integration.SecretKey = secretKey
	}

	err := awsSdkClients.DeleteSecretFromSecretManager(orgId, secretKey)
	if err != nil {
		return err
	}
	return nil
}

// Verify the BigqueryIntegration by calling the Bigquery API.
// Returns error if invalid, otherwise nil.
func VerifyBigqueryIntegration(
	ctx context.Context,
	awsSdkClients *awsLib.AwsSdkClients,
	orgId string,
	integration *structs.GcpIntegration,
) error {
	if integration == nil {
		return errors.New("the input bigqueryIntegration is nil")
	}
	if integration.ServiceAccountPrivateKeyJson == "" {
		// Get the service account private key json from the AWS Secrets Manager.
		err := GetBigqueryIntegrationServiceAccountPrivateKeyJson(
			awsSdkClients, orgId, integration)
		if err != nil {
			return err
		}
	}

	// Create a Bigquery client with the service account private key json.
	client, err := bigquery.NewClient(
		ctx, integration.GcpProjectId, option.WithCredentialsJSON([]byte(integration.ServiceAccountPrivateKeyJson)))
	if err != nil {
		return err
	}

	dataset, err := client.Datasets(ctx).Next()
	if err != nil {
		return err
	}

	if dataset == nil {
		return errors.New("failed to VerifyBigqueryIntegration. the dataset is not found")
	}
	if dataset.ProjectID == integration.GcpProjectId {
		return nil
	} else {
		return errors.New("failed to VerifyBigqueryIntegration. the project id is not matched")
	}
}

// List GoogleBigquery Projects Using REST API
func ListProjects(integration *structs.GcpIntegration) (*BigqueryProjectListResponse, error) {
	if integration == nil || integration.ServiceAccountPrivateKeyJson == "" {
		return nil, fmt.Errorf("integration is invalid")
	}
	serviceAccountPrivateKey := structs.GcpAuthPrivateKey{}
	err := json.Unmarshal([]byte(integration.ServiceAccountPrivateKeyJson), &serviceAccountPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed parse the private key: %w", err)
	}
	accessToken, err := GetAccessToken(
		integration.ServiceAccountEmail, serviceAccountPrivateKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessToken: %w", err)
	}

	// Call Rest API
	httpRequest, err := http.NewRequest(
		"GET", "https://bigquery.googleapis.com/bigquery/v2/projects", nil)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *accessToken))
	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	// Check the response status code.
	if httpResponse.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get call bigquery rest api, status code: %d", httpResponse.StatusCode)
	}
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	result := &BigqueryProjectListResponse{}
	err = json.Unmarshal(responseBody, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
