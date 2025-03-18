package structs

import (
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"time"

	"github.com/slack-go/slack"
)

type PartnerService string //@name PartnerService

const (
	PartnerService_BIGQUERY PartnerService = "BIGQUERY" // for GCP Bigquery.
	PartnerService_CHATBOT  PartnerService = "CHATBOT"
	PartnerService_DRIVE    PartnerService = "DRIVE" // for Google Drive.
	PartnerService_EMAIL    PartnerService = "EMAIL"
	PartnerService_STORAGE  PartnerService = "STORAGE" // for Google Cloud Storage.
)

type Partner string //@name Partner

const (
	Partner_GCP    Partner = "GCP"
	Partner_GOOGLE Partner = "GOOGLE"
	Partner_SLACK  Partner = "SLACK"
)

type OrganizationStatus string //@name OrganizationStatus

const (
	OrganizationStatus_ACTIVE           OrganizationStatus = "ACTIVE"
	OrganizationStatus_PENDING_APPROVAL OrganizationStatus = "PENDING_APPROVAL"
	OrganizationStatus_SUSPENDED        OrganizationStatus = "SUSPENDED"
	OrganizationStatus_DELETED          OrganizationStatus = "DELETED"
)

type IntegrationStatus string //@name IntegrationStatus

const (
	IntegrationStatus_CREATED      IntegrationStatus = "CREATED"
	IntegrationStatus_VERIFIED     IntegrationStatus = "VERIFIED"
	IntegrationStatus_NOT_VERIFIED IntegrationStatus = "NOT_VERIFIED"
)

// The entire Integration config containing the configs of all supported integrations.
type IntegrationConfig struct {
	AwsMarketplaceIntegrationConfig *AwsIntegrationConfig         `json:"awsMarketplaceIntegrationConfig,omitempty"`
	GoogleDriveIntegrationConfig    *GoogleDriveIntegrationConfig `json:"googleDriveIntegrationConfig,omitempty"`
	SlackIntegrationConfig          *SlackIntegrationConfig       `json:"slackIntegrationConfig,omitempty"`
} //@name IntegrationConfig

type SlackIntegrationConfig struct {
	OAuthUrl string `json:"oauthUrl,omitempty"` // the oauth url for clients to install suger slack app.
} //@name SlackIntegrationConfig

type GoogleDriveIntegrationConfig struct {
	ClientId    string `json:"clientId,omitempty"`
	RedirectUri string `json:"redirectUri,omitempty"`
	OauthHost   string `json:"oauthHost,omitempty"`
	UserId      string `json:"userId,omitempty"` // The Suger user ID.
} //@name GoogleDriveIntegrationConfig

type ApiClientAccessToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type" default:"Bearer"`
	ExpiresIn   int32     `json:"expires_in" default:"3600"`     // The token expires in 1 hour
	ExpiresOn   time.Time `json:"expires_on" format:"date-time"` // The UTC timestamp when the token expires
} //@name ApiClientAccessToken

type IntegrationInfo struct {
	// For GCP Marketplace / Google Bigquery / Google Drive / Google Cloud Storage
	GcpIntegration *GcpIntegration `json:"gcpIntegration,omitempty"`
	// Integration for Slack
	SlackIntegration *SlackIntegration `json:"slackIntegration,omitempty"`
} //@name IntegrationInfo

type OrganizationInfo struct {
	// Store the basic configuration information of the organization.
	OrganizationConfigInfo *OrganizationConfigInfo `json:"organizationConfigInfo,omitempty"`
	// Store the notification configuration information of the organization.
	NotificationConfigInfo *NotificationConfigInfo `json:"notificationConfigInfo,omitempty"`
	// Store the new client signup page configuration information of the organization.
	ClientSignupPageConfigInfo *ClientSignupPageConfigInfo `json:"clientSignupPageConfigInfo,omitempty"`
} //@name OrganizationInfo

