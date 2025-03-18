package aws_lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/marketplaceagreement"
	"github.com/aws/aws-sdk-go-v2/service/marketplacecatalog"
	mcas "github.com/aws/aws-sdk-go-v2/service/marketplacecommerceanalytics"
	"github.com/aws/aws-sdk-go-v2/service/marketplaceentitlementservice"
	"github.com/aws/aws-sdk-go-v2/service/marketplacemetering"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/log"
	"github.com/sugerio/workflow-service-trial/shared"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	// EmptyStringSHA256 is the hex encoded sha256 value of an empty string
	EmptyStringSHA256 = `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
)

type AwsSdkClients struct {
	ctx                  context.Context
	logger               log.Logger
	roleArn              string
	webIdentityTokenFile string
	awsRegion            string
	watcher              *fsnotify.Watcher
	awsConfig            aws.Config
	callerIdentity       *sts.GetCallerIdentityOutput
	// awsConfigsFromAssumeRole_Cache *ristretto.Cache
	// AWS Clients

	s3Client             *s3.Client
	secretsManagerClient *secretsmanager.Client
	sesClient            *ses.Client
	snsClient            *sns.Client
	sqsClient            *sqs.Client
	stsClient            *sts.Client
	meteringClient       *marketplacemetering.Client
}

var customEndpointResolver = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
	switch service {
	case marketplacecatalog.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://catalog.marketplace.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case marketplaceentitlementservice.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://entitlement.marketplace.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case sns.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://sns.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case marketplacemetering.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://metering.marketplace.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case mcas.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://marketplacecommerceanalytics.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case cloudformation.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://cloudformation.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case eventbridge.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://events.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case marketplaceagreement.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://agreement-marketplace.us-east-1.amazonaws.com",
			SigningName:   "aws-marketplace",
			SigningRegion: "us-east-1",
		}, nil
	default:
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}
})

var customEndpointResolverForAssumeRole = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
	switch service {
	case marketplacecatalog.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://catalog.marketplace.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case marketplaceentitlementservice.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://entitlement.marketplace.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case sns.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://sns.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case marketplacemetering.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://metering.marketplace.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case s3.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://s3.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case mcas.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://marketplacecommerceanalytics.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case cloudformation.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://cloudformation.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case eventbridge.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://events.us-east-1.amazonaws.com",
			SigningRegion: "us-east-1",
		}, nil
	case marketplaceagreement.ServiceID:
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://agreement-marketplace.us-east-1.amazonaws.com",
			SigningName:   "aws-marketplace",
			SigningRegion: "us-east-1",
		}, nil
	default:
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}
})

// Create the AWS SDK clients from environment variables.
// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_REGION are required.
// It is used for Github tests.
func NewAwsSdkClients_ENV(ctx context.Context, logger log.Logger) (*AwsSdkClients, error) {
	awsRegion := shared.GetEnv("AWS_REGION", "")
	if awsRegion == "" {
		return nil, errors.New("the Env viriable AWS_REGION is empty")
	}
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	awsConfig.EndpointResolverWithOptions = customEndpointResolver

	awsSdkClients := &AwsSdkClients{
		ctx:       ctx,
		logger:    logger,
		awsRegion: awsRegion,
		awsConfig: awsConfig,
	}
	awsSdkClients.createClients()
	if err = awsSdkClients.createCaches(); err != nil {
		return nil, err
	}

	return awsSdkClients, nil
}

// Create the AWS SDK clients from SSO with the given profile.
// It is used for Local tests by getting the awsConfig from Local AWS SSO login.
func NewAwsSdkClients_SSO(
	ctx context.Context, logger log.Logger, profileName string) (*AwsSdkClients, error) {
	//awsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profileName))
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Log("guidance", fmt.Sprintf("Please login via command: aws sso login --profile %s", profileName))
		return nil, err
	}
	awsConfig.EndpointResolverWithOptions = customEndpointResolver

	awsSdkClients := &AwsSdkClients{
		ctx:       ctx,
		logger:    logger,
		awsRegion: awsConfig.Region,
		awsConfig: awsConfig,
	}
	awsSdkClients.createClients()
	if err = awsSdkClients.createCaches(); err != nil {
		return nil, err
	}

	return awsSdkClients, nil
}

// NewAwsSdkClients_LOCALSTACK Create an AWS client that uses a LocalStack container to mock AWS services.
func NewAwsSdkClients_LOCALSTACK(ctx context.Context, logger log.Logger) (*AwsSdkClients, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost.localstack.cloud:4566",
			SigningRegion: "us-east-1",
		}, nil
	})
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		return nil, err
	}
	awsSdkClients := &AwsSdkClients{
		ctx:       ctx,
		logger:    logger,
		awsRegion: awsConfig.Region,
		awsConfig: awsConfig,
	}
	awsSdkClients.createClients()
	// By default, the S3 client uses virtual hosted bucket addressing, but this is not well supported in LocalStack.
	// So have to UsePathStyle client
	awsSdkClients.createPathStyleS3ClientForLocalStack()
	if err = awsSdkClients.createCaches(); err != nil {
		return nil, err
	}
	return awsSdkClients, nil
}

// Create the AWS SDK clients from IRSA (IAM Role for Service Account).
// It is used on dev & prod services running inside EKS nodes.
func NewAwsSdkClients_IRSA(ctx context.Context, logger log.Logger) (*AwsSdkClients, error) {
	roleArn := shared.GetEnv("AWS_ROLE_ARN", "")
	if roleArn == "" {
		return nil, errors.New("the Env viriable AWS_ROLE_ARN is empty")
	}
	webIdentityTokenFile := shared.GetEnv("AWS_WEB_IDENTITY_TOKEN_FILE", "")
	if webIdentityTokenFile == "" {
		return nil, errors.New("the Env viriable AWS_WEB_IDENTITY_TOKEN_FILE is empty")
	}
	awsRegion := shared.GetEnv("AWS_REGION", "")
	if awsRegion == "" {
		return nil, errors.New("the Env viriable AWS_REGION is empty")
	}
	awsSdkClients := &AwsSdkClients{
		ctx:                  ctx,
		logger:               logger,
		roleArn:              roleArn,
		webIdentityTokenFile: webIdentityTokenFile,
		awsRegion:            awsRegion,
	}

	err := awsSdkClients.populateAwsConfig()
	if err != nil {
		return nil, err
	}
	awsSdkClients.createClients()
	if err = awsSdkClients.createCaches(); err != nil {
		return nil, err
	}

	// Set up the fsnotify watcher to watch the refresh of webIdentityTokenFile's content.
	// If any updates, reset the awsConfig.
	if awsSdkClients.watcher, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}
	if err = awsSdkClients.watcher.Add(webIdentityTokenFile); err != nil {
		return nil, err
	}
	go awsSdkClients.onWebIdentityTokenFileUpdate()

	return awsSdkClients, nil
}

func (asc *AwsSdkClients) GetS3Client() *s3.Client {
	return asc.s3Client
}

func (asc *AwsSdkClients) GetStsClient() *sts.Client {
	return asc.stsClient
}

func (asc *AwsSdkClients) GetSecretsManagerClient() *secretsmanager.Client {
	return asc.secretsManagerClient
}

func (asc *AwsSdkClients) GetSqsClient() *sqs.Client {
	return asc.sqsClient
}

func (asc *AwsSdkClients) GetSnsClient() *sns.Client {
	return asc.snsClient
}

func (asc *AwsSdkClients) GetSesClient() *ses.Client {
	return asc.sesClient
}

func (asc *AwsSdkClients) GetMeteringClient() *marketplacemetering.Client {
	return asc.meteringClient
}

func (asc *AwsSdkClients) GetCallerIdentity() (*sts.GetCallerIdentityOutput, error) {
	if asc.callerIdentity == nil {
		var err error
		if asc.callerIdentity, err = asc.stsClient.GetCallerIdentity(asc.ctx, nil); err != nil {
			return nil, err
		}
	}
	return asc.callerIdentity, nil
}

// Ge the signed GetCallerIdentity request using Signature Version 4.
func (asc *AwsSdkClients) GetSignedRequest(gcpTargetResource string) (*http.Request, error) {
	request, err := http.NewRequest("POST", "https://sts.amazonaws.com/?Action=GetCallerIdentity&Version=2011-06-15", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Host", "sts.amazonaws.com")
	request.Header.Set("x-goog-cloud-target-resource", gcpTargetResource)

	awsCredentials, err := asc.awsConfig.Credentials.Retrieve(asc.ctx)
	if err != nil {
		return nil, err
	}
	signer := v4.NewSigner()
	err = signer.SignHTTP(asc.ctx, awsCredentials, request, EmptyStringSHA256, "sts", "us-east-1", time.Now())

	return request, err
}

// Get the token of the signed GetCallerIdentity request for GCP Workload Identity Federation.
func (asc *AwsSdkClients) GetSignedRequestToken_GcpWorkloadIdentityFederation(gcpTargetResource string) (string, error) {
	signedRequest, err := asc.GetSignedRequest(gcpTargetResource)
	if err != nil {
		return "", err
	}

	token := SubjectToken_GcpWorkloadIdentityFederation{
		Url:    signedRequest.URL.String(),
		Method: signedRequest.Method,
	}

	for headerKey, headerValueList := range signedRequest.Header {
		for _, headerValue := range headerValueList {
			token.Headers = append(token.Headers, KeyValuePair{
				Key:   headerKey,
				Value: headerValue,
			})
		}
	}

	tokenJson, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	return url.QueryEscape(string(tokenJson)), nil
}

// For unit testing mock.
func (asc *AwsSdkClients) SetMockCallerIdentityOutput(output *sts.GetCallerIdentityOutput) {
	asc.callerIdentity = output
}

func (asc *AwsSdkClients) AssumeRole(roleArn string, externalID string, roleSessionName string) (*sts.AssumeRoleOutput, error) {
	input := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &roleSessionName,
	}
	if externalID != "" {
		input.ExternalId = &externalID
	}
	return asc.stsClient.AssumeRole(asc.ctx, input)
}

func withExternalID(externalID string) func(*stscreds.AssumeRoleOptions) {
	return func(options *stscreds.AssumeRoleOptions) {
		if externalID != "" {
			options.ExternalID = &externalID
		}
	}
}

func (asc *AwsSdkClients) AwsConfigFromAssumeRole(roleArn string, externalID string) aws.Config {
	awsConfig := asc.awsConfig
	creds := stscreds.NewAssumeRoleProvider(asc.stsClient, roleArn, withExternalID(externalID))
	awsConfig.Credentials = aws.NewCredentialsCache(creds)
	awsConfig.EndpointResolverWithOptions = customEndpointResolverForAssumeRole
	// asc.awsConfigsFromAssumeRole_Cache.SetWithTTL(roleArn, awsConfig, 1, time.Hour)
	return awsConfig
}

func (asc *AwsSdkClients) AwsConfigFromAssumeRoleWithRegion(region string, roleArn string, externalID string) aws.Config {
	creds := stscreds.NewAssumeRoleProvider(asc.stsClient, roleArn, withExternalID(externalID))
	awsConfig := aws.Config{
		Region:      region,
		Credentials: aws.NewCredentialsCache(creds),
	}
	// asc.awsConfigsFromAssumeRole_Cache.SetWithTTL(roleArn, awsConfig, 1, time.Hour)
	return awsConfig
}

func (asc *AwsSdkClients) populateAwsConfig() error {
	stsClient := sts.NewFromConfig(aws.Config{Region: asc.awsRegion})
	webIdentityRoleProvider := stscreds.NewWebIdentityRoleProvider(stsClient, asc.roleArn, stscreds.IdentityTokenFile(asc.webIdentityTokenFile))
	var err error
	asc.awsConfig, err = config.LoadDefaultConfig(asc.ctx, config.WithCredentialsProvider(webIdentityRoleProvider))
	asc.awsConfig.EndpointResolverWithOptions = customEndpointResolver
	return err
}

func (asc *AwsSdkClients) createClients() {
	// to avoid potential risk, removed the aws client tracing
	//otelaws.AppendMiddlewares(
	//	&asc.awsConfig.APIOptions,
	//	// propagator add a http param will break the sign, replace with NoOp propagator
	//	// https://github.com/open-telemetry/opentelemetry-go-contrib/issues/3368
	//	otelaws.WithTextMapPropagator(propagation.NewCompositeTextMapPropagator()),
	//)
	asc.s3Client = s3.NewFromConfig(asc.awsConfig)
	asc.secretsManagerClient = secretsmanager.NewFromConfig(asc.awsConfig)
	asc.sesClient = ses.NewFromConfig(asc.awsConfig)
	asc.snsClient = sns.NewFromConfig(asc.awsConfig)
	asc.sqsClient = sqs.NewFromConfig(asc.awsConfig)
	asc.stsClient = sts.NewFromConfig(asc.awsConfig)
	asc.meteringClient = marketplacemetering.NewFromConfig(asc.awsConfig)
}

// By default, the S3 client uses virtual hosted bucket addressing, but this is not well supported in LocalStack.
func (asc *AwsSdkClients) createPathStyleS3ClientForLocalStack() {
	asc.s3Client = s3.NewFromConfig(asc.awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func (asc *AwsSdkClients) createCaches() error {
	// var err error
	// if asc.awsConfigsFromAssumeRole_Cache, err = ristretto.NewCache(&ristretto.Config{
	// 	NumCounters: 500,
	// 	MaxCost:     128,
	// 	BufferItems: 128,
	// }); err != nil {
	// 	return err
	// }
	return nil
}

func (asc *AwsSdkClients) onWebIdentityTokenFileUpdate() {
	defer asc.watcher.Close()
	for {
		select {
		case event, ok := <-asc.watcher.Events:
			if !ok {
				continue
			}
			if event.Name == asc.webIdentityTokenFile && (event.Op&(fsnotify.Create|fsnotify.Write) != 0) {
				// Update the AwsConfig with the new webIdentityToken.
				err := asc.populateAwsConfig()
				if err != nil {
					asc.logger.Log("error", fmt.Sprintf("Failed to populateAwsConfig onWebIdentityTokenFileUpdate: %v", err))
					continue
				}
				asc.createClients()
			}
			continue
		case err, ok := <-asc.watcher.Errors:
			if !ok {
				continue
			}
			asc.logger.Log("error", fmt.Sprintf("Error from fsnotify watcher onWebIdentityTokenFileUpdate: %v", err))
			continue
		case <-asc.ctx.Done():
			asc.logger.Log("msg", "Exit from fsnotify watcher onWebIdentityTokenFileUpdate since of ctx.Done")
			return
		}
	}
}

// Create the secret in the AWS secret manager. Return error if failed to create.
func (asc *AwsSdkClients) CreateSecretInSecretManager(
	orgId string, secretId string, secretString string, description string) error {
	err := ValidateSecretIdWithOrgId(orgId, secretId)
	if err != nil {
		return err
	}

	client := asc.GetSecretsManagerClient()
	if client == nil {
		return fmt.Errorf("failed to get secrets manager client")
	}

	_, err = client.CreateSecret(
		asc.ctx,
		&secretsmanager.CreateSecretInput{
			Name:         aws.String(secretId),
			Description:  aws.String(description),
			SecretString: aws.String(secretString),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// Delete the secret by given the secretId from the AWS secret manager. Return error if failed to delete.
func (asc *AwsSdkClients) DeleteSecretFromSecretManager(orgId string, secretId string) error {
	err := ValidateSecretIdWithOrgId(orgId, secretId)
	if err != nil {
		return err
	}

	client := asc.GetSecretsManagerClient()
	if client == nil {
		return fmt.Errorf("failed to get secrets manager client")
	}

	_, err = client.DeleteSecret(
		asc.ctx,
		&secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretId),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		},
	)
	if err != nil {
		var resourceNotFoundException *secretsmanagertypes.ResourceNotFoundException
		if errors.As(err, &resourceNotFoundException) { // Secret not found, return nil as success.
			return nil
		}
		return err
	}

	return nil
}

// Get the secret by given the secretId from the AWS secret manager. Return (nil, nil) if not found.
// If the secretId is not connected to any orgId, then leave the orgId empty.
func (asc *AwsSdkClients) GetSecretFromSecretManager(orgId string, secretId string) (*string, error) {
	err := ValidateSecretIdWithOrgId(orgId, secretId)
	if err != nil {
		return nil, err
	}

	client := asc.GetSecretsManagerClient()
	if client == nil {
		return nil, fmt.Errorf("failed to get secrets manager client")
	}

	getSecretValueOutput, err := client.GetSecretValue(
		asc.ctx,
		&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretId),
		},
	)
	if err != nil {
		var resourceNotFoundException *secretsmanagertypes.ResourceNotFoundException
		if errors.As(err, &resourceNotFoundException) { // Secret not found, return nil as success.
			return nil, nil
		}
		return nil, err
	}

	return getSecretValueOutput.SecretString, nil
}

// Update the secret by given the secretId in the AWS secret manager. Return error if failed to update.
func (asc *AwsSdkClients) UpdateSecretInSecretManager(orgId string, secretId string, secretValue string) error {
	err := ValidateSecretIdWithOrgId(orgId, secretId)
	if err != nil {
		return err
	}

	client := asc.GetSecretsManagerClient()
	if client == nil {
		return fmt.Errorf("failed to get secrets manager client")
	}

	_, err = client.UpdateSecret(
		asc.ctx,
		&secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(secretId),
			SecretString: aws.String(secretValue),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// Create or update the secret by given the secretId in the AWS secret manager. Return error if failed to create or update.
func (asc *AwsSdkClients) CreateOrUpdateSecretInSecretManager(
	orgId string, secretId string, secretValue string, description string) error {
	secret, err := asc.GetSecretFromSecretManager(orgId, secretId)
	if err != nil {
		return err
	}

	if secret == nil {
		return asc.CreateSecretInSecretManager(orgId, secretId, secretValue, description)
	}

	return asc.UpdateSecretInSecretManager(orgId, secretId, secretValue)
}

// Ensure the secretId has prefix "org_orgId_" to avoid
// Return nil if valid, otherwise return error.
func ValidateSecretIdWithOrgId(orgId string, secretId string) error {
	// Skip validate if it is env of test.
	if shared.IsTestEnv() {
		return nil
	}

	// All secrets without "org_" prefix can be accessed without an orgID
	if !strings.HasPrefix(secretId, "org_") {
		return nil
	}

	if orgId == "" {
		return errors.New("orgID is required to retrieve a secret")
	}

	secretIdPrefix := fmt.Sprintf("org_%s", orgId)
	if strings.HasPrefix(secretId, secretIdPrefix) {
		return nil
	}

	return fmt.Errorf("secretId %s is not valid for orgId %s", secretId, orgId)
}

// CreateAwsSdkClientsForTesting creates AwsSdkClients for local test, github test or return error.
func CreateAwsSdkClientsForTesting(
	ctx context.Context, logger log.Logger, authMethod string) (*AwsSdkClients, error) {
	switch authMethod {
	case shared.AWS_AUTH_METHOD_SSO:
		return NewAwsSdkClients_SSO(ctx, logger, structs.AWS_PROFILE_TEST)
	case shared.AWS_AUTH_METHOD_ENV:
		return NewAwsSdkClients_ENV(ctx, logger)
	default:
		return nil, fmt.Errorf("invalid AWS_AUTH_METHOD: %s for testing", authMethod)
	}
}
