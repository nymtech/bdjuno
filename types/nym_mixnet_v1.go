package types

import (
	"encoding/json"
)

type MixnodeStatus int

const (
	InActiveSet = iota
	InStandbySet
	Inactive
	InRewardedSet
)

func (status MixnodeStatus) String() string {
	return []string{"in_active_set", "in_standby_set", "inactive", "in_rewarded_set"}[status]
}

type MixnodeV1 struct {
	IdentityKey      string
	IsBonded         bool
	LastMixnetStatus MixnodeStatus
}

func NewMixnodeV1(
	identityKey string,
	isBonded bool,
	lastMixnetStatus MixnodeStatus,
) MixnodeV1 {
	return MixnodeV1{
		IdentityKey:      identityKey,
		IsBonded:         isBonded,
		LastMixnetStatus: lastMixnetStatus,
	}
}

type GatewayV1 struct {
	IdentityKey string
	IsBonded    bool
}

func NewGatewayV1(
	identityKey string,
	isBonded bool,
) GatewayV1 {
	return GatewayV1{
		IdentityKey: identityKey,
		IsBonded:    isBonded,
	}
}

type MixnetV1RewardMixnodeMessage struct {
	RewardMixnode struct {
		Params struct {
			Uptime           string `json:"uptime"`
			InActiveSet      bool   `json:"in_active_set"`
			RewardBlockstamp int    `json:"reward_blockstamp"`
		} `json:"params"`
		Identity string `json:"identity"`
	} `json:"reward_mixnode"`
}

func ParseMixnetV1RewardMixnodeMessage(data []byte) (*MixnetV1RewardMixnodeMessage, error) {
	var result *MixnetV1RewardMixnodeMessage
	err := json.Unmarshal(data, &result)
	return result, err
}
