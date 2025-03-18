package structs

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

type WorkflowCallerPolicy string //@name WorkflowCallerPolicy

const (
	WorkflowCallerPolicy_Any           WorkflowCallerPolicy = "any"
	WorkflowCallerPolicy_None          WorkflowCallerPolicy = "none"
	WorkflowCallerPolicy_FromAList     WorkflowCallerPolicy = "workflowsFromAList"
	WorkflowCallerPolicy_FromSameOwner WorkflowCallerPolicy = "workflowsFromSameOwner"
)

type WorkflowReleaseChannel string //@name WorkflowReleaseChannel

const (
	WorkflowReleaseChannel_Stable   WorkflowReleaseChannel = "stable"
	WorkflowReleaseChannel_Beta     WorkflowReleaseChannel = "beta"
	WorkflowReleaseChannel_Nightly  WorkflowReleaseChannel = "nightly"
	WorkflowReleaseChannel_Dev      WorkflowReleaseChannel = "dev"
	WorkflowReleaseChannel_External WorkflowReleaseChannel = "external"
)

type WorkflowSettingExecutionMode string //@name WorkflowSettingExecutionMode

const (
	WorkflowSettingExecutionMode_Regular WorkflowSettingExecutionMode = "regular"
	WorkflowSettingExecutionMode_Queue   WorkflowSettingExecutionMode = "queue"
)

// n8n WorkflowExecuteMode
type WorkflowExecutionMode string //@name WorkflowExecutionMode
const (
	WorkflowExecutionMode_Cli        WorkflowExecutionMode = "cli"
	WorkflowExecutionMode_Error      WorkflowExecutionMode = "error"
	WorkflowExecutionMode_Integrated WorkflowExecutionMode = "integrated"
	WorkflowExecutionMode_Internal   WorkflowExecutionMode = "internal"
	WorkflowExecutionMode_Manual     WorkflowExecutionMode = "manual"
	WorkflowExecutionMode_Retry      WorkflowExecutionMode = "retry"
	WorkflowExecutionMode_Trigger    WorkflowExecutionMode = "trigger"
	WorkflowExecutionMode_Webhook    WorkflowExecutionMode = "webhook"
)

type WorkflowPushBackend string //@name WorkflowPushBackend

const (
	WorkflowPushBackend_SSE       WorkflowPushBackend = "sse"
	WorkflowPushBackend_WebSocket WorkflowPushBackend = "websocket"
)

type WorkflowDeploymentType string //@name WorkflowDeploymentType

const (
	WorkflowDeployment_Default     WorkflowDeploymentType = "default"
	WorkflowDeployment_N8nInternal WorkflowDeploymentType = "n8n-internal"
	WorkflowDeployment_Cloud       WorkflowDeploymentType = "cloud"
	WorkflowDeployment_DesktopMac  WorkflowDeploymentType = "desktop_mac"
	WorkflowDeployment_DesktopWin  WorkflowDeploymentType = "win"
)

type WorkflowExpressionEvaluatorType string //@name WorkflowExpressionEvaluatorType

const (
	WorkflowExpressionEvaluatorType_Tmpl       WorkflowExpressionEvaluatorType = "tmpl"
	WorkflowExpressionEvaluatorType_Tournament WorkflowExpressionEvaluatorType = "tournament"
)

type WorkflowRole string //@name WorkflowRole

const (
	WorkflowRole_Default WorkflowRole = "default"
	WorkflowRole_Owner   WorkflowRole = "owner"
	WorkflowRole_Member  WorkflowRole = "member"
	WorkflowRole_Admin   WorkflowRole = "admin"
)

type WorkflowExecutionStatus string //@name WorkflowExecutionStatus

const (
	WorkflowExecutionStatus_Canceled WorkflowExecutionStatus = "canceled"
	WorkflowExecutionStatus_Crashed  WorkflowExecutionStatus = "crashed"
	WorkflowExecutionStatus_Error    WorkflowExecutionStatus = "error"
	WorkflowExecutionStatus_Failed   WorkflowExecutionStatus = "failed"
	WorkflowExecutionStatus_New      WorkflowExecutionStatus = "new"
	WorkflowExecutionStatus_Running  WorkflowExecutionStatus = "running"
	WorkflowExecutionStatus_Success  WorkflowExecutionStatus = "success"
	WorkflowExecutionStatus_Unknown  WorkflowExecutionStatus = "unknown"
	WorkflowExecutionStatus_Waiting  WorkflowExecutionStatus = "waiting"
	WorkflowExecutionStatus_Warning  WorkflowExecutionStatus = "warning"
)

type WorkflowBinaryFileType string //@name WorkflowBinaryFileType

const (
	WorkflowBinaryFileType_Text  WorkflowBinaryFileType = "text"
	WorkflowBinaryFileType_JSON  WorkflowBinaryFileType = "json"
	WorkflowBinaryFileType_Image WorkflowBinaryFileType = "image"
	WorkflowBinaryFileType_Audio WorkflowBinaryFileType = "audio"
	WorkflowBinaryFileType_Video WorkflowBinaryFileType = "video"
	WorkflowBinaryFileType_PDF   WorkflowBinaryFileType = "pdf"
	WorkflowBinaryFileType_HTML  WorkflowBinaryFileType = "html"
)

type WorkflowNodeOnError string //@name WorkflowNodeOnError

const (
	WorkflowNodeOnError_ContinueErrorOutput   WorkflowNodeOnError = "continueErrorOutput"
	WorkflowNodeOnError_ContinueRegularOutput WorkflowNodeOnError = "continueRegularOutput"
	WorkflowNodeOnError_StopWorkflow          WorkflowNodeOnError = "stopWorkflow"
)

type WorkflowSaveDataExecution string //@name WorkflowSaveDataExecution

const (
	WorkflowSaveDataExecution_Default WorkflowSaveDataExecution = "DEFAULT"
	WorkflowSaveDataExecution_All     WorkflowSaveDataExecution = "all"
	WorkflowSaveDataExecution_None    WorkflowSaveDataExecution = "none"
)

type WorkflowVersionNotifications struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
	InfoURL  string `json:"infoUrl"`
} //@name WorkflowVersionNotifications

type WorkflowOauthCallbackUrls struct {
	Oauth1 string `json:"oauth1"`
	Oauth2 string `json:"oauth2"`
} //@name WorkflowOauthCallbackUrls

type WorkflowTelemetry struct {
	Enabled bool                     `json:"enabled"`
	Config  *WorkflowTelemetryConfig `json:"config"`
} //@name WorkflowTelemetry

type WorkflowTelemetryConfig struct {
	Key string `json:"key"`
	URL string `json:"url"`
} //@name WorkflowTelemetryConfig

type WorkflowPosthog struct {
	Enabled                 bool   `json:"enabled"`
	APIHost                 string `json:"apiHost"`
	APIKey                  string `json:"apiKey"`
	Autocapture             bool   `json:"autocapture"`
	DisableSessionRecording bool   `json:"disableSessionRecording"`
	Debug                   bool   `json:"debug"`
}

type WorkflowUserManagement struct {
	Quota                int    `json:"quota"`
	ShowSetupOnFirstLoad bool   `json:"showSetupOnFirstLoad"`
	SMTPSetup            bool   `json:"smtpSetup"`
	AuthenticationMethod string `json:"authenticationMethod"`
} //@name WorkflowUserManagement

type WorkflowSSO struct {
	Saml *WorkflowSSOLogin `json:"saml"`
	Ldap *WorkflowSSOLogin `json:"ldap"`
} //@name WorkflowSSO

type WorkflowSSOLogin struct {
	LoginEnabled bool   `json:"loginEnabled"`
	LoginLabel   string `json:"loginLabel"`
} //@name WorkflowSSOLogin

type WorkflowPublicAPISwaggerUI struct {
	Enabled bool `json:"enabled"`
} //@name WorkflowPublicAPISwaggerUI

type WorkflowPublicAPI struct {
	Enabled       bool                        `json:"enabled"`
	LatestVersion int                         `json:"latestVersion"`
	Path          string                      `json:"path"`
	SwaggerUI     *WorkflowPublicAPISwaggerUI `json:"swaggerUi"`
} //@name WorkflowPublicAPI

