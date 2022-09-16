package wasm

import (
	"encoding/base64"
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v3/types"
	juno "github.com/forbole/juno/v3/types"
	"github.com/shopspring/decimal"
)

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	log.Trace().Str("txhash", tx.TxHash).Msg("wasm HandleMsg")

	if len(tx.Logs) == 0 {
		return nil
	}

	switch cosmosMsg := msg.(type) {
	case *wasmtypes.MsgStoreCode:
		err := m.HandleMsgStoreCode(index, tx, cosmosMsg)
		if err != nil {
			return fmt.Errorf("error while handling MsgStoreCode: %s", err)
		}
	case *wasmtypes.MsgInstantiateContract:
		err := m.HandleMsgInstantiateContract(index, tx, cosmosMsg)
		if err != nil {
			return fmt.Errorf("error while handling MsgInstantiateContract: %s", err)
		}
	case *wasmtypes.MsgExecuteContract:
		err := m.HandleMsgExecuteContract(index, tx, cosmosMsg)
		if err != nil {
			return fmt.Errorf("error while handling MsgExecuteContract: %s", err)
		}
	case *wasmtypes.MsgMigrateContract:
		err := m.HandleMsgMigrateContract(index, tx, cosmosMsg)
		if err != nil {
			return fmt.Errorf("error while handling MsgMigrateContract: %s", err)
		}
	case *wasmtypes.MsgUpdateAdmin:
		err := m.HandleMsgUpdateAdmin(cosmosMsg)
		if err != nil {
			return fmt.Errorf("error while handling MsgUpdateAdmin: %s", err)
		}
	case *wasmtypes.MsgClearAdmin:
		err := m.HandleMsgClearAdmin(cosmosMsg)
		if err != nil {
			return fmt.Errorf("error while handling MsgClearAdmin: %s", err)
		}
	}

	return nil
}

// HandleMsgStoreCode allows to properly handle a MsgStoreCode
// The Store Code Event is to upload the contract code on the chain, where a Code ID is returned
func (m *Module) HandleMsgStoreCode(index int, tx *juno.Tx, msg *wasmtypes.MsgStoreCode) error {
	// Get store code event
	event, err := tx.FindEventByType(index, wasmtypes.EventTypeStoreCode)
	if err != nil {
		return fmt.Errorf("error while searching for EventTypeInstantiate: %s", err)
	}

	// Get code ID from store code event
	codeIDKey, err := tx.FindAttributeByKey(event, wasmtypes.AttributeKeyCodeID)
	if err != nil {
		return fmt.Errorf("error while searching for AttributeKeyContractAddr: %s", err)
	}

	codeID, err := strconv.ParseUint(codeIDKey, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing code id to int64: %s", err)
	}

	return m.db.SaveWasmCode(
		types.NewWasmCode(
			msg.Sender, msg.WASMByteCode, msg.InstantiatePermission, codeID, tx.Height,
		),
	)
}

