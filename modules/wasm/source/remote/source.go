package remote

import (
	"encoding/json"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/forbole/juno/v3/node/remote"
	"github.com/rs/zerolog/log"

	wasmsource "github.com/forbole/bdjuno/v3/modules/wasm/source"
)

var (
	_ wasmsource.Source = &Source{}
)

// Source implements wasmsource.Source using a remote node
type Source struct {
	*remote.Source
	wasmClient wasmtypes.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, wasmClient wasmtypes.QueryClient) *Source {
	return &Source{
		Source:     source,
		wasmClient: wasmClient,
	}
}

// GetContractInfo implements wasmsource.Source
func (s Source) GetContractInfo(height int64, contractAddr string) (*wasmtypes.QueryContractInfoResponse, error) {
	res, err := s.wasmClient.ContractInfo(
		remote.GetHeightRequestContext(s.Ctx, height),
		&wasmtypes.QueryContractInfoRequest{
			Address: contractAddr,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while getting contract info: %s", err)
	}

	return res, nil
}

// GetContractStates implements wasmsource.Source
func (s Source) GetContractStates(height int64, contractAddr string) ([]wasmtypes.Model, error) {

	var models []wasmtypes.Model
	var nextKey []byte
	var stop = false
	for !stop {
		res, err := s.wasmClient.AllContractState(
			remote.GetHeightRequestContext(s.Ctx, height),
			&wasmtypes.QueryAllContractStateRequest{
				Address: contractAddr,
				Pagination: &query.PageRequest{
					Key:   nextKey,
					Limit: 100, // Query 100 states at time
				},
			},
		)
		if err != nil {
			return nil, fmt.Errorf("error while getting contract state: %s", err)
		}

		nextKey = res.Pagination.NextKey
		stop = len(res.Pagination.NextKey) == 0
		models = append(models, res.Models...)
	}

	return models, nil
}

func (s Source) GetNymMixnetContractV1MixnodeProfitMargin(height int64, contractAddress string, identityKey string, cdc codec.Codec) (*uint8, error) {
	body := []byte(fmt.Sprintf(`{
		"get_mixnode_at_height": {
			"mix_identity": "%s",
			"height": %v
		}
	}`, identityKey, height))

	request := wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddress,
		QueryData: body,
	}
	res, err := s.wasmClient.SmartContractState(s.Ctx, &request)

	if err != nil {
		log.Err(err).Str("identityKey", identityKey).Msg("Could not get profit margin")
		return nil, err
	}

	// marshal to JSON to avoid a protobuf definition of the MixnodeSnap (auto-generated below)
	jsonBytes, err := cdc.MarshalJSON(res)

	if err != nil {
		log.Err(err).Msg("Could marshal response to JSON")
		return nil, err
	}

	var mixnodeSnapshotPartial *MixnodeSnapshot = &MixnodeSnapshot{}
	err = json.Unmarshal(jsonBytes, mixnodeSnapshotPartial)
	if err != nil {
		log.Err(err).Msg("Could not marshal response to JSON")
		return nil, err
	}

	value := uint8(mixnodeSnapshotPartial.Data.MixNode.ProfitMarginPercent)
	return &value, nil
}

type MixnodeSnapshot struct {
	Data struct {
		PledgeAmount struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"pledge_amount"`
		Owner       string `json:"owner"`
		Layer       int    `json:"layer"`
		BlockHeight int    `json:"block_height"`
		MixNode     struct {
			Host                string `json:"host"`
			MixPort             int    `json:"mix_port"`
			VerlocPort          int    `json:"verloc_port"`
			HttpApiPort         int    `json:"http_api_port"`
			SphinxKey           string `json:"sphinx_key"`
			IdentityKey         string `json:"identity_key"`
			Version             string `json:"version"`
			ProfitMarginPercent int    `json:"profit_margin_percent"`
		} `json:"mix_node"`
		Proxy              interface{} `json:"proxy"`
		AccumulatedRewards string      `json:"accumulated_rewards"`
		EpochRewards       struct {
			Params struct {
				RewardBlockstamp int    `json:"reward_blockstamp"`
				Uptime           string `json:"uptime"`
				InActiveSet      bool   `json:"in_active_set"`
			} `json:"params"`
			Result struct {
				Reward string `json:"reward"`
				Lambda string `json:"lambda"`
				Sigma  string `json:"sigma"`
			} `json:"result"`
			EpochId int `json:"epoch_id"`
		} `json:"epoch_rewards"`
	} `json:"data"`
}
