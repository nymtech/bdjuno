package types

import (
	"github.com/cosmos/cosmos-sdk/types/query"
	"time"
)

// ========================= Delegation Response =========================

type NyxNymMixnetV1DelegationResponse struct {
	Delegations []NyxNymMixnetV1Delegation `json:"delegations"`
	Pagination  *query.PageResponse        `json:"pagination"`
}

type NyxNymMixnetV1Delegation struct {
	DelegatorAddress   string                             `json:"delegator_address"`
	MixnodeIdentityKey string                             `json:"mixnode_identity_key"`
	Start              NyxNymMixnetV1DelegationTimestamp  `json:"start"`
	End                *NyxNymMixnetV1DelegationTimestamp `json:"end"`
	Delegation         Coin                               `json:"delegation"`
}

type NyxNymMixnetV1DelegationTimestamp struct {
	Timestamp time.Time `json:"timestamp"`
	Height    uint64    `json:"height"`
}

type NyxNymMixnetV1Reward struct {
	Timestamp       time.Time `json:"timestamp"`
	Height          uint64    `json:"height"`
	TotalNodeReward Coin      `json:"total_node_reward"`
	Reward          Coin      `json:"reward"`
	EpochApy        float64   `json:"epoch_apy"`
}

type NyxNymMixnetV1Rewards struct {
	DelegatorAddress   string                             `json:"delegator_address"`
	MixnodeIdentityKey string                             `json:"mixnode_identity_key"`
	Start              NyxNymMixnetV1DelegationTimestamp  `json:"start"`
	End                *NyxNymMixnetV1DelegationTimestamp `json:"end"`
	Delegation         Coin                               `json:"delegation"`
	Rewards            []NyxNymMixnetV1Reward             `json:"rewards"`
	TotalRewards       Coin                               `json:"total_rewards"`
	APY                float64                            `json:"apy"`
}