type ClientSignupPageConfigInfo struct {
	// The signup ID for the new client signup page url.
	// It is populated by Suger service when creating the new client signup page.
	SignupId string `json:"signupId,omitempty"`
	// The company name of the seller/ISV to show in the client signup page.
	CompanyName string `json:"companyName,omitempty"`
	// The company logo url of the seller/ISV.
	CompanyLogoUrl string `json:"companyLogoUrl,omitempty"`
	// The cc email contacts of the new client signup notification.
	CcEmails []string `json:"ccEmails,omitempty"`
	// The email subject of the new client signup notification.
	NotificationEmailSubject string `json:"notificationEmailSubject,omitempty"`
	// The signup landing image aws url of the organization
	LandingImageUrl string `json:"landingImageUrl,omitempty"`
	// The video link of the product. Optional.
	VideoLink string `json:"videoLink,omitempty"`
	// The public notes to append in new client signup notification email.
	PublicNotes string `json:"publicNotes,omitempty"`
	// If true, the cloud partner logo will no show in the new client signup page.
	HideCloudPartnerLogo bool `json:"hideCloudPartnerLogo,omitempty"`
	// Custom template used as template
	CustomTemplate string `json:"customTemplate,omitempty"`
	// If true, custom template will be used
	EnableCustomTemplate bool `json:"enableCustomTemplate,omitempty"`
	// Enable headless entitlements report.
	// Once enabled, the headless entitlements report will be sent as notification events.
	// Each report contains the headless entitlements who are created in the past 3 days, but never have new client signup connected.
	EnableHeadlessEntitlementsReport bool `json:"enableHeadlessEntitlementsReport,omitempty"`

	// Enable to update buyer information with the new client signup info.
	// If true, the buyer information will be updated with the new client signup info.
	// If false, the buyer information will not be updated with the new client signup info.
	EnableUpdateBuyer bool `json:"enableUpdateBuyer,omitempty"`
} //@name ClientSignupPageConfigInfo

type OrganizationConfigInfo struct {
	// Enable the product whitelist for the organization.
	EnableProductWhitelist bool `json:"enableProductWhitelist,omitempty"`
	// The product whitelist for the organization.
	ProductWhitelist []string `json:"productWhitelist,omitempty"`
	// The Client ID of the custom auth0 application. This is used to allow login with the custom SSO.
	Auth0ApplicationClientId string `json:"auth0ApplicationClientId,omitempty"`
	// Whether to enforce the custom SSO login via the custom auth0 application.
	// If true, the user can only login via the custom auth0 application.
	// If false, the user can login via the custom auth0 application, or the default Suger auth0 application.
	EnforceCustomLogin bool `json:"enforceCustomLogin,omitempty"`
	// Whether to use new filler filed mapping
	EnableSalesforceAwsFieldMappingV2 bool `json:"enableSalesforceAwsFieldMappingV2,omitempty"`
} //@name OrganizationConfigInfo

type NotificationConfigInfo struct {
	// Enable the product whitelist for the webhook notification.
	// The default is false, and allow all the products notifications sent via webhook.
	EnableWebhookProductWhitelist bool `json:"enableWebhookNotification,omitempty"`
	// The product whitelist (suger Product Id) for the webhook notification.
	// If the product is not in the whitelist, the notification will be sent via webhook.
	WebhookProductWhitelist []string `json:"webhookProductWhitelist,omitempty"`
	// Enable to email notification events to organization admins/editors.
	// The default is false, and does not send emails to organization admins/editors.
	EnableEmailNotification bool `json:"enableEmailNotification,omitempty"`
	// Enable to email notification events to buyers
	DisableEmailNotificationOnOfferReady bool `json:"disableEmailNotificationOnOfferReady,omitempty"`
	// The list of email notification scope configs that defines which
	// NotificationScopes are sent to which emails.
	// Only applicable when EnableEmailNotification is true.
	EmailNotificationScopeConfigs []EmailNotificationScopeConfig `json:"emailNotificationScopeConfigs,omitempty"`
} //@name NotificationConfigInfo

type EmailNotificationScopeConfig struct {
	// The email recipients who will receive this email notifications
	Recipients []rdsDbLib.IdentityUser `json:"recipients,omitempty"`
} //@name EmailNotificationScopeConfig

