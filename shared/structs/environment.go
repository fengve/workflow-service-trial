package structs

import (
	"github.com/Netflix/go-env"
)

// Attention: Only set the environment variable with required=true when you are 100% sure that it's required by all servcies.
type Environment struct {
	Env   string `env:"ENV"`
	RdsDb struct {
		Endpoint         string `env:"RDS_DB_ENDPOINT"`
		ReadEndpoint     string `env:"RDS_DB_READ_ENDPOINT"`
		Port             string `env:"RDS_DB_PORT"`
		Name             string `env:"RDS_DB_NAME,default=postgres"`
		User             string `env:"RDS_DB_USER"`
		PasswordSecretId string `env:"RDS_DB_PASSWORD_SECRET_ID"`
		Password         string `env:"RDS_DB_PASSWORD"` // For marketplace-service only
	}
	ServicePort struct {
		MarketplaceService string `env:"MARKETPLACE_SERVICE_PORT"` // For marketplace-service only.
		WorkflowService    string `env:"WORKFLOW_SERVICE_PORT"`    // For workflow-service only.
	}
	Slack struct {
		AppScope       string `env:"SLACK_APP_SCOPE"`
		ClientId       string `env:"SLACK_CLIENT_ID"`
		ClientSecretId string `env:"SLACK_CLIENT_SECRET_ID"` // The AWS Secret Manager Secret ID to store the Slack client secret.
		ClientSecret   string
	}
	Temporal struct {
		HostPort string `env:"TEMPORAL_HOST_PORT"`
	}
	AllowOrigins                 string `env:"CORS_ALLOW_ORIGINS,default=*"`     // For marketplace-service only
	NotificationEventSqsQueueUrl string `env:"NOTIFICATION_EVENT_SQS_QUEUE_URL"` // sqs queue url for notification events.
	SugerApiEndpoint             string `env:"SUGER_API_ENDPOINT"`
	AwsAuthMethod                string `env:"AWS_AUTH_METHOD"`
	Extras                       env.EnvSet
}
