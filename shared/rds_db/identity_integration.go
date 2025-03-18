package rds_db

import (
	"context"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

// CreateIntegration Create a new integration
func (q *Queries) CreateIntegration(
	ctx context.Context,
	params *structs.CreateIntegrationParams) (*structs.IdentityIntegration, error) {
	paramsRdsDbLib, err := params.ToRdsDbLib()
	if err != nil {
		return nil, err
	}
	paramsRdsDbLib.Status = string(structs.IntegrationStatus_CREATED)

	integrationRdsDbLib, err := q.rdsDbQueries.CreateIntegration(ctx, *paramsRdsDbLib)
	if err != nil {
		return nil, err
	}

	// Convert to structs.IdentityIntegration
	return structs.FromIdentityIntegration(&integrationRdsDbLib)
}

// GetIntegration return error if not found
func (q *Queries) GetIntegration(
	ctx context.Context,
	orgID string,
	partner structs.Partner,
	service structs.PartnerService) (*structs.IdentityIntegration, error) {
	integrationRdsDbLib, err := q.rdsDbQueries.GetIntegration(
		ctx,
		rdsDbLib.GetIntegrationParams{
			OrganizationID: orgID,
			Partner:        string(partner),
			Service:        string(service),
		})
	if err != nil {
		return nil, err
	}

	return structs.FromIdentityIntegration(&integrationRdsDbLib)
}
