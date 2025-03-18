package temporal_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sugerio/workflow-service-trial/shared/temporal"
	"go.temporal.io/sdk/converter"
)

func Test_DataConverter(t *testing.T) {
	defaultDc := converter.GetDefaultDataConverter()

	cryptDc := temporal.NewEncryptionDataConverter(
		converter.GetDefaultDataConverter(),
		temporal.EncryptionDataConverterOptions{
			Key: "test-encrypt-key",
		},
	)

	defaultPayloads, err := defaultDc.ToPayloads("Testing")
	require.NoError(t, err)

	encryptedPayloads, err := cryptDc.ToPayloads("Testing")
	require.NoError(t, err)

	require.NotEqual(t, defaultPayloads.Payloads[0].GetData(), encryptedPayloads.Payloads[0].GetData())

	var result string
	err = cryptDc.FromPayloads(encryptedPayloads, &result)
	require.NoError(t, err)

	require.Equal(t, "Testing", result)
}