type WorkflowTemplates struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
} //@name WorkflowTemplates

type WorkflowDeployment struct {
	Type WorkflowDeploymentType `json:"type"`
} //@name WorkflowDeployment

type WorkflowAllowedModules struct {
	BuiltIn  []string `json:"builtIn,omitempty"`
	External []string `json:"external,omitempty"`
} //@name WorkflowAllowedModules

type WorkflowLicense struct {
	Environment string `json:"environment"`
} //@name WorkflowLicense

type WorkflowSettingVariables struct {
	Limit int `json:"limit"`
} //@name WorkflowSettingVariables

type WorkflowExpressions struct {
	Evaluator WorkflowExpressionEvaluatorType `json:"evaluator"`
} //@name WorkflowExpressions

type WorkflowMfa struct {
	Enabled bool `json:"enabled"`
} //@name WorkflowMfa

type WorkflowUIBanners struct {
	Dismissed []string `json:"dismissed"`
} //@name WorkflowUIBanners

type WorkflowAI struct {
	Enabled bool `json:"enabled"`
} //@name WorkflowAI

type WorkflowHistory struct {
	PruneTime        int `json:"pruneTime"`
	LicensePruneTime int `json:"licensePruneTime"`
} //@name WorkflowHistory

type WorkflowUISettings struct {
	EndpointForm                      string                        `json:"endpointForm,omitempty"`
	EndpointFormTest                  string                        `json:"endpointFormTest,omitempty"`
	EndpointFormWaiting               string                        `json:"endpointFormWaiting,omitempty"`
	EndpointWebhook                   string                        `json:"endpointWebhook,omitempty"`
	EndpointWebhookTest               string                        `json:"endpointWebhookTest,omitempty"`
	SaveDataErrorExecution            WorkflowSaveDataExecution     `json:"saveDataErrorExecution"`
	SaveDataSuccessExecution          WorkflowSaveDataExecution     `json:"saveDataSuccessExecution"`
	SaveManualExecutions              bool                          `json:"saveManualExecutions"`
	ExecutionTimeout                  int64                         `json:"executionTimeout"`
	MaxExecutionTimeout               int64                         `json:"maxExecutionTimeout"`
	WorkflowCallerPolicyDefaultOption WorkflowCallerPolicy          `json:"workflowCallerPolicyDefaultOption"`
	OauthCallbackUrls                 WorkflowOauthCallbackUrls     `json:"oauthCallbackUrls"`
	Timezone                          string                        `json:"timezone"`
	URLBaseWebhook                    string                        `json:"urlBaseWebhook"`
	URLBaseEditor                     string                        `json:"urlBaseEditor"`
	VersionCli                        string                        `json:"versionCli"`
	ReleaseChannel                    WorkflowReleaseChannel        `json:"releaseChannel"`
	VersionNotifications              *WorkflowVersionNotifications `json:"versionNotifications"`
	InstanceID                        string                        `json:"instanceId"`
	Telemetry                         *WorkflowTelemetry            `json:"telemetry"`
	PersonalizationSurveyEnabled      bool                          `json:"personalizationSurveyEnabled"`
	DefaultLocale                     string                        `json:"defaultLocale"`
	UserManagement                    *WorkflowUserManagement       `json:"userManagement"`
	SSO                               *WorkflowSSO                  `json:"sso"`
	PublicAPI                         *WorkflowPublicAPI            `json:"publicApi"`
	WorkflowTagsDisabled              bool                          `json:"workflowTagsDisabled"`
	LogLevel                          string                        `json:"logLevel"`
	HiringBannerEnabled               bool                          `json:"hiringBannerEnabled"`
	Templates                         *WorkflowTemplates            `json:"templates"`
	OnboardingCallPromptEnabled       bool                          `json:"onboardingCallPromptEnabled"`
	MissingPackages                   bool                          `json:"missingPackages"`
	ExecutionMode                     WorkflowSettingExecutionMode  `json:"executionMode"`
	PushBackend                       WorkflowPushBackend           `json:"pushBackend"`
	CommunityNodesEnabled             bool                          `json:"communityNodesEnabled"`
	Deployment                        *WorkflowDeployment           `json:"deployment"`
	IsNpmAvailable                    bool                          `json:"isNpmAvailable"`
	AllowedModules                    *WorkflowAllowedModules       `json:"allowedModules"`
	Enterprise                        map[string]bool               `json:"enterprise"`
	HideUsagePage                     bool                          `json:"hideUsagePage"`
	License                           *WorkflowLicense              `json:"license"`
	Variables                         *WorkflowSettingVariables     `json:"variables"`
	Expressions                       *WorkflowExpressions          `json:"expressions"`
	Mfa                               *WorkflowMfa                  `json:"mfa"`
	Banners                           *WorkflowUIBanners            `json:"banners"`
	Ai                                *WorkflowAI                   `json:"ai"`
	WorkflowHistory                   *WorkflowHistory              `json:"workflowHistory"`
} //@name WorkflowUISettings

type GetWorkflowUISettingsResponse struct {
	Data *WorkflowUISettings `json:"data"`
} //@name GetWorkflowUISettingsResponse

type WorkflowUserResponseGlobalRole struct {
	CreatedAt *time.Time   `json:"createdAt,omitempty"`
	UpdatedAt *time.Time   `json:"updatedAt,omitempty"`
	ID        string       `json:"id"`
	Name      WorkflowRole `json:"name"`
	Scope     string       `json:"scope"`
} //@name WorkflowUserResponseGlobalRole

type WorflowUserPersonalizationAnswers struct {
	Email                string      `json:"email,omitempty"`
	CodingSkill          string      `json:"codingSkill,omitempty"`
	CompanyIndustry      []string    `json:"companyIndustry,omitempty"`
	CompanySize          string      `json:"companySize,omitempty"`
	OtherCompanyIndustry string      `json:"otherCompanyIndustry,omitempty"`
	OtherWorkArea        string      `json:"otherWorkArea,omitempty"`
	WorkArea             interface{} `json:"workArea,omitempty"`
} //@name WorflowUserPersonalizationAnswers

type WorkflowUserSettings struct {
	IsOnboarded               bool   `json:"isOnboarded,omitempty"`
	FirstSuccessfulWorkflowId string `json:"firstSuccessfulWorkflowId,omitempty"`
	UserActivated             bool   `json:"userActivated,omitempty"`
	AllowSSOManualLogin       bool   `json:"allowSSOManualLogin,omitempty"`
} //@name WorkflowUserSettings

type WorkflowUser struct {
	ID                     string                            `json:"id"`
	Email                  string                            `json:"email"`
	FirstName              string                            `json:"firstName"`
	LastName               string                            `json:"lastName"`
	PersonalizationAnswers WorflowUserPersonalizationAnswers `json:"personalizationAnswers"`
	CreatedAt              time.Time                         `json:"createdAt"`
	IsPending              bool                              `json:"isPending"`
	HasRecoveryCodesLeft   bool                              `json:"hasRecoveryCodesLeft"`
	GlobalRole             WorkflowUserResponseGlobalRole    `json:"globalRole"`
	GlobalScopes           []string                          `json:"globalScopes"`
	SignInType             string                            `json:"signInType"`
	Disabled               bool                              `json:"disabled"`
	Settings               WorkflowUserSettings              `json:"settings"`
	InviteAcceptURL        string                            `json:"inviteAcceptUrl,omitempty"`
	IsOwner                bool                              `json:"isOwner"`
	FeatureFlags           interface{}                       `json:"featureFlags"`
} //@name WorkflowUser

type GetWorkflowLoginResponse struct {
	Data *WorkflowUser `json:"data"`
} //@name GetWorkflowLoginResponse

type ListWorkflowUsersResponse struct {
	Data []WorkflowUser `json:"data"`
} //@name ListWorkflowUsersResponse

type ListWorkflowCrendentialsResponse struct {
	Data []string `json:"data"`
} //@name ListWorkflowCrendentialsResponse

