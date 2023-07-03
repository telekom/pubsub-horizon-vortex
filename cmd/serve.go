package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"vortex/service/config"
	"vortex/service/vortex"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the actual service",
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfiguration()
		log.Info().Msg("Starting vortex...")
		vortex.StartPipeline(config.Current)
	},
}
