package api_test

// Command to run this test only
// go test -v service/workflow_service/api/service_test.go service/workflow_service/api/middleware_test.go

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MiddlewareTestSuite struct {
	suite.Suite
}

func Test_MiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func (s *MiddlewareTestSuite) Test_RecoverPanic() {
	assert := require.New(s.T())

	// Get Proxy Request template.
	request, err := GetAPIGatewayProxyRequest_CreateOrganization()
	assert.Nil(err)

	// Test panic middleware.
	request.Body = ""
	request.HTTPMethod = "GET"
	request.Path = "/panic"
	response, err := testFiberLambda.Proxy(request)
	// should be nil
	assert.Nil(err)
	// should return 500
	assert.Equal(500, response.StatusCode)
}
