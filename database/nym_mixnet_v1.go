package database

import (
	"encoding/json"
	"fmt"
	dbtypes "github.com/forbole/bdjuno/v3/database/types"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	cosmosTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v3/types"
	juno "github.com/forbole/juno/v3/types"
)

// EnsureExistsNymMixnetV1Mixnode ensures a mixnode is in the store
func (db *Db) EnsureExistsNymMixnetV1Mixnode(mixnode types.MixnodeV1) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v1_mixnode(identity_key, is_bonded, last_mixnet_status) 
VALUES ($1, $2, $3) 
ON CONFLICT DO NOTHING
`
	_, err := db.Sql.Exec(stmt,
		mixnode.IdentityKey, mixnode.IsBonded, mixnode.LastMixnetStatus.String(),
	)
	if err != nil {
		return fmt.Errorf("error while ensuring Nym mixnode exists: %s", err)
	}

	return nil
}

// SaveNymMixnetV1Mixnode allows to create or update a mixnode
func (db *Db) SaveNymMixnetV1Mixnode(mixnode types.MixnodeV1) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v1_mixnode(identity_key, is_bonded, last_mixnet_status) 
VALUES ($1, $2, $3) 
ON CONFLICT (identity_key) DO UPDATE 
	SET identity_key = excluded.identity_key, 
		is_bonded = excluded.is_bonded, 
		last_mixnet_status = excluded.last_mixnet_status 
`
	_, err := db.Sql.Exec(stmt,
		mixnode.IdentityKey, mixnode.IsBonded, mixnode.LastMixnetStatus.String(),
	)
	if err != nil {
		return fmt.Errorf("error while saving Nym mixnode: %s", err)
	}

	return nil
}

// SaveNymMixnetV1Gateway allows to store the wasm params
func (db *Db) SaveNymMixnetV1Gateway(gateway types.GatewayV1) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v1_gateway(identity_key, is_bonded) 
VALUES ($1, $2) 
ON CONFLICT (identity_key) DO UPDATE 
	SET identity_key = excluded.identity_key, 
		is_bonded = excluded.is_bonded, 
`
	_, err := db.Sql.Exec(stmt,
		gateway.IdentityKey, gateway.IsBonded,
	)
	if err != nil {
		return fmt.Errorf("error while saving Nym gateway: %s", err)
	}

	return nil
}

func (db *Db) GetNymMixnetV1MixnodeEvent(eventKind string, identityKey string, sender *string, height *int64, executedAt *string) ([]dbtypes.NyxNymMixnetV1MixnodeEventsRow, error) {
	filter := fmt.Sprintf("WHERE event_kind = '%s' AND identity_key = '%s'", eventKind, identityKey)
	order := "height ASC, executed_at ASC"
	if sender != nil {
		filter = fmt.Sprintf("%s AND sender = '%s'", filter, *sender)
	}
	if height != nil {
		filter = fmt.Sprintf("%s AND height >= %d", filter, *height)
	} else if executedAt != nil {
		filter = fmt.Sprintf("%s AND executed_at >= '%s'", filter, *executedAt)
	}
	stmt := fmt.Sprintf(`SELECT * FROM nyx_nym_mixnet_v1_mixnode_events %s ORDER BY %s`, filter, order)

	var rows []dbtypes.NyxNymMixnetV1MixnodeEventsRow
	err := db.Sqlx.Select(&rows, stmt)
	return rows, err
}

// SaveNymMixnetV1MixnodeEvent allows to store the wasm contract events
func (db *Db) SaveNymMixnetV1MixnodeEvent(eventKind string, actor string, proxy *string, identityKey string, amount *cosmosTypes.Coins, dataType string, dataJson string, executeContract types.WasmExecuteContract, tx *juno.Tx) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v1_mixnode_events 
(event_kind, actor, sender, proxy, identity_key, amount, fee, contract_address, event_type, attributes, executed_at, height, hash) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	var dbAmount interface{}
	if amount != nil {
		dbAmount = pq.Array(dbtypes.NewDbCoins(*amount))
	}
	fee := pq.Array(dbtypes.NewDbCoins(tx.GetFee()))

	_, err := db.Sql.Exec(stmt,
		eventKind, actor,
		executeContract.Sender,
		proxy,
		identityKey,
		dbAmount, fee,
		executeContract.ContractAddress, dataType, dataJson,
		executeContract.ExecutedAt, executeContract.Height, tx.TxHash)
	if err != nil {
		return fmt.Errorf("error while saving wasm execute contracts: %s", err)
	}

	return nil
}

// HasNymMixnetV1MixnodeRewardingEvent checks if a rewarding event has been saved
func (db *Db) HasNymMixnetV1MixnodeRewardingEvent(identityKey string, tx *juno.Tx) (bool, error) {
	stmt := `
