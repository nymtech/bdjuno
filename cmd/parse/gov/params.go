package gov

import (
	"github.com/spf13/cobra"

	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/callisto/v4/modules/distribution"
	"github.com/forbole/callisto/v4/modules/gov"
	"github.com/forbole/callisto/v4/modules/mint"
	"github.com/forbole/callisto/v4/modules/slashing"
	"github.com/forbole/callisto/v4/modules/staking"
	modulestypes "github.com/forbole/callisto/v4/modules/types"
	parsecmdtypes "github.com/forbole/juno/v5/cmd/parse/types"
	"github.com/forbole/juno/v5/types/config"
)

func paramsCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Get the current parameters of the gov module",
		RunE: func(cmd *cobra.Command, args []string) error {
			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			sources, err := modulestypes.BuildSources(config.Cfg.Node, parseCtx.EncodingConfig)
			if err != nil {
				return err
			}

			// Get the database
			db := database.Cast(parseCtx.Database)

			// Build expected modules of gov modules
			distrModule := distribution.NewModule(sources.DistrSource, parseCtx.EncodingConfig.Codec, db)
			mintModule := mint.NewModule(sources.MintSource, parseCtx.EncodingConfig.Codec, db)
			slashingModule := slashing.NewModule(sources.SlashingSource, parseCtx.EncodingConfig.Codec, db)
			stakingModule := staking.NewModule(sources.StakingSource, parseCtx.EncodingConfig.Codec, db)

			// Build the gov module
			govModule := gov.NewModule(sources.GovSource, distrModule, mintModule, slashingModule, stakingModule, parseCtx.EncodingConfig.Codec, db)

			height, err := parseCtx.Node.LatestHeight()
			if err != nil {
				return err
			}

			return govModule.UpdateParams(height)
		},
	}
}
