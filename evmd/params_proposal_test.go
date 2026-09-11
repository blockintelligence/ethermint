package evmd_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/evmos/ethermint/evmd"
	"github.com/evmos/ethermint/testutil"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

func TestFeeMarketAwareParamChangeHandlerWritesModuleStore(t *testing.T) {
	app := testutil.Setup(false, nil)
	ctx := app.NewUncachedContext(false, tmproto.Header{ChainID: testutil.ChainID})

	params := app.FeeMarketKeeper.GetParams(ctx)
	require.False(t, params.NoBaseFee)
	elasticity := params.ElasticityMultiplier

	content := paramproposal.NewParameterChangeProposal(
		"enable no-base-fee",
		"legacy param change should update the v4 module store",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemarkettypes.ModuleName,
				string(feemarkettypes.ParamStoreKeyNoBaseFee),
				"true",
			),
		},
	)

	handler := evmd.NewFeeMarketAwareParamChangeHandler(app.ParamsKeeper, app.FeeMarketKeeper)
	require.NoError(t, handler(ctx, content))

	got := app.FeeMarketKeeper.GetParams(ctx)
	require.True(t, got.NoBaseFee, "ParameterChangeProposal must update GetParams after the v4 store")
	require.Equal(t, elasticity, got.ElasticityMultiplier, "unrelated params must be preserved")
}