// n8n IBinaryData
type WorkflowBinaryData struct {
	Data string `json:"data"`
	// If the data is base64 encoded
	Base64Encoded bool                   `json:"base64Encoded"`
	MimeType      string                 `json:"mimeType"`
	FileType      WorkflowBinaryFileType `json:"fileType,omitempty"`
	FileName      string                 `json:"fileName,omitempty"`
	Directory     string                 `json:"directory,omitempty"`
	FileExtension string                 `json:"fileExtension,omitempty"`
	FileSize      string                 `json:"fileSize,omitempty"`
	ID            string                 `json:"id,omitempty"`
} //@name WorkflowBinaryData

type WorkflowNodeExecutionDataInnerMap map[string]interface{} //@name WorkflowNodeExecutionDataInnerMap

// n8n INodeExecutionData
type WorkflowNodeExecutionData struct {
	WorkflowNodeExecutionDataInnerMap
	Json       map[string]interface{}          `json:"json,omitempty"`
	Binary     map[string][]WorkflowBinaryData `json:"binary"`
	Error      *WorkflowExecutionError         `json:"error"`
	PairedItem interface{}                     `json:"pairedItem,omitempty"`
	Index      int64                           `json:"index,omitempty"`
} //@name WorkflowNodeExecutionData

// n8n ISourceData
type WorkflowSourceData struct {
	PreviousNode string `json:"previousNode,omitempty"`
	// If undefined "0" gets used
	PreviousNodeOutput int64 `json:"previousNodeOutput"`
	// If undefined "0" gets used
	PreviousNodeRun int64 `json:"previousNodeRun"`
} //@name WorkflowSourceData

// n8n ITaskSubRunMetadata
type WorkflowExecutionTaskSubRunMetadata struct {
	Node     string `json:"node"`
	RunIndex int64  `json:"runIndex"`
} //@name WorkflowExecutionTaskSubRunMetadata

type WorkflowExecutionTaskMetadata struct {
	SubRun []WorkflowExecutionTaskSubRunMetadata `json:"subRun,omitempty"`
} //@name WorkflowExecutionTaskMetadata

// n8n ITaskData
type WorkflowExecutionTaskData struct {
	StartTime       int64                                    `json:"startTime"`
	ExecutionTime   int64                                    `json:"executionTime"`
	ExecutionStatus WorkflowExecutionStatus                  `json:"executionStatus,omitempty"`
	Data            map[string][]NodeData                    `json:"data,omitempty"`
	InputOverride   map[string][][]WorkflowNodeExecutionData `json:"inputOverride,omitempty"`
	Error           *WorkflowExecutionError                  `json:"error,omitempty"`
	Source          []WorkflowSourceData                     `json:"source,omitempty"`
	Metadata        *WorkflowExecutionTaskMetadata           `json:"metadata,omitempty"`
} //@name WorkflowExecutionTaskData

// n8n ManualRunPayload
type WorkflowManualRunRequest struct {
	WorkflowData    *WorkflowEntity  `json:"workflowData,omitempty"`
	StartNodes      []string         `json:"startNodes,omitempty"`
	DestinationNode string           `json:"destinationNode,omitempty"`
	RunData         *WorkflowRunData `json:"runData,omitempty"`
	PinData         *WorkflowPinData `json:"pinData,omitempty"`
} //@name WorkflowManualRunRequest

// IRunData is a map of node Name to array of execution task data.
type WorkflowRunData map[string][]WorkflowExecutionTaskData //@name WorkflowRunData

// IPinData is a map of node Name to array of node execution data.
type WorkflowPinData map[string][]WorkflowNodeExecutionData //@name WorkflowPinData

type WorkflowManualRunResponse struct {
	Data *WorkflowManualRunResponseData `json:"data,omitempty"`
} //@name WorkflowExecutionPushResponse

// n8n IExecutionPushResponse
type WorkflowManualRunResponseData struct {
	ExecutionId       string `json:"executionId,omitempty"`
	WaitingForWebhook bool   `json:"waitingForWebhook,omitempty"`
} //@name WorkflowManualRunResponseData

// n8n INewWorkflowData
type NewWorkflowData struct {
	Name                  string `json:"name,omitempty"`
	OnboardingFlowEnabled bool   `json:"onboardingFlowEnabled,omitempty"`
} //@name NewWorkflowData

type WorkflowNewNameResponse struct {
	Data NewWorkflowData `json:"data"`
} //@name WorkflowNewNameResponse

// n8n WorkflowEntity : IWorkflowDb : IWorkflowBase
type WorkflowEntity struct {
	ID           string                             `json:"id"`
	Name         string                             `json:"name,omitempty"`
	Active       bool                               `json:"active"` // Whether this workflow is active or not.
	Connections  map[string]WorkflowNodeConnections `json:"connections,omitempty"`
	Meta         interface{}                        `json:"meta,omitempty"`
	Nodes        []WorkflowNode                     `json:"nodes,omitempty"`
	PinData      interface{}                        `json:"pinData,omitempty"`
	Settings     *WorkflowSettings                  `json:"settings,omitempty"`
	StaticData   map[string]interface{}             `json:"staticData,omitempty"`
	Tags         []interface{}                      `json:"tags,omitempty"` //
	TriggerCount int                                `json:"triggerCount,omitempty"`
	VersionId    string                             `json:"versionId,omitempty"`
	CreatedAt    *time.Time                         `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time                         `json:"updatedAt,omitempty"`
	SugerOrgId   string                             `json:"sugerOrgId,omitempty"`
} //@name WorkflowEntity

// n8n INode
type WorkflowNode struct {
	ID                string                                    `json:"id"`
	Name              string                                    `json:"name"`
	TypeVersion       float64                                   `json:"typeVersion"`
	Type              string                                    `json:"type"`
	Position          []int64                                   `json:"position,omitempty"` // [x, y]
	Disabled          bool                                      `json:"disabled,omitempty"`
	Notes             string                                    `json:"notes,omitempty"`
	NotesInFlow       bool                                      `json:"notesInFlow,omitempty"`
	RetryOnFail       bool                                      `json:"retryOnFail,omitempty"`
	MaxTries          int64                                     `json:"maxTries,omitempty"`
	WaitBetweenTries  int64                                     `json:"waitBetweenTries,omitempty"`
	AlwaysOutputData  bool                                      `json:"alwaysOutputData,omitempty"`
	ExecutionOnce     bool                                      `json:"executionOnce,omitempty"`
	OnError           WorkflowNodeOnError                       `json:"onError,omitempty"`
	ContinueOnFail    bool                                      `json:"continueOnFail,omitempty"`
	Parameters        map[string]interface{}                    `json:"parameters,omitempty"`
	Credentials       map[string]WorkflowNodeCredentialsDetails `json:"credentials,omitempty"`
	WebhookId         string                                    `json:"webhookId,omitempty"`
	ExtendsCredential string                                    `json:"extendsCredential,omitempty"`
	SugerOrgId        string                                    `json:"sugerOrgId,omitempty"`
} //@name WorkflowNode

type ListWorkflowsResponse struct {
	Count int64            `json:"count,omitempty"`
	Data  []WorkflowEntity `json:"data"`
} //@name ListWorkflowsResponse

type ListActiveWorkflowIdsResponse struct {
	Data []string `json:"data"` // List of active workflow IDs.
} //@name ListActiveWorkflowIdsResponse

type GetWorkflowResponse struct {
	Data *WorkflowEntity `json:"data,omitempty"`
} //@name GetWorkflowResponse

type UpdateWorkflowResponse struct {
	Data *WorkflowEntity `json:"data,omitempty"`
} //@name UpdateWorkflowResponse

type DeleteWorkflowResponse struct {
	Data bool `json:"data"`
} //@name DeleteWorkflowResponse

type WorkflowNodeCredentialsDetails struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
} //@name WorkflowNodeCredentialsDetails

type WorkflowNodeConnections map[string]WorkflowNodeInputConnections //@name WorkflowNodeConnections

type WorkflowNodeInputConnections [][]WorkflowConnection //@name WorkflowNodeInputConnections

type WorkflowConnection struct {
	// The node the connection is to
	Node string `json:"node"`
	// The type of the input on destination node (for example "main")
	Type string `json:"type"`
	// The output/input-index of destination node (if node has multiple inputs/outputs of the same type)
	Index int64 `json:"index"`
} //@name WorkflowConnection

type WorkflowSettings struct {
	Timezone                 string                    `json:"timezone,omitempty"`
	ErrorWorkflow            string                    `json:"errorWorkflow,omitempty"`
	CallerIds                string                    `json:"callerIds,omitempty"`
	CallerPolicy             WorkflowCallerPolicy      `json:"callerPolicy,omitempty"`
	SaveDataErrorExecution   WorkflowSaveDataExecution `json:"saveDataErrorExecution,omitempty"`
	SaveDataSuccessExecution WorkflowSaveDataExecution `json:"saveDataSuccessExecution,omitempty"`
	SaveManualExecutions     interface{}               `json:"saveManualExecutions,omitempty"`
	SaveExecutionProgress    interface{}               `json:"saveExecutionProgress,omitempty"`
	ExecutionTimeout         int64                     `json:"executionTimeout,omitempty"`
	ExecutionOrder           string                    `json:"executionOrder,omitempty"`
	SugerOrgId               string                    `json:"sugerOrgId,omitempty"`
} //@name WorkflowSettings

type WorkflowNodeExecutionError struct {
	Name        string `json:"name,omitempty"`
	Message     string `json:"message,omitempty"`
	Description string `json:"description,omitempty"`
} //@name WorkflowNodeExecutionError

type WorkflowNodeExecutionResult struct {
	ExecutionStatus WorkflowExecutionStatus      `json:"executionStatus,omitempty"`
	Errors          []WorkflowNodeExecutionError `json:"errors,omitempty"`
} //@name WorkflowNodeExecutionResult

type WorkflowExecutionError struct {
	Cause         interface{}            `json:"cause,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Description   string                 `json:"description,omitempty"`
	ErrorResponse interface{}            `json:"errorResponse,omitempty"`
	Functionality string                 `json:"functionality,omitempty" enum:"regular,configuration-node"`
	LineNumber    *int64                 `json:"lineNumber,omitempty"`
	Message       string                 `json:"message,omitempty"`
	Timestamp     int64                  `json:"timestamp,omitempty"`
	WorkflowId    string                 `json:"workflowId,omitempty"`
	Node          *WorkflowNode          `json:"node,omitempty"`
} //@name WorkflowExecutionError

