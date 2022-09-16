package database

import (
	"encoding/json"
	"fmt"
	cosmosTypes "github.com/cosmos/cosmos-sdk/types"
	dbtypes "github.com/forbole/bdjuno/v3/database/types"
	"github.com/forbole/bdjuno/v3/types"
	juno "github.com/forbole/juno/v3/types"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// EnsureExistsNymMixnetV2Mixnode ensures a mixnode is in the store
func (db *Db) EnsureExistsNymMixnetV2Mixnode(mixnode types.MixnodeV2) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v2_mixnode(mix_id, identity_key, is_bonded, last_mixnet_status) 
VALUES ($1, $2, $3, $4) 
ON CONFLICT DO NOTHING
`
	_, err := db.Sql.Exec(stmt,
		mixnode.MixId, mixnode.IdentityKey, mixnode.IsBonded, mixnode.LastMixnetStatus.String(),
	)
	if err != nil {
		return fmt.Errorf("error while ensuring Nym mixnode exists: %s", err)
	}

	return nil
}

// SaveNymMixnetV2Mixnode allows to create or update a mixnode
func (db *Db) SaveNymMixnetV2Mixnode(mixnode types.MixnodeV2) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v2_mixnode(mix_id, identity_key, is_bonded, last_mixnet_status) 
VALUES ($1, $2, $3, $4) 
ON CONFLICT (identity_key) DO UPDATE 
	SET mix_id = excluded.mix_id,
	    identity_key = excluded.identity_key, 
		is_bonded = excluded.is_bonded, 
		last_mixnet_status = excluded.last_mixnet_status 
`
	_, err := db.Sql.Exec(stmt,
		mixnode.MixId, mixnode.IdentityKey, mixnode.IsBonded, mixnode.LastMixnetStatus.String(),
	)
	if err != nil {
		return fmt.Errorf("error while saving Nym mixnode: %s", err)
	}

	return nil
}

// HasNymMixnetV2MixnodeRewardingEvent checks if a rewarding event has been saved
func (db *Db) HasNymMixnetV2MixnodeRewardingEvent(mixId uint32, tx *juno.Tx) (bool, error) {
	stmt := `
SELECT COUNT(height) FROM nyx_nym_mixnet_v2_mixnode_reward WHERE mix_id = $1 AND height = $2 AND hash = $3
`
	var count int
	err := db.Sql.QueryRow(stmt, mixId, tx.Height, tx.TxHash).Scan(&count)

	return count > 0, err
}

// SaveNymMixnetV2MixnodeRewardingEvent allows to store the mixnode rewarding events
func (db *Db) SaveNymMixnetV2MixnodeRewardingEvent(mixId uint32, operatorReward cosmosTypes.Coins, delegatesReward cosmosTypes.Coins, priorDelegates cosmosTypes.Coins, priorUnitDelegation decimal.Decimal, apy float64, event cosmosTypes.StringEvent, executeContract types.WasmExecuteContract, tx *juno.Tx) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v2_mixnode_reward 
(sender, mix_id, operator_reward, delegates_reward, prior_delegates, prior_unit_delegation, apy, contract_address, event_type, attributes, executed_at, height, hash) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT DO NOTHING`

	var attr = make(map[string]interface{}) // could be `map[string]string` however leaving to handle objects as values
	for _, entry := range event.Attributes {
		attr[entry.Key] = entry.Value
	}
	bytes, _ := json.Marshal(attr)

	dbOperatorReward := pq.Array(dbtypes.NewDbCoins(operatorReward))
	dbDelegatesReward := pq.Array(dbtypes.NewDbCoins(delegatesReward))
	dbPriorDelegates := pq.Array(dbtypes.NewDbCoins(priorDelegates))

	_, err := db.Sql.Exec(stmt,
		executeContract.Sender,
		mixId,
		dbOperatorReward,
		dbDelegatesReward,
		dbPriorDelegates,
		priorUnitDelegation.IntPart(),
		apy,
		executeContract.ContractAddress, event.Type, string(bytes),
		executeContract.ExecutedAt, executeContract.Height, tx.TxHash)
	if err != nil {
		return fmt.Errorf("error while saving wasm execute contracts: %s", err)
	}

	return nil
}

// SaveNymMixnetV2MixnodeStatus allows to store when the mixnet rewarded set changes
func (db *Db) SaveNymMixnetV2MixnodeStatus(mixId uint32, status string, routingScore int64, executeContract types.WasmExecuteContract, tx *juno.Tx) error {
	stmt := `
INSERT INTO nyx_nym_mixnet_v2_mixnode_status 
(mixnet_status, routing_score, mix_id, executed_at, height, hash ) 
VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Sql.Exec(stmt,
		status, routingScore, mixId,
		executeContract.ExecutedAt, executeContract.Height, tx.TxHash)
	if err != nil {
		return fmt.Errorf("error while saving wasm execute contracts: %s", err)
	}

	return nil
}
