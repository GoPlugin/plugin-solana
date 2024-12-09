package exporter

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goplugin/plugin-common/pkg/logger"
	commonMonitoring "github.com/goplugin/plugin-common/pkg/monitoring"
	"github.com/goplugin/plugin-common/pkg/utils/tests"

	"github.com/goplugin/plugin-solana/pkg/monitoring/metrics/mocks"
	"github.com/goplugin/plugin-solana/pkg/monitoring/testutils"
	"github.com/goplugin/plugin-solana/pkg/monitoring/types"
)

func TestSlotHeight(t *testing.T) {
	ctx := tests.Context(t)
	m := mocks.NewSlotHeight(t)
	m.On("Set", mock.Anything, mock.Anything, mock.Anything).Once()
	m.On("Cleanup").Once()

	factory := NewSlotHeightFactory(logger.Test(t), m)

	chainConfig := testutils.GenerateChainConfig()
	exporter, err := factory.NewExporter(commonMonitoring.ExporterParams{ChainConfig: chainConfig})
	require.NoError(t, err)

	// happy path
	exporter.Export(ctx, types.SlotHeight(10))
	exporter.Cleanup(ctx)

	// test passing uint64 instead of SlotHeight - should not call mock
	// SlotHeight alias of uint64
	exporter.Export(ctx, uint64(10))
}
