package bigquery_test

// Command to run this test only
// go test -v shared/bigquery/init_test.go shared/bigquery/auth_test.go

import (
	"context"
	"encoding/json"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	sharedBigquery "github.com/sugerio/workflow-service-trial/integration/bigquery"
	"github.com/sugerio/workflow-service-trial/shared/structs"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
)

type AuthTestSuite struct {
	suite.Suite
}

func Test_AuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) Test() {
	s.T().Run("Test GetAccessToken from ServiceAccountPrivateKey", func(t *testing.T) {
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		bigqueryIntegration, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		// Get Access Token with invalid info. Fails.
		accessToken, err := sharedBigquery.GetAccessToken("test@email.com", "invalid-private-key")
		assert.Nil(accessToken, "Access token should be nil")
		assert.Error(err, "Expected error from GetAccessToken")

		serviceAccountPrivateKey := structs.GcpAuthPrivateKey{}
		err = json.Unmarshal([]byte(bigqueryIntegration.ServiceAccountPrivateKeyJson), &serviceAccountPrivateKey)
		assert.Nil(err)

		// Get Access Token with valid info. This should succeed.
		accessToken, err = sharedBigquery.GetAccessToken(
			bigqueryIntegration.ServiceAccountEmail, serviceAccountPrivateKey.PrivateKey)
		assert.NotNil(accessToken, "Access token should not be nil")
		assert.NoError(err, "Expected no error from GetAccessToken")
	})

	s.T().Run("Test Bigquery Client from ServiceAccountPrivateKeyJson", func(t *testing.T) {
		ctx := context.Background()
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		bigqueryIntegration, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		client, err := bigquery.NewClient(
			ctx, "suger-stag", option.WithCredentialsJSON([]byte(bigqueryIntegration.ServiceAccountPrivateKeyJson)))
		assert.Nil(err)
		assert.NotNil(client)

		dataset, err := client.Datasets(ctx).Next()
		assert.Nil(err)
		assert.NotNil(dataset)
		assert.Equal("suger-stag", dataset.ProjectID)
	})
}
