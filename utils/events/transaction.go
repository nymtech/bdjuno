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

	log.Debug().
		Int("events_count", len(events)).
		Msg("EVENTS: Starting conversion of events to logs")

	// Create a log entry for each message index by grouping events with the same msg_index
	msgIndexMap := make(map[uint32][]abci.Event)

	// Group events by message index, default to 0 if not specified
	for i, event := range events {
		msgIndex := uint32(0) // Default to first message

		// Try to find msg_index attribute
		for _, attr := range event.Attributes {
			if attr.Key == "msg_index" {
				// Try to parse as integer
				var idx int
				_, err := fmt.Sscanf(attr.Value, "%d", &idx)
				if err == nil && idx >= 0 {
					msgIndex = uint32(idx)
					log.Debug().
						Int("event_index", i).
						Str("event_type", event.Type).
						Uint32("msg_index", msgIndex).
						Msg("EVENTS: Found explicit msg_index in event")
					break
				}
			}
		}

		// Add event to the appropriate message group
		msgIndexMap[msgIndex] = append(msgIndexMap[msgIndex], event)
		log.Debug().
			Int("event_index", i).
			Str("event_type", event.Type).
			Uint32("assigned_msg_index", msgIndex).
			Msg("EVENTS: Assigned event to message group")
	}

	// If we couldn't find any message indices, use all events for the first message
	if len(msgIndexMap) == 0 && len(events) > 0 {
		log.Debug().Msg("EVENTS: No message indices found, assigning all events to message index 0")
		msgIndexMap[0] = events
	}

	// Log the grouping results
	for msgIndex, events := range msgIndexMap {
		log.Debug().
			Uint32("msg_index", msgIndex).
			Int("events_count", len(events)).
			Msg("EVENTS: Message group events count")
	}

	// Create logs for each message index
	for msgIndex, msgEvents := range msgIndexMap {
		stringEvents := sdk.StringifyEvents(msgEvents)

		msgLog := sdk.ABCIMessageLog{
			MsgIndex: msgIndex,
			Events:   stringEvents,
		}
		logs = append(logs, msgLog)

		log.Debug().
			Uint32("msg_index", msgIndex).
			Int("events_count", len(stringEvents)).
			Msg("EVENTS: Created log entry for message index")
	}

	log.Debug().
		Int("created_logs_count", len(logs)).
		Msg("EVENTS: Completed conversion of events to logs")

	return logs
}

// UpdateTransactionLogs updates the transaction logs in the database if they're empty
// and there are events available. Returns true if logs were updated, false otherwise.
func UpdateTransactionLogs(tx *juno.Transaction, db *database.Db) bool {
	// Skip if logs are already present
	if len(tx.Logs) > 0 {
		log.Debug().
			Str("txhash", tx.TxHash).
			Int("logs_count", len(tx.Logs)).
			Msg("EVENTS: Transaction already has logs, skipping")
		return false
	}

	// Skip if no events are present
	if len(tx.Events) == 0 {
		log.Debug().
			Str("txhash", tx.TxHash).
			Msg("EVENTS: Transaction has no events, skipping")
		return false
	}

	log.Info().
		Str("txhash", tx.TxHash).
		Int("events_count", len(tx.Events)).
		Msg("EVENTS: Enriching transaction logs from events")

	// Log some details about the events for debugging
	for i, event := range tx.Events {
		log.Debug().
			Str("txhash", tx.TxHash).
			Int("event_index", i).
			Str("event_type", event.Type).
			Int("attributes_count", len(event.Attributes)).
			Msg("EVENTS: Processing event")
	}

	// Convert events to logs format
	logs := ConvertEventsToLogs(tx.Events)

	// Skip if no logs were created
	if len(logs) == 0 {
		log.Debug().
			Str("txhash", tx.TxHash).
			Msg("EVENTS: No logs were created from events")
		return false
	}

	log.Debug().
		Str("txhash", tx.TxHash).
		Int("created_logs_count", len(logs)).
		Msg("EVENTS: Successfully converted events to logs")

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

	log.Info().
		Str("txhash", tx.TxHash).
		Int("logs_count", len(logs)).
		Msg("EVENTS: Successfully updated transaction logs in database")

	// Update the transaction object with the new logs
	tx.Logs = logs

	return true
}
