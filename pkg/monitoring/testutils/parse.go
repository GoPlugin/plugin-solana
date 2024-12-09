package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goplugin/plugin-solana/pkg/monitoring/types"
)

func ParseTxDetails(t *testing.T, in interface{}) []types.TxDetails {
	out, err := types.MakeTxDetails(in)
	require.NoError(t, err)
	return out
}
