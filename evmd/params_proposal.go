package evmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	feemarketkeeper "github.com/evmos/ethermint/x/feemarket/keeper"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

// NewFeeMarketAwareParamChangeHandler wraps the legacy ParameterChangeProposal
// handler so feemarket changes are copied from the unused x/params subspace
// into the v4 module store that GetParams actually reads.
func NewFeeMarketAwareParamChangeHandler(pk paramskeeper.Keeper, fmk feemarketkeeper.Keeper) govv1beta1.Handler {
	legacy := params.NewParamChangeProposalHandler(pk)
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		if err := legacy(ctx, content); err != nil {
			return err
		}

		p, ok := content.(*paramproposal.ParameterChangeProposal)
		if !ok {
			return nil
		}

		keys := make([]string, 0, len(p.Changes))
		for _, c := range p.Changes {
			if c.Subspace == feemarkettypes.ModuleName {
				keys = append(keys, c.Key)
			}
		}
		return fmk.ApplyLegacySubspaceChanges(ctx, keys)
	}
}