type WorkflowExecutionSummary struct {
	Id                  string                                 `json:"id,omitempty"`
	ExecutionError      *WorkflowExecutionError                `json:"executionError,omitempty"`
	Finished            bool                                   `json:"finished,omitempty"`
	LastNodeExecuted    string                                 `json:"lastNodeExecuted,omitempty"`
	Mode                WorkflowExecutionMode                  `json:"mode,omitempty"`
	NodeExecutionStatus map[string]WorkflowNodeExecutionResult `json:"nodeExecutionStatus,omitempty"`
	RetryOf             string                                 `json:"retryOf,omitempty"`
	RetrySuccessId      string                                 `json:"retrySuccessId,omitempty"`
	Status              WorkflowExecutionStatus                `json:"status,omitempty"`
	StartedAt           *time.Time                             `json:"startedAt,omitempty"`
	StoppedAt           *time.Time                             `json:"stoppedAt,omitempty"`
	WaitTill            *time.Time                             `json:"waitTill,omitempty"`
	WorkflowId          string                                 `json:"workflowId,omitempty"`
	WorkflowName        string                                 `json:"workflowName,omitempty"`
} //@name WorkflowExecutionSummary

type ListWorkflowExecutionsResponse struct {
	Data *ListWorkflowExecutionsResponseData `json:"data,omitempty"`
} //@name ListWorkflowExecutionsResponse

type ListWorkflowExecutionsResponseData struct {
	Count     int64                      `json:"count,omitempty"`
	Results   []WorkflowExecutionSummary `json:"results,omitempty"`
	Estimated bool                       `json:"estimated,omitempty"`
} //@name ListWorkflowExecutionsResponseData

