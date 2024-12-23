package solana

import (
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"

	"github.com/goplugin/plugin-common/pkg/utils/tests"
)

func TestConfigDigester(t *testing.T) {
	programID, err := solana.PublicKeyFromBase58("HW3ipKzeeduJq6f1NqRCw4doknMeWkfrM4WxobtG3o5v")
	require.NoError(t, err)
	stateID, err := solana.PublicKeyFromBase58("ES64UceMzVRQ1t9j7VZKHi7A2cJ4seVmbKNmbtFZUiYz")
	require.NoError(t, err)
	digester := OffchainConfigDigester{
		ProgramID: programID,
		StateID:   stateID,
	}

	// Test ConfigDigester by using a known raw state account + known expected digest
	var state State
	err = bin.NewBorshDecoder(mockState.Raw).Decode(&state)
	require.NoError(t, err)
	config, err := ConfigFromState(tests.Context(t), state)
	require.NoError(t, err)

	actualDigest, err := digester.ConfigDigest(tests.Context(t), config)
	require.NoError(t, err)

	expectedDigest := mockState.ConfigDigestHex
	require.Equal(t, expectedDigest, actualDigest.Hex())
}
