package keeper_test

import (
	"reflect"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/evmos/ethermint/testutil"
	"github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/suite"
)

type ParamsTestSuite struct {
	testutil.BaseTestSuite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestSetGetParams() {
	params := types.DefaultParams()
	suite.App.FeeMarketKeeper.SetParams(suite.Ctx, params)
	testCases := []struct {
		name      string
		paramsFun func() interface{}
		getFun    func() interface{}
		expected  bool
	}{
		{
			"success - Checks if the default params are set correctly",
			func() interface{} {
				return types.DefaultParams()
			},
			func() interface{} {
				return suite.App.FeeMarketKeeper.GetParams(suite.Ctx)
			},
			true,
		},
		{
			"success - Check ElasticityMultiplier is set to 3 and can be retrieved correctly",
			func() interface{} {
				params.ElasticityMultiplier = 3
				suite.App.FeeMarketKeeper.SetParams(suite.Ctx, params)
				return params.ElasticityMultiplier
			},
			func() interface{} {
				return suite.App.FeeMarketKeeper.GetParams(suite.Ctx).ElasticityMultiplier
			},
			true,
		},
		{
			"success - Check BaseFeeEnabled is computed with its default params and can be retrieved correctly",
			func() interface{} {
				suite.App.FeeMarketKeeper.SetParams(suite.Ctx, types.DefaultParams())
				return true
			},
			func() interface{} {
				return suite.App.FeeMarketKeeper.GetParams(suite.Ctx).IsBaseFeeEnabled(suite.Ctx.BlockHeight())
			},
			true,
		},
		{
			"success - Check BaseFeeEnabled is computed with alternate params and can be retrieved correctly",
			func() interface{} {
				params.NoBaseFee = true
				params.EnableHeight = 5
				suite.App.FeeMarketKeeper.SetParams(suite.Ctx, params)
				return true
			},
			func() interface{} {
				return suite.App.FeeMarketKeeper.GetParams(suite.Ctx).IsBaseFeeEnabled(suite.Ctx.BlockHeight())
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			outcome := reflect.DeepEqual(tc.paramsFun(), tc.getFun())
			suite.Require().Equal(tc.expected, outcome)
		})
	}
}

func (suite *ParamsTestSuite) TestApplyLegacySubspaceChanges() {
	params := types.DefaultParams()
	params.MinGasPrice = sdkmath.LegacyNewDec(1)
	params.ElasticityMultiplier = 4
	suite.Require().NoError(suite.App.FeeMarketKeeper.SetParams(suite.Ctx, params))

	newPrice := sdkmath.LegacyNewDec(99)
	ss := suite.App.GetSubspace(types.ModuleName)
	ss.Set(suite.Ctx, types.ParamStoreKeyMinGasPrice, newPrice)

	err := suite.App.FeeMarketKeeper.ApplyLegacySubspaceChanges(suite.Ctx, []string{string(types.ParamStoreKeyMinGasPrice)})
	suite.Require().NoError(err)

	got := suite.App.FeeMarketKeeper.GetParams(suite.Ctx)
	suite.Require().True(got.MinGasPrice.Equal(newPrice), "min gas price should come from the subspace change")
	suite.Require().Equal(uint32(4), got.ElasticityMultiplier, "unrelated module-store params must be preserved")
}