type WorkflowExecutionsQueryFilter struct {
	ID             string                    `json:"id,omitempty"`
	Finished       bool                      `json:"finished,omitempty"`
	Mode           string                    `json:"mode,omitempty"`
	RetryOf        string                    `json:"retryOf,omitempty"`
	RetrySuccessId string                    `json:"retrySuccessId,omitempty"`
	Status         []WorkflowExecutionStatus `json:"status,omitempty"`
	WorkflowId     string                    `json:"workflowId,omitempty"`
	Metadata       []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"metadata,omitempty"`
	StartedAfter  *time.Time `json:"startedAfter,omitempty"`
	StartedBefore *time.Time `json:"startedBefore,omitempty"`
} //@name WorkflowExecutionsQueryFilter

type ListWorkflowCurrentExecutionsResponse struct {
	Data []WorkflowExecutionSummary `json:"data"`
} //@name ListWorkflowCurrentExecutionsResponse

// n8n IExecutingWorkflowData representing an active execution
type ExecutingWorkflowData struct {
	ExecutionData *WorkflowExecutionDataProcess
	StartedAt     *time.Time
	Status        WorkflowExecutionStatus
	// postExecutePromises
	// responsePromise
	// n8n field name is workflowExecution (too common), so we use a different name
	WorkflowExecutionRun *WorkflowExecutionCancelableRun
}

// details of an active execution
type WorkflowExecutionDataProcess struct {
	DestinationNode    string // optional
	RestartExecutionId string // optional
	ExecutionMode      WorkflowExecutionMode
	ExecutionData      *WorkflowRunExecutionData // optional
	RunData            *WorkflowRunData          // optional
	PinData            *WorkflowPinData          // optional
	RetryOf            string                    // optional
	SessionId          string                    // optional
	StartNodes         []string                  // optional
	WorkflowData       *WorkflowEntity
	UserId             string
}

// n8n PCancelable<IRun>
type WorkflowExecutionCancelableRun struct {
	Ctx          context.Context
	Cancel       context.CancelFunc
	WaitErrChan  chan error
	WaitDataChan chan *WorkflowRunExecutionData
	Finished     bool
	Data         *WorkflowRunExecutionData
	Mode         WorkflowExecutionMode
	StartedAt    *time.Time
	StoppedAt    *time.Time
	Status       WorkflowExecutionStatus
	// WailTill?
}

func (c *WorkflowExecutionCancelableRun) Wait(requestCtx context.Context, responseSendChan chan struct{}) (*WorkflowRunExecutionData, error) {
	select {
	// Request Context wwas canceled
	case <-requestCtx.Done():
		return nil, requestCtx.Err()
	// Execution was canceled
	case <-c.Ctx.Done():
		return nil, c.Ctx.Err()
	// execution end with error
	case err := <-c.WaitErrChan:
		return nil, err
	// execution end with result
	case data := <-c.WaitDataChan:
		return data, nil
	// response has been send
	case <-responseSendChan:
		return nil, nil
	}
}

// n8n IRunExecutionData
type WorkflowRunExecutionData struct {
	StartData     *WorkflowRunExecutionStartData     `json:"startData,omitempty"`
	ResultData    *WorkflowRunExecutionResultData    `json:"resultData,omitempty"`
	ExecutionData *WorkflowRunExecutionExecutionData `json:"executionData,omitempty"`
	WaitTill      *time.Time                         `json:"waitTill,omitempty"`
} //@name WorkflowRunExecutionData

// n8n anonymous type
type WorkflowRunExecutionStartData struct {
	DestinationNode string   `json:"destinationNode,omitempty"`
	RunNodeFilter   []string `json:"runNodeFilter,omitempty"`
} //@name WorkflowRunExecutionStartData

// n8n anonymous type
type WorkflowRunExecutionResultData struct {
	Error            string                                  `json:"error,omitempty"`
	RunData          map[string][]*WorkflowExecutionTaskData `json:"runData,omitempty"`
	PinData          interface{}                             `json:"pinData,omitempty"`
	LastNodeExecuted string                                  `json:"lastNodeExecuted,omitempty"`
	MetaData         map[string]string                       `json:"metaData,omitempty"`
} //@name WorkflowRunExecutionResultData

// n8n anonymous type
type WorkflowRunExecutionExecutionData struct {
	NodeExecutionStack     *NodeExecutionStack              `json:"nodeExecutionStack,omitempty"`
	WaitingExecution       map[string][]NodeData            `json:"waitingExecution,omitempty"`
	WaitingExecutionSource map[string][]ExecutionSourceData `json:"waitingExecutionSource,omitempty"`
}

type ExecutionSourceData struct {
	PreviousNode       string
	PreviousNodeOutput int
}

type Run struct {
	Data       *WorkflowRunExecutionData `json:"data"`
	Finished   bool
	Mode       WorkflowExecutionMode
	WaitTill   *time.Time
	StartedAt  *time.Time
	StoppedAt  *time.Time
	Status     WorkflowExecutionStatus
	NeedDelete bool
}

// This is the old version of WorkflowExecution
// Will be removed once the new version is fully rolled out.
type WorkflowExecution_Old struct {
	Id             string                  `json:"id,omitempty"`
	Data           interface{}             `json:"data,omitempty"`
	Finished       bool                    `json:"finished,omitempty"`
	Mode           WorkflowExecutionMode   `json:"mode,omitempty"`
	Status         WorkflowExecutionStatus `json:"status,omitempty"`
	RetryOf        string                  `json:"retryOf,omitempty"`
	RetrySuccessId string                  `json:"retrySuccessId,omitempty"`
	StartedAt      *time.Time              `json:"startedAt,omitempty"`
	StoppedAt      *time.Time              `json:"stoppedAt,omitempty"`
	WaitTill       *time.Time              `json:"waitTill,omitempty"`
	WorkflowData   *WorkflowEntity         `json:"workflowData,omitempty"`
	WorkflowId     string                  `json:"workflowId,omitempty"`
} //@name WorkflowExecution_Old

type WorkflowExecution struct {
	Id             string                    `json:"id,omitempty"`
	Data           *WorkflowRunExecutionData `json:"data,omitempty"`
	Finished       bool                      `json:"finished,omitempty"`
	Mode           WorkflowExecutionMode     `json:"mode,omitempty"`
	Status         WorkflowExecutionStatus   `json:"status,omitempty"`
	RetryOf        string                    `json:"retryOf,omitempty"`
	RetrySuccessId string                    `json:"retrySuccessId,omitempty"`
	StartedAt      *time.Time                `json:"startedAt,omitempty"`
	StoppedAt      *time.Time                `json:"stoppedAt,omitempty"`
	WaitTill       *time.Time                `json:"waitTill,omitempty"`
	WorkflowData   *WorkflowEntity           `json:"workflowData,omitempty"`
	WorkflowId     string                    `json:"workflowId,omitempty"`
} //@name WorkflowExecution

type GetWorkflowExecutionResponse struct {
	Data *WorkflowExecution `json:"data"`
} //@name GetWorkflowExecutionResponse

// This is the old version of GetWorkflowExecutionResponse
// Will be removed once the new version is fully rolled out.
type GetWorkflowExecutionResponse_Old struct {
	Data *WorkflowExecution_Old `json:"data"`
} //@name GetWorkflowExecutionResponse_Old

type RetryWorkflowExecutionRequest struct {
	LoadWorkflow bool `json:"loadWorkflow,omitempty"`
} //@name RetryWorkflowExecutionRequest

type RetryWorkflowExecutionResponse struct {
	Data bool `json:"data"`
} //@name RetryWorkflowExecutionResponse

type WorkflowExecutionStopData struct {
	Finished  bool                    `json:"finished,omitempty"`
	Mode      WorkflowExecutionMode   `json:"mode,omitempty"`
	StartedAt *time.Time              `json:"startedAt,omitempty"`
	StoppedAt *time.Time              `json:"stoppedAt,omitempty"`
	Status    WorkflowExecutionStatus `json:"status,omitempty"`
} //@name WorkflowExecutionStopData

type StopWorkflowExecutionResponse struct {
	Data *WorkflowExecutionStopData `json:"data"`
} //@name StopWorkflowExecutionResponse

type DeleteWorkflowExecutionsRequest struct {
	DeleteBefore *time.Time             `json:"deleteBefore,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	Ids          []string               `json:"ids,omitempty"`
} //@name DeleteWorkflowExecutionsRequest

type DeleteWorkflowTestWebhookResponse struct {
	Data bool `json:"data"`
} //@name DeleteWorkflowTestWebhookResponse

type ListAllWorkflowNodesResposne []interface{} //@name ListAllWorkflowNodesResposne

type GetWorkflowDynamicNodeParametersRequest struct {
	MethodName string `json:"methodName"`
	Path       string `json:"path"`
	SugerOrgId string `json:"sugerOrgId"`

	NodeTypeAndVersion    WorkflowNodeTypeAndVersion `json:"nodeTypeAndVersion"`
	CurrentNodeParameters map[string]interface{}     `json:"currentNodeParameters"`

	Filter          string `json:"filter,omitempty"`
	LoadOptions     string `json:"loadOptions,omitempty"`
	PaginationToken string `json:"paginationToken,omitempty"`
} //@name GetWorkflowDynamicNodeParametersRequest

type WorkflowNodeTypeAndVersion struct {
	Name    string  `json:"name"`
	Version float64 `json:"version"`
} //@name WorkflowNodeTypeAndVersion

type GetDynamicNodeParametersResponse_ResourceLocatorResults struct {
	Data *WorkflowNodeListSearchResult `json:"data"`
} //@name GetDynamicNodeParametersResponse_ResourceLocatorResults

type WorkflowNodeListSearchResult struct {
	Results         []WorkflowNodeListSearchItem `json:"results"`
	PaginationToken string                       `json:"paginationToken,omitempty"`
} //@name WorkflowNodeListSearchResult

type WorkflowNodeListSearchItem struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"` // can be string, int64, float64 or bool
	Action      string      `json:"action,omitempty"`
	Description string      `json:"description,omitempty"`
	Icon        string      `json:"icon,omitempty"`
	Routing     interface{} `json:"routing,omitempty"`
	Url         string      `json:"url,omitempty"`
} //@name WorkflowNodeListSearchItem

type GetDynamicNodeParametersResponse_ResourceMapperFields struct {
	Data *WorkflowResourceMapperFields `json:"data"`
} //@name GetDynamicNodeParametersResponse_ResourceMapperFields

type WorkflowResourceMapperFields struct {
	Fields []WorkflowResourceMapperField `json:"fields"`
} //@name WorkflowResourceMapperFields

type WorkflowResourceMapperField struct {
	ID               string `json:"id"`
	DisplayName      string `json:"displayName"`
	DefaultMatch     bool   `json:"defaultMatch"`
	CanBeUsedToMatch bool   `json:"canBeUsedToMatch,omitempty"`

	Required bool                          `json:"required"`
	Display  bool                          `json:"display"`
	Type     string                        `json:"type,omitempty"`
	Removed  bool                          `json:"removed,omitempty"`
	Options  []WorkflowNodePropertyOptions `json:"options,omitempty"`
	ReadOnly bool                          `json:"readOnly,omitempty"`
} //@name WorkflowResourceMapperField

type WorkflowNodePropertyOptions struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"` // can be string, int64, float64 or bool
	Action      string      `json:"action,omitempty"`
	Description string      `json:"description,omitempty"`
	Routing     interface{} `json:"routing,omitempty"`
} //@name WorkflowNodePropertyOptions

type GetDynamicNodeParametersResponse_Options struct {
	Data []WorkflowNodePropertyOptions `json:"data"`
} //@name GetDynamicNodeParametersResponse_Options

type WorkflowActivateMode string //@name WorkflowActivateMode

const (
	WorkflowActivateMode_Init            WorkflowActivateMode = "init"
	WorflowActivateMode_Create           WorkflowActivateMode = "create"
	WorflowActivateMode_Update           WorkflowActivateMode = "update"
	WorflowActivateMode_Activate         WorkflowActivateMode = "activate"
	WorflowActivateMode_Manual           WorkflowActivateMode = "manual"
	WorflowActivateMode_LeadershipChange WorkflowActivateMode = "leadershipChange"
)