SELECT COUNT(height) FROM nyx_nym_mixnet_v1_mixnode_reward WHERE identity_key = $1 AND height = $2 AND hash = $3
`
	var count int
	err := db.Sql.QueryRow(stmt, identityKey, tx.Height, tx.TxHash).Scan(&count)

	return count > 0, err
}

func (db *Db) GetNymMixnetV1MixnodeRewardEvent(identityKey string, heightMin uint64, heightMax *uint64) ([]dbtypes.NyxNymMixnetV1MixnodeRewardRow, error) {
	stmt := fmt.Sprintf(`SELECT * FROM nyx_nym_mixnet_v1_mixnode_reward WHERE height >= %d AND identity_key = '%s'`, heightMin, identityKey)
	if heightMax != nil && *heightMax > 0 {
		stmt = fmt.Sprintf("%s AND height <= %d", stmt, *heightMax)
	}
	stmt = fmt.Sprintf("%s ORDER BY height ASC", stmt)
	var rows []dbtypes.NyxNymMixnetV1MixnodeRewardRow
	err := db.Sqlx.Select(&rows, stmt)
	log.Info().Int("count", len(rows)).Err(err).Msg(stmt)
	return rows, err
}

// SaveNymMixnetV1MixnodeRewardingEvent allows to store the mixnode rewarding events
func (db *Db) SaveNymMixnetV1MixnodeRewardingEvent(identityKey string, totalNodeReward cosmosTypes.Coins, totalDelegations cosmosTypes.Coins, operatorReward cosmosTypes.Coins, unitDelegatorReward decimal.Decimal, apy float64, stakingSupply cosmosTypes.Coins, profitMarginPercentage int, event cosmosTypes.StringEvent, executeContract types.WasmExecuteContract, tx *juno.Tx) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v1_mixnode_reward 
(sender, identity_key, total_node_reward, total_delegations, operator_reward, unit_delegator_reward, apy, staking_supply, profit_margin_percentage, contract_address, event_type, attributes, executed_at, height, hash) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT DO NOTHING`

	var attr = make(map[string]interface{}) // could be `map[string]string` however leaving to handle objects as values
	for _, entry := range event.Attributes {
		attr[entry.Key] = entry.Value
	}
	bytes, _ := json.Marshal(attr)

	dbTotalNodeReward := pq.Array(dbtypes.NewDbCoins(totalNodeReward))
	dbTotalDelegations := pq.Array(dbtypes.NewDbCoins(totalDelegations))
	dbOperatorReward := pq.Array(dbtypes.NewDbCoins(operatorReward))
	dbStakingSupply := pq.Array(dbtypes.NewDbCoins(stakingSupply))

	_, err := db.Sql.Exec(stmt,
		executeContract.Sender,
		identityKey,
		dbTotalNodeReward,
		dbTotalDelegations,
		dbOperatorReward,
		unitDelegatorReward.IntPart(),
		apy,
		dbStakingSupply,
		profitMarginPercentage,
		executeContract.ContractAddress, event.Type, string(bytes),
		executeContract.ExecutedAt, executeContract.Height, tx.TxHash)
	if err != nil {
		return fmt.Errorf("error while saving wasm execute contracts: %s", err)
	}

	return nil
}

// SaveNymMixnetV1MixnodeStatus allows to store when the mixnet rewarded set changes
func (db *Db) SaveNymMixnetV1MixnodeStatus(identityKey string, status string, routingScore int, executeContract types.WasmExecuteContract, tx *juno.Tx) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v1_mixnode_status 
(mixnet_status, routing_score, identity_key, executed_at, height, hash ) 
VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Sql.Exec(stmt,
		status, routingScore, identityKey,
		executeContract.ExecutedAt, executeContract.Height, tx.TxHash)
	if err != nil {
		return fmt.Errorf("error while saving wasm execute contracts: %s", err)
	}

	return nil
}
