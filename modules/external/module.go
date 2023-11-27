package external

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v5/modules"
	"github.com/forbole/juno/v5/types/config"
	"github.com/invopop/jsonschema"
	"github.com/rs/zerolog/log"
)

const (
	ModuleName = "external"
)

var (
	_ modules.Module            = &Module{}
	_ modules.TransactionModule = &Module{}
)

type Module struct {
	cfg      *Config
	cdc      codec.Codec
	cdcProto codec.ProtoCodecMarshaler
}

func NewModule(cfg config.Config, cdc codec.Codec, registry codectypes.InterfaceRegistry) *Module {
	bz, err := cfg.GetBytes()
	if err != nil {
		panic(err)
	}

	externalCfg, err := ParseConfig(bz)
	if err != nil {
		panic(err)
	}

	if externalCfg == nil {
		log.Info().Msg("config is nil")
		defaultConfig := DefaultConfig()
		externalCfg = &defaultConfig
	}

	log.Info().Str("url", externalCfg.URL).Msg("External API config")

	cdcProto := codec.NewProtoCodec(registry)

	// use reflection to write out the current JSON Schema of the Cosmos transaction response types
	schema := jsonschema.Reflect(&sdk.TxResponse{})
	// schema := jsonschema.ReflectFromType(sdk.TxResponse)
	json, _ := schema.MarshalJSON()
	os.WriteFile("schema.json", json, 0600)
	log.Info().Msg("Wrote `schema.json` with Cosmos transaction types")

	return &Module{
		cfg:      externalCfg,
		cdc:      cdc,
		cdcProto: cdcProto,
	}
}

func (m *Module) Name() string {
	return ModuleName
}