type SlackIntegration struct {
	RedirectUrl string `json:"redirectUrl,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
	// The scope of the access token. multiple scopes are separated by comma.
	Scope           string                              `json:"scope,omitempty"`
	TokenType       string                              `json:"tokenType,omitempty"`
	BotUserID       string                              `json:"botUserId,omitempty"`
	AppID           string                              `json:"appId,omitempty"`
	Team            *slack.OAuthV2ResponseTeam          `json:"team,omitempty"`
	IncomingWebhook *slack.OAuthResponseIncomingWebhook `json:"incomingWebhook,omitempty"`
	Enterprise      *slack.OAuthV2ResponseEnterprise    `json:"enterprise,omitempty"`
	AuthedUser      *slack.OAuthV2ResponseAuthedUser    `json:"authedUser,omitempty"`
	RefreshToken    string                              `json:"refreshToken,omitempty"`
	ExpiresIn       int                                 `json:"expiresIn,omitempty"`

	// Deprecated. Use EnableNotification instead. It will be removed in the future.
	// Disable the general slack notification for the organization.
	// The default is false, which means the general slack notification is enabled.
	DisableNotification bool `json:"disableNotification,omitempty"`

	// Enable the slack notification
	// The default is false, which means the slack notification is disabled.
	EnableNotification bool `json:"enableNotification,omitempty"`

	// The list of Slack notification scope configs that defines which
	// NotificationScopes are sent to which slack channels or users.
	// Only applicable when EnableNotification is true.
	NotificationScopeConfigs []SlackNotificationScopeConfig `json:"notificationScopeConfigs,omitempty"`
} //@name SlackIntegration

type SlackNotificationScopeConfig struct {
	// The slack channels who will receive the notifications
	Channels []SlackChannel `json:"channels,omitempty"`
	// the notification scopes that define which type of notification events shall be sent to the Recipients.
} //@name SlackNotificationScopeConfig

type SlackChannel struct {
	// the channel ID
	ID string `json:"id,omitempty"`
	// the channel name
	Name string `json:"name"`
} //@name SlackChannel

type GetApiClientAccessTokenParams struct {
	OrganizationID string `json:"organizationID" validate:"required"`
	ID             string `json:"id" validate:"required"`     // The ID of the API Client.
	Secret         string `json:"secret" validate:"required"` // The secret of the API Client.
} //@name GetApiClientAccessTokenParams

type OrgFile struct {
	ID           string    `json:"id"`   // File ID
	Name         string    `json:"name"` // File Name
	Key          string    `json:"key"`  // File Key for S3 Bucket
	CreationTime time.Time `json:"creationTime" format:"date-time"`
	SignedUrl    string    `json:"signedUrl,omitempty"` // AWS S3 signed url, expires in 1 hour.
} //@name OrgFile

type AwsIntegrationConfig struct {
	// The clients need to create an IAM role named by this for Suger to access their AWS services.
	IamRoleName string `json:"iamRoleName"`
	// JSON format string. The assume role policy for the IAM role
	// so that Suger service can assume the role to access client's AWS services.
	AssumeRolePolicy string `json:"assumeRolePolicy"`
	// The ARNs of the required AWS managed policies for Suger to access client's AWS services.
	ManagedPolicyArns []string `json:"managedPolicyArns"`
	// The Url of CloudFormation stack to allow client to create IAM role for suger to access their AWS services.
	CloudFormationStackUrl string `json:"cloudFormationStackUrl"`
	// The external ID for assuming IAM role. If empty, means no external ID set or needed.
	ExternalID string `json:"externalID"`
} //@name AwsIntegrationConfig

type GcpIntegration struct {
	// The GCP Organization ID of the GCP Marketplace ISV/Seller.
	// Required for GCP marketplace resell, such as Reseller Private Offer Plan.
	GcpOrganizationId string `json:"gcpOrganizationId,omitempty"`
	// The GCP Project ID of the GCP Marketplace ISV/Seller where the marketplace producer portal is enabled.
	GcpProjectId string `json:"gcpProjectId,omitempty"`
	// The GCP Project Number of the GCP Marketplace ISV/Seller where the marketplace producer portal is enabled.
	GcpProjectNumber       string `json:"gcpProjectNumber,omitempty"`
	WorkloadIdentityPoolId string `json:"workloadIdentityPoolId,omitempty"`
	IdentityProviderId     string `json:"identityProviderId,omitempty"`

	// The service account email to access GCP resources.
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty"`
	// The service account private key Json content, downloaded from GCP console.
	// It is used to authenticate the service account to access GCP resources.
	// Not available in the response of the API.
	ServiceAccountPrivateKeyJson string `json:"serviceAccountPrivateKeyJson,omitempty"`
	// Impersonate subject used by the service account to access GCP resources. Required in Gmail Integration.
	ImpersonateSubject string `json:"impersonateSubject,omitempty"`
	// Google cloud storage buckets name accessible by the service account. Required for Google Cloud Storage integration.
	Buckets []string `json:"buckets,omitempty"`
	// The secret key used to store/retrieve ServiceAccountPrivateKeyJson in AWS Secrets Manager.
	// For internal usage only.
	SecretKey string `json:"secretKey,omitempty"`
	// The Google OAuth Credential to access Google APIs.
	OauthCredential *GoogleOauthCredential `json:"oauthCredential,omitempty"`
	// The Google Oauth 2.0 scopes to access Google APIs.
	// Defined in https://developers.google.com/identity/protocols/oauth2/scopes
	// Optional, if not set, Suger will use the default scope defined internally.
	OauthScopes []string `json:"oauthScopes,omitempty"`

	// The GCP Marketplace Partner ID, it is also called as Provider ID somewhere.
	PartnerId string `json:"partnerId,omitempty"`

	// The resource name of the Pub/Sub topic to receive notifications from Google
	// when a user signs up for your service, purchases a plan, or changes an existing plan.
	PubsubTopic string `json:"pubsubTopic,omitempty"`
	// The resource name of the Pub/Sub subscription to receive notifications from Google cloud marketplace.
	PubsubSubscription string `json:"pubsubSubscription,omitempty"`
	// The array of service resource names of the listings in GCP Marketplace.
	ServiceNames []string `json:"serviceNames,omitempty"`
	// The GCP storage bucket name to store the GCP Marketplace reports.
	ReportBucket string `json:"reportBucket,omitempty"`
	// The UTC date when GCP Marketplace reports start to generate.
	ReportStartDate *time.Time `json:"reportStartDate,omitempty" format:"date-time"`
	// Is GCP Marketplace Report full-sync done.
	ReportFullSyncDone bool `json:"reportFullSyncDone,omitempty"`

	// Disable GCP Marketplace Report sync.
	// If true, Suger stop to sync GCP Marketplace reports.
	ReportSyncDisabled bool `json:"reportSyncDisabled,omitempty"`
	// Is GCP Marketplace Revenue Record full-sync done.
	RevenueRecordFullSyncDone bool `json:"revenueRecordFullSyncDone,omitempty"`
	// Disable GCP Marketplace Revenue Record sync.
	// If true, Suger stop to sync GCP Marketplace Revenue Records.
	RevenueRecordSyncDisabled bool `json:"revenueRecordSyncDisabled,omitempty"`

	// Disable Usage Metering to GCP Marketplace.
	// If true, Suger stop to report usage records to GCP Marketplace.
	UsageMeteringDisabled bool `json:"usageMeteringDisabled,omitempty"`

	// Enable GCP marketplace sync from GCP Chrome.
	// If true, Suger will sync GCP Marketplace Product & Private Offer from GCP Chrome.
	EnableChromeSync bool `json:"enableChromeSync,omitempty"`

	// Enable manually approve the GCP Marketplace Entitlement.
	// If true, Suger will not automatically approve the GCP Marketplace Entitlement.
	// Util the GCP Marketplace Entitlement is manually approved, it will not be activated.
	EnableManualApproveEntitlement bool `json:"enableManualApproveEntitlement,omitempty"`

	// The Login URL for GCP Marketplace buyers to login to your service with or without SSO JWT token.
	// If not set, the login will be redirected to the Product's fulfillment URL.
	LoginUrl string `json:"loginUrl,omitempty"`

	// If enabled, Suger will redirect the new client to the signup page even the entitlement is not found.
	// If disabled, Suger will redirect the new client to the error page if the entitlement is not found.
	SignupRedirectWithoutEntitlementEnabled bool `json:"signupRedirectWithoutEntitlementEnabled,omitempty"`
} //@name GcpIntegration

type GoogleOauthCredential struct {
	RedirectUrl  string `json:"redirectUrl,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
	TokenType    string `json:"tokenType,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	AcquiredOn   int64  `json:"acquiredOn,omitempty"` // UTC timestamp on receiving the auth response
	ExpiresIn    int64  `json:"expiresIn,omitempty"`
	Scope        string `json:"scope,omitempty"`
} //@name GoogleOauthCredential

// GCP Auth Private Key Json.
type GcpAuthPrivateKey struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
} //@name GcpAuthPrivateKey

type AwsSnsSubscriptionConfirmationEvent struct {
	Type             string `json:"Type"`
	MessageId        string `json:"MessageId,omitempty"`
	Token            string `json:"Token,omitempty"`
	TopicArn         string `json:"TopicArn,omitempty"`
	Message          string `json:"Message,omitempty"`
	SubscribeURL     string `json:"SubscribeURL,omitempty"`
	SignatureVersion string `json:"SignatureVersion,omitempty"`
	Signature        string `json:"Signature,omitempty"`
	SigningCertURL   string `json:"SigningCertURL,omitempty"`
	Timestamp        string `json:"Timestamp,omitempty"`
} //@name AwsSnsSubscriptionConfirmationEvent
