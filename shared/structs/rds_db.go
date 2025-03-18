package structs

import (
	"encoding/json"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"time"
)

type IdentityIntegration struct {
	OrganizationID string          `json:"organizationID"`
	Partner        string          `json:"partner" enums:"AWS,AZURE,GCP,ALIBABA"`
	Service        string          `json:"service" enums:"MARKETPLACE,BILLING"`
	Status         string          `json:"status" enums:"CREATED,VERIFIED,NOT_VERIFIED"`
	Info           IntegrationInfo `json:"info"`
	CreationTime   time.Time       `json:"creationTime" format:"date-time"`
	CreatedBy      string          `json:"createdBy"`
	LastUpdateTime time.Time       `json:"lastUpdateTime" format:"date-time"`
	LastUpdatedBy  string          `json:"lastUpdatedBy"`
} //@name IdentityIntegration

type IdentityOrganization struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	EmailDomain        string           `json:"emailDomain"`
	Website            string           `json:"website"`
	Description        string           `json:"description"`
	CreationTime       time.Time        `json:"creationTime" format:"date-time"`
	Status             string           `json:"status"`
	LastUpdateTime     time.Time        `json:"lastUpdateTime" format:"date-time"`
	AllowedAuthMethods []string         `json:"allowedAuthMethods"`
	CreatedBy          string           `json:"createdBy"`
	AuthID             string           `json:"authID"`
	Info               OrganizationInfo `json:"info"`
} //@name IdentityOrganization

type ListOrganizationsByUserRow struct {
	ID                   string           `json:"id"`
	Name                 string           `json:"name"`
	EmailDomain          string           `json:"emailDomain"`
	Website              string           `json:"website"`
	Description          string           `json:"description"`
	CreationTime         time.Time        `json:"creationTime" format:"date-time"`
	Status               string           `json:"status"`
	LastUpdateTime       time.Time        `json:"lastUpdateTime" format:"date-time"`
	AllowedAuthMethods   []string         `json:"allowedAuthMethods"`
	CreatedBy            string           `json:"createdBy"`
	AuthID               string           `json:"authID"`
	Info                 OrganizationInfo `json:"info"`
	UserRole             string           `json:"userRole"`
	JoinOrganizationTime time.Time        `json:"joinOrganizationTime" format:"date-time"`
} //@name ListOrganizationsByUserRow

type UpdateIntegrationParams struct {
	OrganizationID string          `json:"organizationID" validate:"required"`
	Partner        Partner         `json:"partner" validate:"required"`
	Service        PartnerService  `json:"service" validate:"required"`
	Info           IntegrationInfo `json:"info" validate:"required"`
} //@name UpdateIntegrationParams

type AuditingAction struct {
	SerialID       int64     `json:"serialID"`
	OrganizationID string    `json:"organizationID"`
	EntityType     string    `json:"entityType"`
	EntityID       string    `json:"entityID"`
	ActionType     string    `json:"actionType"`
	CreationTime   time.Time `json:"creationTime" format:"date-time"`
	CreatedBy      string    `json:"createdBy"`
} //@name AuditingAction

type UpdateWebhookParams struct {
	OrganizationID string  `json:"organizationID" validate:"required"`
	ID             string  `json:"id" validate:"required"`
	PayloadUrl     *string `json:"payloadUrl,omitempty" validate:"optional"`
	Secret         *string `json:"secret,omitempty" validate:"optional"`
	Status         *string `json:"status,omitempty" validate:"optional"`
} //@name UpdateWebhookParams

type AuditingEvent struct {
	ID             string      `json:"id"`
	OrganizationID string      `json:"organizationID"`
	EventType      string      `json:"eventType" enums:"AWS_MARKETPLACE,AZURE_MARKETPLACE,GCP_MARKETPLACE"`
	Status         string      `json:"status" enums:"AUDITED,PENDING,FAILED,DONE"`
	CreationTime   time.Time   `json:"creationTime" format:"date-time"`   // When the event is received and audited.
	LastUpdateTime time.Time   `json:"lastUpdateTime" format:"date-time"` // when the event is updated.
	OtherTypeEvent interface{} `json:"otherTypeEvent,omitempty"`
} //@name AuditingEvent

type CreateIntegrationParams struct {
	OrganizationID string          `json:"organizationID" validate:"required"`
	Partner        Partner         `json:"partner" validate:"required"`
	Service        PartnerService  `json:"service" validate:"required"`
	Info           IntegrationInfo `json:"info" validate:"required"`
	CreatedBy      string          `json:"createdBy"`
} //@name CreateIntegrationParams

// ToRdsDbLib converts from CreateIntegrationParams to rdsDbLib.CreateIntegrationParams.
func (ob *CreateIntegrationParams) ToRdsDbLib() (*rdsDbLib.CreateIntegrationParams, error) {
	result := rdsDbLib.CreateIntegrationParams{}
	CopyCommonFields(ob, &result)
	result.Partner = string(ob.Partner)
	result.Service = string(ob.Service)

	var err error
	if result.Info, err = json.Marshal(ob.Info); err != nil {
		return nil, err
	}

	return &result, nil
}
