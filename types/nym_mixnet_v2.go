package types

import "encoding/json"

type MixnodeV2 struct {
	MixId            uint32
	IdentityKey      string
	IsBonded         bool
	LastMixnetStatus MixnodeStatus
}

func NewMixnodeV2(
	mixId uint32,
	identityKey string,
	isBonded bool,
	lastMixnetStatus MixnodeStatus,
) MixnodeV2 {
	return MixnodeV2{
		MixId:            mixId,
		IdentityKey:      identityKey,
		IsBonded:         isBonded,
		LastMixnetStatus: lastMixnetStatus,
	}
}

type MixnetV2MessageRewardMixnode struct {
	RewardMixnode struct {
		MixId       uint32 `json:"mix_id"`
		Performance string `json:"performance"`
	} `json:"reward_mixnode"`
}

func ParseMixnetV2MessageRewardMixnode(data []byte) (*MixnetV2MessageRewardMixnode, error) {
	var result *MixnetV2MessageRewardMixnode
	err := json.Unmarshal(data, &result)
	return result, err
}

type MixnetV2EventRewardMixnode struct {
	MixId               string `json:"mix_id"`
	OperatorReward      string `json:"operator_reward"`
	PriorDelegates      string `json:"prior_delegates"`
	DelegatesReward     string `json:"delegates_reward"`
	IntervalDetails     string `json:"interval_details"`
	ContractAddress     string `json:"_contract_address"`
	PriorUnitDelegation string `json:"prior_unit_delegation"`
}

func ParseMixnetV2EventRewardMixnod(data []byte) (*MixnetV2EventRewardMixnode, error) {
	var result *MixnetV2EventRewardMixnode
	err := json.Unmarshal(data, &result)
	return result, err
}