// HandleMsgInstantiateContract allows to properly handle a MsgInstantiateContract
// Instantiate Contract Event instantiates an executable contract with the code previously stored with Store Code Event
func (m *Module) HandleMsgInstantiateContract(index int, tx *juno.Tx, msg *wasmtypes.MsgInstantiateContract) error {
	// Get instantiate contract event
	event, err := tx.FindEventByType(index, wasmtypes.EventTypeInstantiate)
	if err != nil {
		return fmt.Errorf("error while searching for EventTypeInstantiate: %s", err)
	}

	// Get contract address
	contractAddress, err := tx.FindAttributeByKey(event, wasmtypes.AttributeKeyContractAddr)
	if err != nil {
		return fmt.Errorf("error while searching for AttributeKeyContractAddr: %s", err)
	}

	// Get result data
	resultData, err := tx.FindAttributeByKey(event, wasmtypes.AttributeKeyResultDataHex)
	if err != nil {
		resultData = ""
	}
	resultDataBz, err := base64.StdEncoding.DecodeString(resultData)
	if err != nil {
		return fmt.Errorf("error while decoding result data: %s", err)
	}

	// Get the contract info
	contractInfo, err := m.source.GetContractInfo(tx.Height, contractAddress)
	if err != nil {
		return fmt.Errorf("error while getting proposal: %s", err)
	}

	timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("error while parsing time: %s", err)
	}

	// Get contract info extension
	var contractInfoExt string
	if contractInfo.Extension != nil {
		var extentionI wasmtypes.ContractInfoExtension
		err = m.cdc.UnpackAny(contractInfo.Extension, &extentionI)
		if err != nil {
			return fmt.Errorf("error while getting contract info extension: %s", err)
		}
		contractInfoExt = extentionI.String()
	}

	// Get contract states
	contractStates, err := m.source.GetContractStates(tx.Height, contractAddress)
	if err != nil {
		return fmt.Errorf("error while getting genesis contract states: %s", err)
	}

	contract := types.NewWasmContract(
		msg.Sender, msg.Admin, msg.CodeID, msg.Label, msg.Msg, msg.Funds,
		contractAddress, string(resultDataBz), timestamp,
		contractInfo.Creator, contractInfoExt, contractStates, tx.Height,
	)
	return m.db.SaveWasmContracts(
		[]types.WasmContract{contract},
	)
}

