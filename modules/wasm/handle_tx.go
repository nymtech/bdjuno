package wasm

import (
	juno "github.com/forbole/juno/v5/types"
)

// HandleTx implements modules.TransactionModule
func (m *Module) HandleTx(tx *juno.Tx) error {
	// log.Debug().Str("txhash", tx.TxHash).Msg("wasm HandleTx")
	return nil
}
