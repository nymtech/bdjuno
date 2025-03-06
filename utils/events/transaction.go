package events

import (
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	juno "github.com/forbole/juno/v6/types"
	"github.com/rs/zerolog/log"

	"github.com/forbole/callisto/v4/database"
)

// TransactionLogUpdater represents an interface for updating transaction logs
type TransactionLogUpdater interface {
	UpdateTransactionLogs(hash string, logs []byte) error
}

// ConvertEventsToLogs converts a slice of ABCI events to a slice of SDK ABCIMessageLog
// It groups events by message index when possible
func ConvertEventsToLogs(events []abci.Event) []sdk.ABCIMessageLog {
	var logs []sdk.ABCIMessageLog

	// Create a log entry for each message index by grouping events with the same msg_index
	msgIndexMap := make(map[uint32][]abci.Event)

	// Group events by message index, default to 0 if not specified
	for _, event := range events {
		msgIndex := uint32(0) // Default to first message

		// Try to find msg_index attribute
		for _, attr := range event.Attributes {
			if attr.Key == "msg_index" {
				// Try to parse as integer
				var idx int
				_, err := fmt.Sscanf(attr.Value, "%d", &idx)
				if err == nil && idx >= 0 {
					msgIndex = uint32(idx)
					break
				}
			}
		}

		// Add event to the appropriate message group
		msgIndexMap[msgIndex] = append(msgIndexMap[msgIndex], event)
	}

	// If we couldn't find any message indices, use all events for the first message
	if len(msgIndexMap) == 0 && len(events) > 0 {
		msgIndexMap[0] = events
	}

	// Create logs for each message index
	for msgIndex, msgEvents := range msgIndexMap {
		stringEvents := sdk.StringifyEvents(msgEvents)

		msgLog := sdk.ABCIMessageLog{
			MsgIndex: msgIndex,
			Events:   stringEvents,
		}
		logs = append(logs, msgLog)
	}

	return logs
}

// UpdateTransactionLogs updates the transaction logs in the database if they're empty
// and there are events available. Returns true if logs were updated, false otherwise.
func UpdateTransactionLogs(tx *juno.Transaction, db *database.Db) bool {
	// Skip if logs are already present
	if len(tx.Logs) > 0 {
		return false
	}

	// Skip if no events are present
	if len(tx.Events) == 0 {
		return false
	}

	// Convert events to logs format
	logs := ConvertEventsToLogs(tx.Events)

	// Skip if no logs were created
	if len(logs) == 0 {
		return false
	}

	// Store the logs in the transaction
	logsBz, err := json.Marshal(logs)
	if err != nil {
		log.Error().
			Err(err).
			Str("txhash", tx.TxHash).
			Msg("EVENTS: Error while marshaling logs")
		return false
	}

	// Update the transaction logs
	err = db.UpdateTransactionLogs(tx.TxHash, logsBz)
	if err != nil {
		log.Error().
			Err(err).
			Str("txhash", tx.TxHash).
			Msg("EVENTS: Error while updating transaction logs")
		return false
	}

	// Update the transaction object with the new logs
	tx.Logs = logs

	return true
}
