package structs

import (
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
)

// Convert from rdsDbLib.IdentityIntegration to IdentityIntegration.
func FromIdentityIntegration(integration *rdsDbLib.IdentityIntegration) (*IdentityIntegration, error) {
	result := IdentityIntegration{}
	CopyCommonFields(integration, &result)

	integrationInfo, err := IntegrationInfoFromJson(integration.Info)
	if err != nil {
		return nil, err
	}

	result.Info = integrationInfo

	return &result, nil
}