// HandleMsgExecuteContract allows to properly handle a MsgExecuteContract
// Execute Event executes an instantiated contract
func (m *Module) HandleMsgExecuteContract(index int, tx *juno.Tx, msg *wasmtypes.MsgExecuteContract) error {
	log.Trace().Str("txhash", tx.TxHash).Msg("wasm HandleMsgExecuteContract")

	//
	// parse the ExecuteContract message body
	//
	msgJson, err := oj.ParseString(string(msg.Msg))

	// use reflection to get the message name by pulling the 1st field name from the JSON struct
	messageName := ""
	v := reflect.ValueOf(msgJson)
	if v.Len() == 1 && len(v.MapKeys()) == 1 {
		messageName = v.MapKeys()[0].String()
	} else {
		log.Warn().Str("txhash", tx.TxHash).Str("messageJson", string(msg.Msg)).Msg("Unable to parse message name from JSON")
	}

	// skip some message types:
	// - `write_k_v` will bloat the pgsql database with contract state imports
	if messageName == "write_k_v" {
		log.Trace().Str("txhash", tx.TxHash).Str("messageName", messageName).Msg("Skipping contract message")
		return err
	}

	log.Debug().Int64("height", tx.Height).Str("txhash", tx.TxHash).Str("messageName", messageName).Msg("Processing contract message")

	// Get Execute Contract event
	event, err := tx.FindEventByType(index, wasmtypes.EventTypeExecute)
	if err != nil {
		return fmt.Errorf("error while searching for EventTypeExecute: %s", err)
	}

	// Get result data
	resultData, err := tx.FindAttributeByKey(event, wasmtypes.AttributeKeyResultDataHex)
	if err != nil {
		resultData = ""
	}
	resultDataBz, err := base64.StdEncoding.DecodeString(resultData)
	if err != nil {
		return fmt.Errorf("error while decoding result data: %s", err)
	}

	timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("error while parsing time: %s", err)
	}

	contractExists, err := m.db.GetWasmContractExists(msg.Contract)

	if !contractExists {
		contractAddress := msg.Contract

		// default values
		contractInfoCreator := "unknown"
		contractInfoAdmin := "unknown"
		contractInfoCodeID := uint64(0)
		contractInfoLabel := ""

		// Check if there is a record of the contract, otherwise look it up
		contractInfo, err := m.source.GetContractInfo(tx.Height, contractAddress)
		if err != nil {
			log.Trace().Str("contractAddress", contractAddress).Int64("height", tx.Height).Msg("Unable to get contract info, using default values...")
		} else {
			contractInfoCreator = contractInfo.Creator
			contractInfoAdmin = contractInfo.Admin
			contractInfoCodeID = contractInfo.CodeID
			contractInfoLabel = contractInfo.Label
		}

		createdBlockHeight := int64(0)
		if contractInfo != nil && contractInfo.Created != nil {
			createdBlockHeight = int64(contractInfo.Created.BlockHeight)
		}

		emptyBytes := make([]byte, 0)
		var initPermission wasmtypes.AccessConfig
		newCode := types.NewWasmCode(
			contractInfoCreator, emptyBytes, &initPermission, contractInfoCodeID, createdBlockHeight,
		)

		err = m.db.SaveWasmCode(newCode)
		if err != nil {
			return fmt.Errorf("error while saving contract code: %s", err)
		}

		// Get contract info extension
		contractInfoExt := ""
		if contractInfo != nil && contractInfo.Extension != nil {
			var extentionI wasmtypes.ContractInfoExtension
			err = m.cdc.UnpackAny(contractInfo.Extension, &extentionI)
			if err != nil {
				return fmt.Errorf("error while getting contract info extension: %s", err)
			}
			contractInfoExt = extentionI.String()
		}

		// Get contract states (at the height the contract was created)
		var contractStates []wasmtypes.Model
		//contractStates, err := m.source.GetContractStates(int64(contractInfo.Created.BlockHeight), contractAddress)
		//if err != nil {
		//	return fmt.Errorf("error while getting genesis contract states: %s", err)
		//}

		err = m.db.SaveAccounts([]types.Account{
			types.NewAccount(msg.Sender)})
		if err != nil {
			log.Debug().Msg(fmt.Errorf("error while saving Sender account %s: %s", msg.Sender, err).Error())
		}
		err = m.db.SaveAccounts([]types.Account{
			types.NewAccount(contractInfoAdmin)})
		if err != nil {
			log.Debug().Msg(fmt.Errorf("error while saving Admin account %s: %s", contractInfo.Admin, err).Error())
		}

		// Set to default values, that will hopefully be overwritten during the next migration of this contract
		emptyRawMessage := []byte("{}")
		emptyFunds := sdk.Coins{sdk.Coin{}}

		contract := types.NewWasmContract(
			msg.Sender, contractInfoAdmin, contractInfoCodeID, contractInfoLabel,
			emptyRawMessage, emptyFunds,
			contractAddress, string(resultDataBz), timestamp,
			contractInfoCreator, contractInfoExt, contractStates, createdBlockHeight,
		)

		err = m.db.SaveWasmContracts(
			[]types.WasmContract{contract},
		)
		if err != nil {
			return fmt.Errorf("error while saving contract info: %s", err)
		}
	}

	execute := types.NewWasmExecuteContract(
		msg.Sender, msg.Contract, msg.Msg, msg.Funds,
		string(resultDataBz), timestamp, tx.Height, tx.TxHash,
	)

	// save a record of the raw contract execution details
	err = m.db.SaveWasmExecuteContract(execute)

	// save a row for each event in the contract execution
	err = m.db.SaveWasmExecuteContractEvents(execute, tx)
	if err != nil {
		log.Err(err).Msg("Could not save events for WasmExecuteContract")
	}

	// process Nym Mixnet v1 messages
	err = m.handleMessageNymMixnetV1(msgJson, messageName, msg, execute, tx)

	//panic("lets stop after one message")

	return err
}

