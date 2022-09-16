package source

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

type Source interface {
	GetContractInfo(height int64, contractAddr string) (*wasmtypes.QueryContractInfoResponse, error)
	GetContractStates(height int64, contractAddress string) ([]wasmtypes.Model, error)
	GetNymMixnetContractV1MixnodeProfitMargin(height int64, contractAddress string, identityKey string, cdc codec.Codec) (*uint8, error)
}
