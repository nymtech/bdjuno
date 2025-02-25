package tx_worker

import (
	"github.com/forbole/juno/v6/modules"
	juno "github.com/forbole/juno/v6/types"
	"github.com/rs/zerolog/log"

	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/callisto/v4/utils/events"
)

var (
	_ modules.Module            = &Module{}
	_ modules.TransactionModule = &Module{}
)

// Module represents the tx_worker module
type Module struct {
	db *database.Db
}

// NewModule returns a new Module instance
func NewModule(db *database.Db) *Module {
	log.Info().Msg("TX_WORKER: Initializing transaction log enrichment module")
	return &Module{
		db: db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "tx_worker"
}

// HandleTx implements modules.TransactionModule
// This is called before any other module receives the transaction
func (m *Module) HandleTx(tx *juno.Transaction) error {
	log.Info().
		Str("txhash", tx.TxHash).
		Uint64("height", tx.Height).
		Int("events_count", len(tx.Events)).
		Int("logs_count", len(tx.Logs)).
		Msg("TX_WORKER: Processing transaction")

	// Log transaction state before processing
	if len(tx.Logs) > 0 {
		log.Debug().
			Str("txhash", tx.TxHash).
			Int("logs_count", len(tx.Logs)).
			Msg("TX_WORKER: Transaction already has logs, skipping")
		return nil
	}

	if len(tx.Events) == 0 {
		log.Debug().
			Str("txhash", tx.TxHash).
			Msg("TX_WORKER: Transaction has no events, skipping")
		return nil
	}

	// Use the utility function to update transaction logs if needed
	updated := events.UpdateTransactionLogs(tx, m.db)

	if updated {
		log.Info().
			Str("txhash", tx.TxHash).
			Int("new_logs_count", len(tx.Logs)).
			Msg("TX_WORKER: Successfully enriched transaction logs")
	} else {
		log.Debug().
			Str("txhash", tx.TxHash).
			Msg("TX_WORKER: Transaction logs not updated")
	}

	return nil
}