type WebhookResponseMode string //@name WebhookResponseMode

const (
	// WebhookResponseMode_OnReceived is default
	WebhookResponseMode_OnReceived   WebhookResponseMode = "onReceived"
	WebhookResponseMode_LastNode     WebhookResponseMode = "lastNode"
	WebhookResponseMode_ResponseNode WebhookResponseMode = "responseNode"
)

type WebhookRespondWith string //@name WebhookRespondWith

const (
	// WebhookRespondWith_FirstIncomingItem is default
	WebhookRespondWith_FirstIncomingItem WebhookRespondWith = "firstIncomingItem"
	WebhookRespondWith_AllIncomingItems  WebhookRespondWith = "allIncomingItems"
	WebhookRespondWith_Binary            WebhookRespondWith = "binary"
	WebhookRespondWith_Json              WebhookRespondWith = "json"
	WebhookRespondWith_NoData            WebhookRespondWith = "noData"
	WebhookRespondWith_Redirect          WebhookRespondWith = "redirect"
	WebhookRespondWith_Text              WebhookRespondWith = "text"
)

type WebhookResponseData string //@name WebhookResponseData

const (
	// WebhookResponseData_FirstEntryJson is default
	WebhookResponseData_FirstEntryJson   WebhookResponseData = "firstEntryJson"
	WebhookResponseData_FirstEntryBinary WebhookResponseData = "firstEntryBinary"
	WebhookResponseData_AllEntries       WebhookResponseData = "allEntries"
	WebhookResponseData_NoData           WebhookResponseData = "noData"
)

type (
	// Spec is the universal spec for all nodes.
	WorkflowNodeSpec struct {
		category NodeObjectCategory

		JsonConfig []byte
		NodeSpec   *WorkflowNodeDescriptionSpec
	}

	// DescriptionSpec is the spec for description node.
	WorkflowNodeDescriptionSpec struct {
		ActivationMessage       string                   `json:"activationMessage,omitempty"`
		BadgeIconUrl            string                   `json:"badgeIconUrl,omitempty"`
		Codex                   *WorkflowNodeCodexSpec   `json:"codex,omitempty"`
		Credentials             []DescriptionCredentials `json:"credentials,omitempty"`
		Defaults                DescriptionDefaults      `json:"defaults"`
		DefaultVersion          interface{}              `json:"defaultVersion"`
		Description             string                   `json:"description,omitempty"`
		DisplayName             string                   `json:"displayName"`
		DocumentationUrl        string                   `json:"documentationUrl,omitempty"`
		EventTriggerDescription string                   `json:"eventTriggerDescription,omitempty"`
		Group                   []string                 `json:"group,omitempty"`
		Icon                    string                   `json:"icon,omitempty"`
		IconUrl                 string                   `json:"iconUrl,omitempty"`
		InputNames              []string                 `json:"inputNames,omitempty"`
		Inputs                  []string                 `json:"inputs"`
		MaxNodes                int                      `json:"maxNodes,omitempty"`
		MockManualExecution     bool                     `json:"mockManualExecution,omitempty"`
		Name                    string                   `json:"name"`
		Outputs                 interface{}              `json:"outputs"` // []string or string
		OutputNames             []string                 `json:"outputNames,omitempty"`
		Polling                 bool                     `json:"polling,omitempty"`
		Properties              []DescriptionProperties  `json:"properties"`
		RequiredInputs          interface{}              `json:"requiredInputs,omitempty"`
		Subtitle                string                   `json:"subtitle,omitempty"`
		SupportsCORS            bool                     `json:"supportsCORS,omitempty"`
		TriggerPanel            DescriptionTriggerPanel  `json:"triggerPanel,omitempty"`
		Version                 interface{}              `json:"version"` // string or int
		Webhooks                []WebhookDescription     `json:"webhooks,omitempty"`
	}

	WorkflowNodeCodexSpec struct {
		Categories []string    `json:"categories,omitempty"`
		Resources  interface{} `json:"resources,omitempty"`
	}

	DescriptionTriggerPanel struct {
		Header         string                    `json:"header,omitempty"`
		ExecutionsHelp DescriptionExecutionsHelp `json:"executionsHelp,omitempty"`
		ActivationHint interface{}               `json:"activationHint,omitempty"`
	}

	DescriptionExecutionsHelp struct {
		Active   string `json:"active,omitempty"`
		Inactive string `json:"inactive,omitempty"`
	}

	WebhookDescription struct {
		HttpMethod                 string `json:"httpMethod"`
		IsFullPath                 bool   `json:"isFullPath,omitempty"`
		Name                       string `json:"name"`
		Path                       string `json:"path"`
		ResponseBinaryPropertyName string `json:"responseBinaryPropertyName,omitempty"`
		ResponseContentType        string `json:"responseContentType,omitempty"`
		ResponsePropertyName       string `json:"responsePropertyName,omitempty"`
		ResponseMode               string `json:"responseMode,omitempty"`
		ResponseData               string `json:"responseData,omitempty"`
		RestartWebhook             bool   `json:"restartWebhook,omitempty"`
		IsForm                     bool   `json:"isForm,omitempty"`
		HasLifecycleMethod         bool   `json:"hasLifecycleMethod,omitempty"`
		NdvHideUrl                 bool   `json:"ndvHideUrl,omitempty"`
		NdvHideMethod              bool   `json:"ndvHideMethod,omitempty"`
	}

	// DescriptionDefaults is the defaults for description field.
	DescriptionDefaults struct {
		Name  string `json:"name,omitempty"`
		Color string `json:"color,omitempty"`
	}

	DescriptionProperties struct {
		DisplayName                     string                    `json:"displayName"`
		Name                            string                    `json:"name"`
		Type                            string                    `json:"type"`
		TypeOptions                     DescriptionTypeOptions    `json:"typeOptions,omitempty"`
		Default                         interface{}               `json:"default"`
		Description                     string                    `json:"description,omitempty"`
		Hint                            string                    `json:"hint,omitempty"`
		DisplayOptions                  DescriptionDisplayOptions `json:"displayOptions,omitempty"`
		Options                         []interface{}             `json:"options,omitempty"`
		Placeholder                     string                    `json:"placeholder,omitempty"`
		IsNodeSetting                   bool                      `json:"isNodeSetting,omitempty"`
		NoDataExpression                bool                      `json:"noDataExpression,omitempty"`
		Required                        bool                      `json:"required,omitempty"`
		Routing                         DescriptionRouting        `json:"routing,omitempty"`
		CredentialTypes                 []string                  `json:"credentialTypes,omitempty"`
		ExtractValue                    DescriptionExtractValue   `json:"extractValue,omitempty"`
		Modes                           []DescriptionModes        `json:"modes,omitempty"`
		RequiresDataPath                string                    `json:"requiresDataPath,omitempty"`
		DoNotInherit                    bool                      `json:"doNotInherit,omitempty"`
		ValidateType                    string                    `json:"validateType,omitempty"`
		IgnoreValidationDuringExecution bool                      `json:"ignoreValidationDuringExecution,omitempty"`
	}

	DescriptionCredentials struct {
		Name           string                               `json:"name"`
		Required       bool                                 `json:"required"`
		DisplayOptions DescriptionCredentialsDisplayOptions `json:"displayOptions,omitempty"`
	}

	DescriptionCredentialsDisplayOptions struct {
		Show DescriptionCredentialsShow `json:"show"`
	}

	DescriptionCredentialsShow struct {
		Authentication []string `json:"authentication"`
	}

	DescriptionModes struct {
		DisplayName  string                       `json:"displayName"`
		Name         string                       `json:"name"`
		Type         string                       `json:"type"`
		Hint         string                       `json:"hint,omitempty"`
		Validation   []DescriptionModesValidation `json:"validation,omitempty"`
		Placeholder  string                       `json:"placeholder,omitempty"`
		URL          string                       `json:"url,omitempty"`
		ExtractValue DescriptionExtractValue      `json:"extractValue,omitempty"`
		InitType     string                       `json:"initType,omitempty"`
		TypeOptions  map[string]interface{}       `json:"typeOptions,omitempty"`
	}

	DescriptionModesValidation struct {
		Type       string                               `json:"type"`
		Properties DescriptionModesValidationProperties `json:"properties"`
	}

	DescriptionModesValidationProperties struct {
		Regex        string `json:"regex"`
		ErrorMessage string `json:"errorMessage"`
	}

	DescriptionRouting struct {
		Request DescriptionRoutingRequest `json:"request"`
	}

	DescriptionRoutingRequest struct {
		URL                          string                         `json:"url"`
		BaseURL                      string                         `json:"baseURL,omitempty"`
		Headers                      map[string]string              `json:"headers,omitempty"`
		Method                       string                         `json:"method"`
		Body                         interface{}                    `json:"body,omitempty"`
		QS                           interface{}                    `json:"qs,omitempty"`
		ArrayFormat                  string                         `json:"arrayFormat,omitempty"`
		Auth                         DescriptionRoutingRequestAuth  `json:"auth,omitempty"`
		DisableFollowRedirect        bool                           `json:"disableFollowRedirect,omitempty"`
		Encoding                     string                         `json:"encoding,omitempty"`
		SkipSslCertificateValidation bool                           `json:"skipSslCertificateValidation,omitempty"`
		ReturnFullResponse           bool                           `json:"returnFullResponse,omitempty"`
		IgnoreHttpStatusErrors       bool                           `json:"ignoreHttpStatusErrors,omitempty"`
		Proxy                        DescriptionRoutingRequestProxy `json:"proxy,omitempty"`
		Timeout                      int                            `json:"timeout,omitempty"`
		Json                         bool                           `json:"json,omitempty"`
	}

	DescriptionRoutingRequestAuth struct {
		UserName        string `json:"userName"`
		Password        string `json:"password"`
		SendImmediately bool   `json:"sendImmediately,omitempty"`
	}

	DescriptionRoutingRequestProxy struct {
		Host     string                        `json:"host"`
		Port     int                           `json:"port"`
		Auth     DescriptionRoutingRequestAuth `json:"auth,omitempty"`
		Protocol string                        `json:"protocol,omitempty"`
	}

	DescriptionExtractValue struct {
		Type  string `json:"type"`
		Regex string `json:"regex"`
	}

	DescriptionDisplayOptions struct {
		Show DescriptionDisplayOptionsShow `json:"show"`
		Hide DescriptionDisplayOptionsShow `json:"hide"`
	}

	DescriptionDisplayOptionsShow map[string][]interface{}

	DescriptionTypeOptions struct {
		Action                  string                    `json:"action,omitempty"`
		ContainerClass          string                    `json:"containerClass,omitempty"`
		AlwaysOpenEditWindow    bool                      `json:"alwaysOpenEditWindow,omitempty"`
		CodeAutocomplete        string                    `json:"codeAutocomplete,omitempty"`
		Editor                  string                    `json:"editor,omitempty"`
		EditorLanguage          string                    `json:"editorLanguage,omitempty"`
		SqlDialect              string                    `json:"sqlDialect,omitempty"`
		LoadOptionsDependsOn    []string                  `json:"loadOptionsDependsOn,omitempty"`
		LoadOptionsMethod       string                    `json:"loadOptionsMethod,omitempty"`
		MaxValue                int                       `json:"maxValue,omitempty"`
		MinValue                int                       `json:"minValue,omitempty"`
		MultipleValues          bool                      `json:"multipleValues,omitempty"`
		MultipleValueButtonText string                    `json:"multipleValueButtonText,omitempty"`
		NumberPrecision         int                       `json:"numberPrecision,omitempty"`
		NumberStepSize          int                       `json:"numberStepSize,omitempty"`
		Password                bool                      `json:"password,omitempty"`
		Rows                    int                       `json:"rows,omitempty"`
		ShowAlpha               bool                      `json:"showAlpha,omitempty"`
		Sortable                bool                      `json:"sortable,omitempty"`
		Expirable               bool                      `json:"expirable,omitempty"`
		ResourceMapper          DescriptionResourceMapper `json:"resourceMapper,omitempty"`
		Filter                  DescriptionFilter         `json:"filter,omitempty"`
	}

	DescriptionResourceMapper struct {
		ResourceMapperMethod string                `json:"resourceMapperMethod"`
		Mode                 string                `json:"mode"`
		ValuesLabel          string                `json:"valuesLabel,omitempty"`
		FieldWords           DescriptionFieldWords `json:"fieldWords,omitempty"`
		AddAllFields         bool                  `json:"addAllFields,omitempty"`
		NoFieldsError        string                `json:"noFieldsError,omitempty"`
		MultiKeyMatch        bool                  `json:"multiKeyMatch,omitempty"`
		SupportAutoMap       bool                  `json:"supportAutoMap,omitempty"`
	}

	DescriptionFieldWords struct {
		Singular string `json:"singular"`
		Plural   string `json:"plural"`
	}

	DescriptionFilter struct {
		CaseSensitive      string   `json:"caseSensitive"`
		LeftValue          string   `json:"leftValue"`
		AllowedCombinators []string `json:"allowedCombinators"`
		MaxConditions      int      `json:"maxConditions"`
		TypeValidation     string   `json:"typeValidation"`
	}
)

