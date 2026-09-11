package keeper

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/store/v2/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/x/evm/types"
)

func (k Keeper) SetTxBloom(ctx sdk.Context, bloom *big.Int) {
	store := ctx.ObjectStore(k.objectKey)
	store.Set(types.ObjectBloomKey(ctx.TxIndex(), ctx.MsgIndex()), bloom)
}

// setTxBloomFromLogs writes the per-tx bloom derived from the given logs.
// It is a no-op when logs is empty so a hook that clears logs does not leave a stale bloom.
func (k Keeper) setTxBloomFromLogs(ctx sdk.Context, logs []*ethtypes.Log) {
	if len(logs) == 0 {
		return
	}
	bloom := ethtypes.Bloom{}
	for _, log := range logs {
		bloom.Add(log.Address.Bytes())
		for _, topic := range log.Topics {
			bloom.Add(topic[:])
		}
	}
	k.SetTxBloom(ctx, bloom.Big())
}

func (k Keeper) CollectTxBloom(ctx sdk.Context) {
	store := prefix.NewObjStore(ctx.ObjectStore(k.objectKey), types.KeyPrefixObjectBloom)
	it := store.Iterator(nil, nil)
	defer it.Close()

	bloom := new(big.Int)
	for ; it.Valid(); it.Next() {
		bloom.Or(bloom, it.Value().(*big.Int))
	}

	k.EmitBlockBloomEvent(ctx, bloom.Bytes())
}