func (m *Module) handleMessageNymMixnetV1(msgJson interface{}, messageName string, msg *wasmtypes.MsgExecuteContract, execute types.WasmExecuteContract, tx *juno.Tx) error {
	var err error

	// extract identity keys from the body of the contract execution message, and keep the last one found
	identityKey := ""
	matches := matchJsonPathIdentityKey(msgJson)
	for _, v := range matches {
		identityKey = fmt.Sprint(v)
		err = m.db.EnsureExistsNymMixnetV1Mixnode(types.NewMixnodeV1(identityKey, true, types.Inactive))
	}

	proxy := matchJsonPathProxy(msgJson)
	actor := getActorFromMessageName(messageName)
	amount := msg.Funds
	data := string(msg.Msg)

	// don't write the system actor events, because they are noisy and are already captured in the table `wasm_execute_contract_event`
	if !strings.HasPrefix(actor, "system") {
		err = m.db.EnsureExistsNymMixnetV1Mixnode(types.NewMixnodeV1(identityKey, true, types.Inactive))
		err = m.db.SaveNymMixnetV1MixnodeEvent(
			messageName, actor, proxy, identityKey, &amount, messageName, data, execute, tx)
		if err != nil {
			log.Err(err).Msg("Error while saving mixnode event")
		}
	}

	switch messageName {
	case "reward_mixnode":
		// try v2 first, then fallback to v1
		_, err = m.tryHandleMixnetV2RewardMessage(msg, execute, tx)
		if err != nil {
			_, err = m.tryHandleMixnetV1RewardMessage(msg, execute, tx)
		}
	}

	for _, txLog := range tx.Logs {
		for _, event := range txLog.Events {

			switch event.Type {
			case "wasm-v2_mix_rewarding":
				mixId := getValueFromNamedAttribute(event.Attributes, "mix_id")
				operatorRewardAttr := getValueFromNamedAttribute(event.Attributes, "operator_reward")
				delegatesRewardAttr := getValueFromNamedAttribute(event.Attributes, "delegates_reward")
				priorDelegatesAttr := getValueFromNamedAttribute(event.Attributes, "prior_delegates")
				priorUnitDelegationAttr := getValueFromNamedAttribute(event.Attributes, "prior_unit_delegation")

				var mixIdInt *uint32
				if mixId != nil {
					value, _ := strconv.ParseUint(*mixId, 10, 32)
					value2 := uint32(value)
					mixIdInt = &value2
				}

				if mixId != nil && mixIdInt != nil && operatorRewardAttr != nil && delegatesRewardAttr != nil && priorDelegatesAttr != nil && priorUnitDelegationAttr != nil {

					err = m.db.EnsureExistsNymMixnetV2Mixnode(types.NewMixnodeV2(*mixIdInt, "", true, types.MixnodeStatus(types.InRewardedSet)))
					if err != nil {
						log.Err(err).Str("msg", string(msg.Msg)).Msg("Error while saving mixnode")
						continue
					}

					// check if processed
					hasRow, err := m.db.HasNymMixnetV2MixnodeRewardingEvent(*mixIdInt, tx)
					if err != nil {
						log.Err(err).Msg("Failed to check if row exists in nyx_nym_mixnet_v2_mixnode_reward")
					}
					if hasRow {
						log.Trace().Uint32("mixIdInt", *mixIdInt).Str("hash", tx.TxHash).Int64("height", tx.Height).Msg("Skipping...")
						continue
					}

					operatorRewardCoin := sdk.NewCoins(sdk.NewCoin("unym", sdk.NewInt(decimal.RequireFromString(*operatorRewardAttr).IntPart())))
					delegatesRewardCoin := sdk.NewCoins(sdk.NewCoin("unym", sdk.NewInt(decimal.RequireFromString(*delegatesRewardAttr).IntPart())))
					priorDelegatesCoin := sdk.NewCoins(sdk.NewCoin("unym", sdk.NewInt(decimal.RequireFromString(*priorDelegatesAttr).IntPart())))
					priorUnitDelegationCoin := decimal.RequireFromString(*priorUnitDelegationAttr)

					// TODO: calculate
					apy := decimal.Zero

					err = m.db.SaveNymMixnetV2MixnodeRewardingEvent(*mixIdInt, operatorRewardCoin, delegatesRewardCoin, priorDelegatesCoin, priorUnitDelegationCoin, apy.InexactFloat64(), event, execute, tx)
					return err
				}

			case "wasm-mix_rewarding":

				eventIdentityKey := getValueFromNamedAttribute(event.Attributes, "identity")
				pledge := getValueFromNamedAttribute(event.Attributes, "pledge")
				delegated := getValueFromNamedAttribute(event.Attributes, "delegated")
				totalNodeRewardAttr := getValueFromNamedAttribute(event.Attributes, "total_node_reward")

				// C_m (in uNYM) is fixed 40 NYMs per 720 epochs = 40_000_000 / 30 / 24
				costPerEpoch := decimal.NewFromInt(55_556)

				// get the profit margin for the epoch
				nodeProfitMargin := decimal.NewFromInt(0)

				sigmaAttr := getValueFromNamedAttribute(event.Attributes, "sigma")
				lambdaAttr := getValueFromNamedAttribute(event.Attributes, "lambda")

				if eventIdentityKey != nil && pledge != nil && delegated != nil && totalNodeRewardAttr != nil && sigmaAttr != nil && lambdaAttr != nil {

					// check if processed
					hasRow, err := m.db.HasNymMixnetV1MixnodeRewardingEvent(*eventIdentityKey, tx)
					if err != nil {
						log.Err(err).Msg("Failed to check if row exists in nyx_nym_mixnet_v1_mixnode_reward")
					}
					if hasRow {
						log.Trace().Str("identityKey", *eventIdentityKey).Str("hash", tx.TxHash).Int64("height", tx.Height).Msg("Skipping...")
						continue
					}

					// lookup the profit margin from the snapshot
					profitMargin, err := m.source.GetNymMixnetContractV1MixnodeProfitMargin(tx.Height, execute.ContractAddress, *eventIdentityKey, m.cdc)
					if err != nil {
						log.Err(err).Str("eventIdentityKey", *eventIdentityKey).Int64("height", tx.Height).Msg("Failed to get profit margin")
					}
					if profitMargin != nil {
						nodeProfitMargin = decimal.NewFromInt(int64(*profitMargin)).Div(decimal.NewFromInt(100))
					}

					sigma, errSigma := decimal.NewFromString(*sigmaAttr)
					lambda, errLambda := decimal.NewFromString(*lambdaAttr)
					totalNodeReward, errTNR := decimal.NewFromString(*totalNodeRewardAttr)

					if errSigma != nil || errLambda != nil || errTNR != nil {
						log.Err(errSigma).Str("sigma", *sigmaAttr).Msg("Error parsing sigma")
						log.Err(errLambda).Str("lambda", *lambdaAttr).Msg("Error parsing lambda")
						log.Err(errTNR).Str("totalNodeReward", *totalNodeRewardAttr).Msg("Error parsing total node reward")
					} else {
						// prevent division by zero
						if sigma.IsZero() {
							log.Warn().Str("sigma", *sigmaAttr).Msg("Skipping to avoid division by zero...")
							continue
						}

						f1 := decimal.NewFromInt(1).Sub(nodeProfitMargin)
						f2 := totalNodeReward.Sub(costPerEpoch)
						operatorReward := nodeProfitMargin.Add(f1.Mul(lambda.Div(sigma))).Mul(f2)                                  // (u_i + (1 - ui).(l/s)).(R_i - CM_i)
						operatorReward = decimal.Min(costPerEpoch, totalNodeReward).Add(decimal.Max(decimal.Zero, operatorReward)) // min(CM_i, R_i) + max(0, ...)

						totalDelegationsInt, _ := sdk.NewIntFromString(*delegated)
						totalPledgeInt, _ := sdk.NewIntFromString(*pledge)

						totalBond := decimal.NewFromInt(totalPledgeInt.Int64()).Add(decimal.NewFromInt(totalDelegationsInt.Int64()))
						stakingSupply := totalBond.Div(sigma)

						// prevent division by zero
						if stakingSupply.IsZero() {
							log.Warn().Int64("totalDelegationsInt", totalDelegationsInt.Int64()).Int64("totalPledgeInt", totalPledgeInt.Int64()).Msg("Skipping to avoid division by zero in stakingSupply...")
							continue
						}
						if totalDelegationsInt.IsZero() {
							log.Warn().Int64("totalDelegationsInt", totalDelegationsInt.Int64()).Msg("Skipping to avoid division by zero in totalDelegationsInt...")
							continue
						}

						totalDelegationsCoin := sdk.NewCoins(sdk.NewCoin("unym", totalDelegationsInt))

						delegatorsReward := decimal.Max(decimal.Zero, totalNodeReward.Sub(operatorReward))
						unitDelegatorReward := decimal.Zero
						apy := decimal.Zero
						if !delegatorsReward.IsZero() {
							unitDelegatorReward = delegatorsReward.Mul(decimal.NewFromInt(1_000_000_000_000)).Div(decimal.NewFromInt(totalDelegationsInt.Int64()))
							apy = unitDelegatorReward.Mul(decimal.NewFromInt(24)).Mul(decimal.NewFromInt(365)).Div(decimal.NewFromInt(1_000_000_000_000))
						}

						totalNodeRewardCoin := sdk.NewCoins(sdk.NewCoin("unym", sdk.NewInt(totalNodeReward.RoundBank(6).IntPart())))
						operatorRewardCoin := sdk.NewCoins(sdk.NewCoin("unym", sdk.NewInt(operatorReward.RoundBank(6).IntPart())))
						stakingSupplyCoin := sdk.NewCoins(sdk.NewCoin("unym", sdk.NewInt(stakingSupply.RoundBank(6).IntPart())))

						profitMarginPercentage := int(nodeProfitMargin.Mul(decimal.NewFromInt(100)).IntPart())

						err = m.db.EnsureExistsNymMixnetV1Mixnode(types.NewMixnodeV1(*eventIdentityKey, true, types.Inactive))

						log.Info().Str("identityKey", *eventIdentityKey).Uint8("profit", *profitMargin).Int64("height", tx.Height).Msg("Saving mixnode rewarding event")

						err = m.db.SaveNymMixnetV1MixnodeRewardingEvent(
							*eventIdentityKey, totalNodeRewardCoin, totalDelegationsCoin, operatorRewardCoin, unitDelegatorReward, apy.InexactFloat64(), stakingSupplyCoin, profitMarginPercentage,
							event, execute, tx)

						if err != nil {
							log.Error().Err(err).Msg("❌ Oh no")
						} else {
							log.Trace().Int64("height", tx.Height).Str("hash", tx.TxHash).Msg("✅ Saved rewarding event")
						}
					}
				}
			}
		}
	}
	return err
}

