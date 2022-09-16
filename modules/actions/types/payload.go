package types

import (
	"github.com/cosmos/cosmos-sdk/types/query"
	"time"
)

// Payload contains the payload data that is sent from Hasura
type Payload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            PayloadArgs            `json:"input"`
}

// GetAddress returns the address associated with this payload, if any
func (p *Payload) GetAddress() string {
	return p.Input.Address
}

// GetIdentityKey returns the identity key associated with this payload, if any
func (p *Payload) GetIdentityKey() *string {
	return &p.Input.IdentityKey
}

// GetHeight returns the block height associated with this payload, if any
func (p *Payload) GetHeight() *int64 {
	return &p.Input.Height
}

func (p *Payload) GetExecutedAtStart() *time.Time {
	return &p.Input.ExecutedAtStart
}

func (p *Payload) GetExecutedAtEnd() *time.Time {
	return &p.Input.ExecutedAtEnd
}

// GetPagination returns the pagination asasociated with this payload, if any
func (p *Payload) GetPagination() *query.PageRequest {
	return &query.PageRequest{
		Offset:     p.Input.Offset,
		Limit:      p.Input.Limit,
		CountTotal: p.Input.CountTotal,
	}
}

type PayloadArgs struct {
	Address         string    `json:"address"`
	IdentityKey     string    `json:"identityKey"`
	Height          int64     `json:"height"`
	Offset          uint64    `json:"offset"`
	Limit           uint64    `json:"limit"`
	CountTotal      bool      `json:"count_total"`
	ExecutedAtStart time.Time `json:"executedAtStart"`
	ExecutedAtEnd   time.Time `json:"executedAtEnd"`
}