func (s *WorkflowNodeSpec) Name() string {
	return s.NodeSpec.Name
}

func (s *WorkflowNodeSpec) GenerateSpec() *WorkflowNodeSpec {
	// Unmarshal raw to nodeSpec
	err := json.Unmarshal(s.JsonConfig, &s.NodeSpec)
	if err != nil {
		fmt.Println("Failed to unmarshal json config")
	}
	return s
}

const (
	// CategoryTrigger is the category of trigger nodes.
	CategoryTrigger = "Trigger"

	// CategoryExecutor is the category of executor nodes.
	CategoryExecutor = "Executor"
)

type (
	SetExecutionStatus func(status WorkflowExecutionStatus) error

	WorkflowExecutionData struct {
		Node *WorkflowNode
	}

	WorkflowExecuteHooks struct {
		NodeExecuteAfter []func(
			ctx context.Context,
			hooks *WorkflowHooks,
			nodeName string,
			result *NodeExecutionResult,
			taskData *WorkflowExecutionTaskData,
			executionData *WorkflowRunExecutionData,
		)
		NodeExecuteBefore []func(
			ctx context.Context,
			hooks *WorkflowHooks,
			nodeName string,
		)
		WorkflowExecuteAfter []func(
			ctx context.Context,
			hooks *WorkflowHooks,
			fullRunData *Run,
		)
		WorkflowExecuteBefore []func(
			ctx context.Context,
			hooks *WorkflowHooks,
			workflowReq *WorkflowEntity,
		)
		SendResponse []func(
			ctx context.Context,
			hooks *WorkflowHooks,
			response *fasthttp.Response,
		)
	}

	WorkflowHooks struct {
		Mode          WorkflowExecutionMode
		WorkflowData  *WorkflowEntity
		ExecutionId   string
		SessionId     string
		RetryOf       string
		HookFunctions WorkflowExecuteHooks
	}

	WorkflowExecuteAdditionalData struct {
		ExecutionId               string
		RestartExecutionId        string
		HttpResponse              *fasthttp.Response
		HttpRequest               *fasthttp.Request
		RestApiUrl                string             // const from os.env
		InstanceBaseUrl           string             // const from os.env
		CbSetExecutionStatus      SetExecutionStatus // CBFunc
		FormWaitingBaseUrl        string
		WebhookBaseUrl            string //const from os.env
		WebhookWaitingBaseUrl     string //const from os.env
		WebhookTestBaseUrl        string //const
		CurrentNodeParameters     map[string]interface{}
		ExecutionTimeoutTimestamp time.Time
		UserId                    string
		Variables                 interface{}
		Hooks                     WorkflowHooks
	}

	WorkflowHookList struct {
		SendResponse func(response *fasthttp.Response)
	}

	NodeData       []map[string]interface{}
	NodeSingleData map[string]interface{}

	NodeExecutionResult struct {
		ExecutionStatus WorkflowExecutionStatus      `json:"executionStatus,omitempty"`
		Errors          []WorkflowNodeExecutionError `json:"errors,omitempty"`
		NextNodeIndex   []int                        `json:"next_node_index,omitempty"`
		TriggerData     NodeData                     `json:"data,omitempty"`
		ExecutorData    []NodeData                   `json:"multipleData,omitempty"`
	}

	NodeExecutionResultDetail struct {
		NodeName string               `json:"nodeName,omitempty"`
		Category NodeObjectCategory   `json:"category,omitempty"`
		Result   *NodeExecutionResult `json:"result,omitempty"`
	}

	NodeExecuteInput struct {
		WorkflowID           string
		Params               *WorkflowNode
		Data                 []NodeData
		ExecutionData        *WorkflowExecutionData
		RunExecutionData     *WorkflowRunExecutionData
		RunIndex             int32
		AdditionalData       *WorkflowExecuteAdditionalData
		NodeExecuteFunctions interface{}
		Mode                 WorkflowExecutionMode
		ActivateMode         WorkflowActivateMode
	}

	NodeTriggerInput struct {
		Node *WorkflowNode
	}

	WorkflowRunnerChannel chan struct{}

	// NodeObjectCategory is the type to classify all objects.
	NodeObjectCategory string

	CloseTrigger func()

	// DynamicParameter Methods

	NodeMethodListSearch func(ctx context.Context, sugerOrgId string, nodeParameters map[string]interface{}, filter string, paginationToken string) (*GetDynamicNodeParametersResponse_ResourceLocatorResults, error)

	NodeMethodLoadOptions func(ctx context.Context, sugerOrgId string, nodeParameters map[string]interface{}) (*GetDynamicNodeParametersResponse_Options, error)

	NodeMethodResourceMapping func(ctx context.Context, sugerOrgId string, nodeParameters map[string]interface{}) (*GetDynamicNodeParametersResponse_ResourceMapperFields, error)

	NodeMethods struct {
		// The key of map is method name
		LoadOptions map[string]NodeMethodLoadOptions
		// The key of map is method name
		ListSearch map[string]NodeMethodListSearch
		// The key of map is method name
		ResourceMapping map[string]NodeMethodResourceMapping
	}

	// Node Default Webhook Methods
	NodeWebhookMethods struct {
		CheckExists func(context.Context, *WorkflowEntity, *WorkflowNode, *WebhookData) (bool, error)
		Create      func(context.Context, *WorkflowEntity, *WorkflowNode, *WebhookData) (bool, error)
		Delete      func(context.Context, *WorkflowEntity, *WorkflowNode, *WebhookData) (bool, error)
	}

	WebhookData struct {
		HttpMethod                      string                        `json:"httpMethod"`
		Node                            string                        `json:"node"` // The node name
		NodeType                        string                        `json:"nodeType"`
		NodeId                          string                        `json:"nodeId"`
		Path                            string                        `json:"path"`
		WebhookDescription              WebhookDescription            `json:"webhookDescription"`
		WorkflowId                      string                        `json:"workflowId"`
		WorkflowExecutionAdditionalData WorkflowExecuteAdditionalData `json:"workflowExecutionAdditionalData"`
		WebhookId                       string                        `json:"webhookId"`
		IsTest                          bool                          `json:"isTest"`
	}

	// RegisteredWebhook is the registered webhook.
	RegisteredWebhook struct {
		WorkflowEntity  *WorkflowEntity
		DestinationNode string // no need to use this
	}

	NodeExecutionStackData struct {
		Node          *WorkflowNode `json:"node"`
		RunResultList []NodeData    `json:"runResultList"`
	}

	NodeExecutionStack struct {
		Nodes *list.List `json:"nodes"`
	}
)