func (m *Module) tryHandleMixnetV1RewardMessage(msg *wasmtypes.MsgExecuteContract, execute types.WasmExecuteContract, tx *juno.Tx) (*types.MixnetV1RewardMixnodeMessage, error) {
	rewardMixnodeMessage, err := types.ParseMixnetV1RewardMixnodeMessage(msg.Msg.Bytes())
	if rewardMixnodeMessage != nil {
		rewardedIdentityKey := rewardMixnodeMessage.RewardMixnode.Identity
		status := "in_active_set"
		statusType := types.MixnodeStatus(types.InActiveSet)
		if !rewardMixnodeMessage.RewardMixnode.Params.InActiveSet {
			status = "in_standby_set"
			statusType = types.MixnodeStatus(types.InStandbySet)
		}
		uptime, err := strconv.Atoi(rewardMixnodeMessage.RewardMixnode.Params.Uptime)
		if err != nil {
			log.Err(err).Str("json", string(msg.Msg.Bytes())).Msg("Unable to parse uptime")
		} else {
			log.Trace().Str("identity_key", rewardedIdentityKey).Msg("✅ Saving mixnode status")
			err = m.db.EnsureExistsNymMixnetV1Mixnode(types.NewMixnodeV1(rewardedIdentityKey, true, statusType))
			if err != nil {
				log.Err(err).Str("msg", string(msg.Msg)).Msg("Error while saving mixnode")
			}
			err = m.db.SaveNymMixnetV1MixnodeStatus(rewardedIdentityKey, status, uptime, execute, tx)
			if err != nil {
				log.Err(err).Str("msg", string(msg.Msg)).Msg("Error while saving mixnode status")
			}
		}
	}
	return rewardMixnodeMessage, err
}

