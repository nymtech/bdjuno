package database

import (
	"fmt"

	dbtypes "github.com/forbole/bdjuno/v4/database/types"
)

// GetMessages returns all messages between block heights
func (db *Db) GetMessages(address string, startHeight int64, endHeight int64, offset uint64, limit uint64) ([]dbtypes.MessageRow, error) {
	query := fmt.Sprintf(`SELECT 
		transaction_hash,
		index,
		type,
		value,
		involved_accounts_addresses,
		height
    FROM message 
         WHERE involved_accounts_addresses @> '{%s}' AND height >= %d AND height <= %d
         ORDER BY height ASC, type ASC, index ASC
         OFFSET %d LIMIT %d`, address, startHeight, endHeight, offset, limit)

	var dbRows []dbtypes.MessageRow
	err := db.Sqlx.Select(&dbRows, query)
	if err != nil {
		return nil, err
	}

	return dbRows, nil
}

// GetMessagesCount returns count of GetMessages
func (db *Db) GetMessagesCount(address string, startHeight int64, endHeight int64) (int, error) {

	stmt, err := db.Sqlx.Prepare(fmt.Sprintf(`SELECT count(height)
    FROM message 
         WHERE involved_accounts_addresses @> '{%s}' AND height >= %d AND height <= %d`, address, startHeight, endHeight))
	if err != nil {
		return 0, err
	}

	var count int
	err = stmt.QueryRow().Scan(&count)

	if err != nil {
		return 0, err
	}
	return count, nil
}
