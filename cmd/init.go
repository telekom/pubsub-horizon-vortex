// Copyright 2024 Deutsche Telekom IT GmbH
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"vortex/service/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new configuration file for local testing",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.InitConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileAlreadyExistsError); ok {
				log.Error().Msg("Configuration already exists!")
			} else {
				log.Fatal().Err(err).Msg("Could not initialize configuration!")
			}
		}
		log.Info().Msg("Configuration initialized!")
	},
}