func (m *Module) tryHandleMixnetV2RewardMessage(msg *wasmtypes.MsgExecuteContract, execute types.WasmExecuteContract, tx *juno.Tx) (*types.MixnetV2MessageRewardMixnode, error) {
	rewardMixnodeMessage, err := types.ParseMixnetV2MessageRewardMixnode(msg.Msg.Bytes())
	if rewardMixnodeMessage != nil {
		mixId := rewardMixnodeMessage.RewardMixnode.MixId
		status := "in_rewarded_set"
		statusType := types.MixnodeStatus(types.InRewardedSet)
		performance, err := strconv.ParseFloat(rewardMixnodeMessage.RewardMixnode.Performance, 32)
		if err != nil {
			log.Err(err).Str("json", string(msg.Msg.Bytes())).Msg("Unable to parse performance")
		} else {
			log.Trace().Int64("height", tx.Height).Str("hash", tx.TxHash).Uint32("mix_id", mixId).Msg("✅ Saving mixnode status")
			// TODO: v2
			err = m.db.EnsureExistsNymMixnetV2Mixnode(types.NewMixnodeV2(mixId, "", true, statusType))
			if err != nil {
				log.Err(err).Str("msg", string(msg.Msg)).Msg("Error while saving mixnode")
			}
			performanceInt := int64(performance * 100)
			err = m.db.SaveNymMixnetV2MixnodeStatus(mixId, status, performanceInt, execute, tx)
			if err != nil {
				log.Err(err).Str("msg", string(msg.Msg)).Msg("Error while saving mixnode status")
			}
		}
	}
	return rewardMixnodeMessage, err
}