func NewNodeExecStack(triggerList []*WorkflowNode) *NodeExecutionStack {
	nodeExecutionStack := list.New()
	for _, triggerNode := range triggerList {
		nodeExecutionStack.PushBack(&NodeExecutionStackData{
			Node:          triggerNode,
			RunResultList: []NodeData{},
		})
	}
	return &NodeExecutionStack{nodeExecutionStack}
}

func (n *NodeExecutionStack) PushBack(stackData *NodeExecutionStackData) {
	n.Nodes.PushBack(stackData)
}

func (n *NodeExecutionStack) PushFront(stackData *NodeExecutionStackData) {
	n.Nodes.PushFront(stackData)
}

func (n *NodeExecutionStack) PopFront() *NodeExecutionStackData {
	ele := n.Nodes.Remove(n.Nodes.Front())
	if ele == nil {
		return nil
	}
	fNode := ele.(*NodeExecutionStackData)
	return fNode
}

func (n *NodeExecutionStack) MarshalJSON() ([]byte, error) {
	ret := make([]*NodeExecutionStackData, 0, n.Nodes.Len())
	for e := n.Nodes.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(*NodeExecutionStackData))
	}
	return json.Marshal(ret)
}

func (n *NodeExecutionStack) UnmarshalJSON(data []byte) error {
	var nodes []*NodeExecutionStackData
	err := json.Unmarshal(data, &nodes)
	if err != nil {
		return err
	}
	n.Nodes = list.New()
	for _, node := range nodes {
		n.Nodes.PushBack(node)
	}
	return nil
}

func (hooks *WorkflowHooks) ExecutionHookFunctionsNodeExecutionAfter(
	ctx context.Context,
	nodeName string,
	result *NodeExecutionResult,
	taskData *WorkflowExecutionTaskData,
	executionData *WorkflowRunExecutionData,
) {
	for _, h := range hooks.HookFunctions.NodeExecuteAfter {
		h(ctx, hooks, nodeName, result, taskData, executionData)
	}
}

func (hooks *WorkflowHooks) ExecutionHookFunctionsWorkflowExecutionAfter(ctx context.Context, fullRunData *Run) {
	for _, h := range hooks.HookFunctions.WorkflowExecuteAfter {
		h(ctx, hooks, fullRunData)
	}
}

func (hooks *WorkflowHooks) ExecutionHookFunctionsNodeExecutionBefore(ctx context.Context, nodeName string) {
	for _, h := range hooks.HookFunctions.NodeExecuteBefore {
		h(ctx, hooks, nodeName)
	}
}

func (hooks *WorkflowHooks) ExecutionHookFunctionsWorkflowExecuteBefore(ctx context.Context, workflowEntity *WorkflowEntity) {
	for _, h := range hooks.HookFunctions.WorkflowExecuteBefore {
		h(ctx, hooks, workflowEntity)
	}
}

func (hooks *WorkflowHooks) ExecutionHookFunctionsSendResponse(ctx context.Context, response *fasthttp.Response) {
	for _, h := range hooks.HookFunctions.SendResponse {
		h(ctx, hooks, response)
	}
}

type WorkflowWebhook struct {
	HookName   string `json:"name"`
	PublicUrl  string `json:"publicUrl,omitempty"`
	WorkflowId string `json:"workflowId,omitempty"`
	NodeId     string `json:"nodeId,omitempty"`
	WebhookId  string `json:"webhookId,omitempty"`
} //@name WorkflowWebhook

type WorkflowFrom struct {
	Path            string `json:"path"`
	FormTitle       string `json:"formTitle"`
	FormDescription string `json:"formDescription"`
	FormFields      struct {
		Values []WorkflowFromFields `json:"values"`
	} `json:"formFields"`
	Options map[string]interface{} `json:"options"`
} //@name WorkflowFrom

// link https://github.com/n8n-io/n8n/blob/7e1036187ff7bd5be990f191a3ac8ef002e7812a/packages/workflow/src/Interfaces.ts#L2641
type WorkflowFromFields struct {
	FieldLabel    string `json:"fieldLabel"`
	FieldType     string `json:"fieldType"`
	RequiredField bool   `json:"requiredField,omitempty"`
	FieldOptions  struct {
		Values []struct {
			Option string `json:"option"`
		} `json:"values,omitempty"`
	} `json:"fieldOptions,omitempty"`
} //@name WorkflowFromFields
type GetWorkflowFromResponse struct {
	Parameters *WorkflowFrom `json:"parameters"`
} //@name GetWorkflowResponse
