package database

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// UpdateTransactionLogs updates the logs field of a transaction
func (db *Db) UpdateTransactionLogs(hash string, logs []byte) error {
	stmt := `UPDATE transaction SET logs = $1 WHERE hash = $2`

	log.Debug().
		Str("txhash", hash).
		Int("logs_bytes", len(logs)).
		Msg("DATABASE: Updating transaction logs")

	result, err := db.SQL.Exec(stmt, logs, hash)
	if err != nil {
		log.Error().
			Err(err).
			Str("txhash", hash).
			Msg("DATABASE: Error while updating transaction logs")
		return fmt.Errorf("error while updating transaction logs: %s", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Warn().
			Err(err).
			Str("txhash", hash).
			Msg("DATABASE: Could not get rows affected count")
	} else {
		log.Debug().
			Str("txhash", hash).
			Int64("rows_affected", rowsAffected).
			Msg("DATABASE: Transaction logs update complete")
	}

	return nil
}