func getActorFromMessageName(messageName string) string {
	switch messageName {

	//
	// MIXNET v1
	//

	// operator messages
	case "compound_operator_reward_on_behalf",
		"compound_operator_reward",
		"bond_mixnode",
		"unbond_mixnode",
		"update_mixnode_config",
		"update_mixnode_config_on_behalf",
		"bond_gateway",
		"unbond_gateway",
		"bond_mixnode_on_behalf",
		"unbond_mixnode_on_behalf",
		"bond_gateway_on_behalf",
		"unbond_gateway_on_behalf":
		return "operator"

	// delegator messages
	case "compound_delegator_reward_on_behalf",
		"compound_delegator_reward",
		"delegate_to_mixnode",
		"undelegate_from_mixnode",
		"delegate_to_mixnode_on_behalf",
		"undelegate_from_mixnode_on_behalf":
		return "delegator"

	// system messages
	case "update_rewarding_validator_address",
		"reconcile_delegations",
		"checkpoint_mixnodes",
		"reward_mixnode",
		"write_rewarded_set",
		"advance_current_epoch",
		"init_epoch":
		return "system.rewarding"

	//
	// VESTING
	//

	case "create_account":
		return "admin.vesting"

	case "withdraw_vested_coins",
		"transfer_ownership",
		"update_staking_address":
		return "vesting_account"

	case
		"update_mixnet_address":
		return "operator"
	case
		"track_unbond_mixnode",
		"track_unbond_gateway",
		"track_undelegation":
		return "system.vesting"

	//
	// MIXNET v2
	//

	case "update_contract_state_params",
		"update_active_set_size",
		"update_rewarding_params",
		"update_interval_config":
		return "admin.mixnet_v2"

	case "reconcile_epoch_events":
		return "system.mixnet_v2"

	case "update_mixnode_cost_params",
		"update_mixnode_cost_params_on_behalf",
		"withdraw_operator_reward",
		"withdraw_operator_reward_on_behalf":
		return "operator"

	case "withdraw_delegator_reward",
		"withdraw_delegator_reward_on_behalf":
		return "delegator"

	case "testing_resolve_all_pending_events":
		return "qa"
	}

	log.Warn().Str("messageName", messageName).Msg("Unable to determine actor from message name")
	return ""
}

