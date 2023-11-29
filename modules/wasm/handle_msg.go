package wasm

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ohler55/ojg/oj"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v4/types"
	juno "github.com/forbole/juno/v5/types"
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
	msgJSON, err := oj.ParseString(string(msg.Msg))

	// use reflection to get the message name by pulling the 1st field name from the JSON struct
	messageName := ""
	v := reflect.ValueOf(msgJSON)
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

	contractExists, _ := m.db.GetWasmContractExists(msg.Contract)

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
		// contractStates, err := m.source.GetContractStates(int64(contractInfo.Created.BlockHeight), contractAddress)
		// if err != nil {
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
	if err != nil {
		log.Err(err).Msg("Could not save WasmExecuteContract")
	}

	// save a row for each event in the contract execution
	err = m.db.SaveWasmExecuteContractEvents(execute, tx)
	if err != nil {
		log.Err(err).Msg("Could not save events for WasmExecuteContract")
	}

	return err
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
