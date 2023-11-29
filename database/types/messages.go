package types

import (
	"encoding/json"

	"github.com/lib/pq"
)

type MessageRow struct {
	TxHash  string         `db:"transaction_hash"`
	Index   int            `db:"index"`
	Type    string         `db:"type"`
	Value   string         `db:"value"`
	Aliases pq.StringArray `db:"involved_accounts_addresses"`
	Height  int64          `db:"height"`
}

type MessageWithValueRow struct {
	TxHash  string         `json:"transaction_hash"`
	Index   int            `json:"index"`
	Type    string         `json:"type"`
	Value   interface{}    `json:"value"`
	Funds   interface{}    `json:"funds"`
	Aliases pq.StringArray `json:"involved_accounts_addresses"`
	Height  int64          `json:"height"`
}

func NewMessageWithValueRow(row MessageRow) (*MessageWithValueRow, error) {
	var value map[string]interface{}
	err := json.Unmarshal([]byte(row.Value), &value)
	if err != nil {
		return nil, err
	}

	var funds interface{}

	if value["funds"] != nil {
		funds = value["funds"]
	}
	if value["amount"] != nil {
		funds = value["amount"]
	}

	var rowWithValue = MessageWithValueRow{
		TxHash:  row.TxHash,
		Index:   row.Index,
		Type:    row.Type,
		Value:   value,
		Funds:   funds,
		Aliases: row.Aliases,
		Height:  row.Height,
	}

	return &rowWithValue, nil
}
