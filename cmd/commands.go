package cmd

import "github.com/rs/zerolog/log"

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("Could not execute root command!")
	}
}
