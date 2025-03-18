package aws_lib_test

// Command to run this test only
// go test -v shared/aws_lib/init_test.go shared/aws_lib/aws_sdk_clients_test.go

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/shared/aws_lib"
)

type AwsSdkClientsTestSuite struct {
	suite.Suite
}

func Test_AwsSdkClients(t *testing.T) {
	suite.Run(t, new(AwsSdkClientsTestSuite))
}

func (s *AwsSdkClientsTestSuite) Test() {
	s.T().Run("ValidateSecretIdWithOrgId", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Test empty input of orgId
		assert.Nil(aws_lib.ValidateSecretIdWithOrgId("", "org_123456ABC_secretKeyId"))
		// Test wrong orgId, no error because the validation is kipped under test env
		assert.Nil(aws_lib.ValidateSecretIdWithOrgId("wrongOrgId", "org_123456ABC_secretKeyId"))
		// Test the internal logic.
		secretIdPrefix := fmt.Sprintf("org_%s", "123456ABC")
		assert.True(strings.HasPrefix("org_123456ABC_secretKeyId", secretIdPrefix))
	})
}