func getValueFromNamedAttribute(attributes []sdk.Attribute, key string) *string {
	for _, attr := range attributes {
		if attr.Key == key {
			return &attr.Value
		}
	}
	return nil
}

// HandleMsgMigrateContract allows to properly handle a MsgMigrateContract
// Migrate Contract Event upgrade the contract by updating code ID generated from new Store Code Event
func (m *Module) HandleMsgMigrateContract(index int, tx *juno.Tx, msg *wasmtypes.MsgMigrateContract) error {
	// Get Migrate Contract event
	event, err := tx.FindEventByType(index, wasmtypes.EventTypeMigrate)
	if err != nil {
		return fmt.Errorf("error while searching for EventTypeMigrate: %s", err)
	}

	// Get result data
	resultData, err := tx.FindAttributeByKey(event, wasmtypes.AttributeKeyResultDataHex)
	if err != nil {
		resultData = ""
	}
	resultDataBz, err := base64.StdEncoding.DecodeString(resultData)
	if err != nil {
		return fmt.Errorf("error while decoding result data: %s", err)
	}

	return m.db.UpdateContractWithMsgMigrateContract(msg.Sender, msg.Contract, msg.CodeID, msg.Msg, string(resultDataBz))
}

// HandleMsgUpdateAdmin allows to properly handle a MsgUpdateAdmin
// Update Admin Event updates the contract admin who can migrate the wasm contract
func (m *Module) HandleMsgUpdateAdmin(msg *wasmtypes.MsgUpdateAdmin) error {
	return m.db.UpdateContractAdmin(msg.Sender, msg.Contract, msg.NewAdmin)
}

// HandleMsgClearAdmin allows to properly handle a MsgClearAdmin
// Clear Admin Event clears the admin which make the contract no longer migratable
func (m *Module) HandleMsgClearAdmin(msg *wasmtypes.MsgClearAdmin) error {
	return m.db.UpdateContractAdmin(msg.Sender, msg.Contract, "")
}

var xIdentityKey jp.Expr

func matchJsonPathIdentityKey(data interface{}) (results []interface{}) {
	if xIdentityKey == nil {
		parsed, err := jp.ParseString(`$.*['identity','mix_identity','identity_key']`)
		xIdentityKey = parsed
		if err != nil {
			log.Err(err).Msg("failed to parse json path for identity key")
		}
	}
	return xIdentityKey.Get(data)
}

var xProxy jp.Expr

func matchJsonPathProxy(data interface{}) *string {
	if xProxy == nil {
		parsed, err := jp.ParseString(`$.*['proxy']`)
		xProxy = parsed
		if err != nil {
			log.Err(err).Msg("failed to parse json path for proxy")
		}
	}
	matches := xProxy.Get(data)
	if len(matches) > 0 {
		val := fmt.Sprint(matches[0])
		return &val
	}
	return nil
}

var xMixId jp.Expr

func matchJsonPathMixId(data interface{}) *uint32 {
	if xMixId == nil {
		parsed, err := jp.ParseString(`$.*['mix_id']`)
		xMixId = parsed
		if err != nil {
			log.Err(err).Msg("failed to parse json path for mix id")
		}
	}
	value, err := strconv.Atoi(fmt.Sprint(xMixId.Get(data)))
	if err != nil {
		log.Err(err).Msg("failed to parse mix id from string value")
		return nil
	}
	value2 := uint32(value)
	return &value2
}
