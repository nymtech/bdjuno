package wasm

import (
	juno "github.com/forbole/juno/v3/types"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// HandleBlock implements modules.BlockModule
func (m *Module) HandleBlock(block *tmctypes.ResultBlock, results *tmctypes.ResultBlockResults, txs []*juno.Tx, vals *tmctypes.ResultValidators) error {
	return nil
}
