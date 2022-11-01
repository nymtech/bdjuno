package types

import "time"

// NyxNymMixnetV1MixnodeEventsRow represents a single row inside the nyx_nym_mixnet_v1_mixnode_events table
type NyxNymMixnetV1MixnodeEventsRow struct {
	EventKind   string  `db:"event_kind"`
	Actor       string  `db:"actor"`
	Sender      string  `db:"sender"`
	Proxy       *string `db:"proxy"`
	IdentityKey string  `db:"identity_key"`

	ContractAddress string `db:"contract_address"`
	EventType       string `db:"event_type"`
	Hash            string `db:"hash"`

	Attributes interface{} `db:"attributes"`
	ExecutedAt time.Time   `db:"executed_at"`

	Fee    *DbCoins `db:"fee"`
	Amount *DbCoins `db:"amount"`
	Height int64    `db:"height"`
}

// NyxNymMixnetV1MixnodeRewardRow represents a single row inside the nyx_nym_mixnet_v1_mixnode_reward table
type NyxNymMixnetV1MixnodeRewardRow struct {
	Sender      string `db:"sender"`
	IdentityKey string `db:"identity_key"`

	TotalNodeReward        DbCoins `db:"total_node_reward"`
	TotalDelegations       DbCoins `db:"total_delegations"`
	OperatorReward         DbCoins `db:"operator_reward"`
	UnitDelegatorReward    uint64  `db:"unit_delegator_reward"` // TODO: should be a decimal type
	Apy                    float64 `db:"apy"`
	StakingSupply          DbCoins `db:"staking_supply"`
	ProfitMarginPercentage uint64  `db:"profit_margin_percentage"`

	ContractAddress string `db:"contract_address"`
	EventType       string `db:"event_type"`

	Attributes interface{} `db:"attributes"`
	ExecutedAt time.Time   `db:"executed_at"`

	Height int64  `db:"height"`
	Hash   string `db:"hash"`
}
