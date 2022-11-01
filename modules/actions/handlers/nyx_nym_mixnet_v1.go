package handlers

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbtypes "github.com/forbole/bdjuno/v3/database/types"
	"github.com/forbole/bdjuno/v3/modules/actions/types"
	"github.com/rs/zerolog/log"
	"time"
)

func NyxNymMixnetV1DelegationsHandler(ctx *types.Context, payload *types.Payload) (interface{}, error) {
	address := payload.GetAddress()
	identityKey := payload.GetIdentityKey()
	height := payload.GetHeight()

	log.Debug().Str("address", address).
		Str("identity_key", *identityKey).
		Interface("height", height).
		Msg("executing NyxNymMixnetV1DelegationsHandler action")

	if identityKey == nil {
		return nil, fmt.Errorf("identity key not specified")
	}

	delegations, err := getDelegations(ctx, *identityKey, address, height)

	log.Debug().Interface("delegations", delegations).Msg("Got delegations")

	return delegations, err
}

func NyxNymMixnetV1RewardsHandler(ctx *types.Context, payload *types.Payload) (interface{}, error) {
	address := payload.GetAddress()
	identityKey := payload.GetIdentityKey()
	height := payload.GetHeight()

	log.Debug().Str("address", address).
		Str("identity_key", *identityKey).
		Interface("height", height).
		Msg("executing NyxNymMixnetV1RewardsHandler action")

	if identityKey == nil {
		return nil, fmt.Errorf("identity key not specified")
	}

	delegations, err := getDelegations(ctx, *identityKey, address, height)
	if err != nil {
		return nil, err
	}

	rewards := make([]types.NyxNymMixnetV1Rewards, len(delegations))

	for i, delegation := range delegations {
		var heightEnd uint64

		if delegation.End != nil {
			heightEnd = delegation.End.Height
		}

		log.Debug().Str("identity", *identityKey).Uint64("start", delegation.Start.Height).Uint64("end", heightEnd).Msg("Getting rewards")

		rewardEventsDb, err := ctx.Db.GetNymMixnetV1MixnodeRewardEvent(*identityKey, delegation.Start.Height, &heightEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to get reward events: %s", err)
		}

		rewardEvents := make([]types.NyxNymMixnetV1Reward, len(rewardEventsDb))
		totalRewards := sdk.NewDec(0)
		delegationDec := sdk.MustNewDecFromStr(delegation.Delegation.Amount)
		for j, event := range rewardEventsDb {
			totalNodeReward := event.TotalNodeReward.ToCoins()[0]
			unitDelegatorReward := sdk.NewDec(int64(event.UnitDelegatorReward)).Quo(sdk.NewDec(1_000_000_000_000))

			//reward := decimal.NewFromInt(int64(event.UnitDelegatorReward)).Mul(decimal.RequireFromString(delegation.Delegation.Amount)).Div(decimal.NewFromInt(1_000_000_000_000))
			reward := delegationDec.Mul(unitDelegatorReward)
			rewardAsInt := reward.RoundInt64()
			totalRewards = totalRewards.Add(sdk.NewDec(rewardAsInt))

			rewardEvents[j] = types.NyxNymMixnetV1Reward{
				Timestamp: event.ExecutedAt,
				Height:    uint64(event.Height),
				TotalNodeReward: types.Coin{
					Amount: totalNodeReward.Amount.String(),
					Denom:  totalNodeReward.Denom,
				},
				Reward: types.Coin{
					Amount: reward.RoundInt().String(),
					Denom:  totalNodeReward.Denom,
				},
				EpochApy: event.Apy,
			}
		}

		endTS := time.Now()
		if delegation.End != nil {
			endTS = delegation.End.Timestamp
		}
		duration := endTS.Sub(delegation.Start.Timestamp)
		durationSecs := int64(duration.Seconds())
		returnPerSecond := totalRewards.Quo(delegationDec).QuoInt64(durationSecs)
		returnPerYear := returnPerSecond.MulInt64(365 * 24 * 60 * 60)

		rewards[i] = types.NyxNymMixnetV1Rewards{
			DelegatorAddress:   delegation.DelegatorAddress,
			MixnodeIdentityKey: delegation.MixnodeIdentityKey,
			Start:              delegation.Start,
			End:                delegation.End,
			Delegation:         delegation.Delegation,
			Rewards:            rewardEvents,
			TotalRewards: types.Coin{
				Amount: totalRewards.Quo(sdk.NewDec(1_000_000)).String(),
				Denom:  "nym",
			},
			APY: returnPerYear.MustFloat64(),
		}
	}

	return rewards, nil
}

func getDelegations(ctx *types.Context, identityKey string, address string, height *int64) ([]types.NyxNymMixnetV1Delegation, error) {
	dbDelegations, err := ctx.Db.GetNymMixnetV1MixnodeEvent("delegate_to_mixnode", identityKey, &address, height, nil)
	if err != nil {
		return nil, fmt.Errorf("error while getting event rows: %s", err)
	}

	log.Debug().Str("identityKey", identityKey).Str("address", address).Int64("height", *height).Int("count delegations", len(dbDelegations)).Msg("Got delegations")

	dbUndelegations, err := ctx.Db.GetNymMixnetV1MixnodeEvent("undelegation", identityKey, &address, height, nil)
	if err != nil {
		return nil, fmt.Errorf("error while getting event rows: %s", err)
	}

	log.Debug().Int("count undelegations", len(dbUndelegations)).Msg("Got undelegations")

	delegations := make([]types.NyxNymMixnetV1Delegation, len(dbDelegations))

	for i, delegation := range dbDelegations {
		undelegation, j := contains(dbUndelegations, delegation.IdentityKey)

		// undelegation must occur after delegation
		if undelegation != nil && undelegation.Height <= delegation.Height {
			dbUndelegations = remove(dbUndelegations, j)
			undelegation = nil
		}

		amountCoins := delegation.Amount.ToCoins()
		amount := "0"
		denom := "unym"

		if len(amount) > 0 {
			amount = amountCoins[0].Amount.String()
			denom = amountCoins[0].Denom
		} else {
			log.Warn().Interface("delegation", delegation).Msg("Zero delegation")
		}

		delegations[i] = types.NyxNymMixnetV1Delegation{
			DelegatorAddress:   delegation.Sender,
			MixnodeIdentityKey: delegation.IdentityKey,
			Start: types.NyxNymMixnetV1DelegationTimestamp{
				Timestamp: delegation.ExecutedAt,
				Height:    uint64(delegation.Height),
			},
			End: nil,
			Delegation: types.Coin{
				Amount: amount,
				Denom:  denom,
			},
		}
		if undelegation != nil {
			delegations[i].End = &types.NyxNymMixnetV1DelegationTimestamp{
				Timestamp: undelegation.ExecutedAt,
				Height:    uint64(undelegation.Height),
			}
			dbUndelegations = remove(dbUndelegations, j)
		}
	}

	return delegations, nil
}

func remove(slice []dbtypes.NyxNymMixnetV1MixnodeEventsRow, indexToRemove int) []dbtypes.NyxNymMixnetV1MixnodeEventsRow {
	return append(slice[:indexToRemove], slice[indexToRemove+1:]...)
}

func contains(arr []dbtypes.NyxNymMixnetV1MixnodeEventsRow, identityKey string) (*dbtypes.NyxNymMixnetV1MixnodeEventsRow, int) {
	for i, item := range arr {
		if item.IdentityKey == identityKey {
			return &item, i
		}
	}
	return nil, 0
}
