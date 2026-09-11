// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package keeper

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ethermint "github.com/evmos/ethermint/types"
	"github.com/evmos/ethermint/x/feemarket/types"
)

// GetParams returns the total set of fee market parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	bz := ctx.KVStore(k.storeKey).Get(types.ParamsKey)
	if len(bz) == 0 {
		k.ss.GetParamSetIfExists(ctx, &params)
	} else {
		k.cdc.MustUnmarshal(bz, &params)
	}
	return params
}

// SetParams sets the fee market params in a single key
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)

	return nil
}

// ApplyLegacySubspaceChanges copies the given legacy subspace keys into the
// module-store params. Only those keys are overlaid so MsgUpdateParams values
// for other fields are preserved. This is required because after the v4
// migration GetParams prefers ParamsKey and ignores subspace writes from
// ParameterChangeProposal.
func (k Keeper) ApplyLegacySubspaceChanges(ctx sdk.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	params := k.GetParams(ctx)
	for _, key := range keys {
		if err := overlaySubspaceParam(ctx, k.ss, &params, key); err != nil {
			return err
		}
	}
	return k.SetParams(ctx, params)
}

func overlaySubspaceParam(ctx sdk.Context, ss paramstypes.Subspace, params *types.Params, key string) error {
	switch key {
	case string(types.ParamStoreKeyNoBaseFee):
		ss.GetIfExists(ctx, types.ParamStoreKeyNoBaseFee, &params.NoBaseFee)
	case string(types.ParamStoreKeyBaseFeeChangeDenominator):
		ss.GetIfExists(ctx, types.ParamStoreKeyBaseFeeChangeDenominator, &params.BaseFeeChangeDenominator)
	case string(types.ParamStoreKeyElasticityMultiplier):
		ss.GetIfExists(ctx, types.ParamStoreKeyElasticityMultiplier, &params.ElasticityMultiplier)
	case string(types.ParamStoreKeyBaseFee):
		var baseFee sdkmath.Int
		ss.GetIfExists(ctx, types.ParamStoreKeyBaseFee, &baseFee)
		if !baseFee.IsNil() {
			params.BaseFee = baseFee
		}
	case string(types.ParamStoreKeyEnableHeight):
		ss.GetIfExists(ctx, types.ParamStoreKeyEnableHeight, &params.EnableHeight)
	case string(types.ParamStoreKeyMinGasPrice):
		ss.GetIfExists(ctx, types.ParamStoreKeyMinGasPrice, &params.MinGasPrice)
	case string(types.ParamStoreKeyMinGasMultiplier):
		ss.GetIfExists(ctx, types.ParamStoreKeyMinGasMultiplier, &params.MinGasMultiplier)
	default:
		return fmt.Errorf("unknown feemarket param key %q", key)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Parent Base Fee
// Required by EIP1559 base fee calculation.
// ----------------------------------------------------------------------------

// GetBaseFee gets the base fee from the store
func (k Keeper) GetBaseFee(ctx sdk.Context) *big.Int {
	params := k.GetParams(ctx)
	if params.NoBaseFee {
		return nil
	}

	baseFee := params.BaseFee.BigInt()
	if baseFee == nil || baseFee.Sign() == 0 {
		// try v1 format
		return k.GetBaseFeeV1(ctx)
	}
	return baseFee
}

// SetBaseFee set's the base fee in the store
func (k Keeper) SetBaseFee(ctx sdk.Context, baseFee *big.Int) {
	params := k.GetParams(ctx)
	params.BaseFee = ethermint.SaturatedNewInt(baseFee)
	err := k.SetParams(ctx, params)
	if err != nil {
		return
	}
}
