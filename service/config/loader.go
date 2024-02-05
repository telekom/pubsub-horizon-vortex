// Copyright 2024 Deutsche Telekom IT GmbH
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strings"
)

var Current Configuration

func LoadConfiguration() {
	configureViper()
	setDefaults()
	readConfiguration()

	if err := viper.Unmarshal(&Current); err != nil {
		log.Fatal().Err(err).Msg("Could not unmarshal current configuration!")
	}

	Current.ApplyLogLevel()
}

func InitConfig() error {
	configureViper()
	setDefaults()
	return viper.SafeWriteConfig()
}

func configureViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("vortex")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func setDefaults() {
	viper.SetDefault("logLevel", "info")

	viper.SetDefault("metrics.enabled", false)
	viper.SetDefault("metrics.port", 8080)

	viper.SetDefault("kafka.brokers", "localhost:9092")
	viper.SetDefault("kafka.topics", []string{"status"})
	viper.SetDefault("kafka.groupName", "vortex")
	viper.SetDefault("kafka.sessionTimeoutSec", 40)

	viper.SetDefault("mongo.url", "mongodb://localhost:27017")
	viper.SetDefault("mongo.database", "horizon")
	viper.SetDefault("mongo.collection", "status")
	viper.SetDefault("mongo.bulkSize", 500)
	viper.SetDefault("mongo.flushIntervalSec", 30)
	viper.SetDefault("mongo.writeConcern.writes", 1)
	viper.SetDefault("mongo.writeConcern.journal", false)
}

func readConfiguration() {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Msg("Configuration file not found but environment variables will be taken into account!")
			viper.AutomaticEnv()
		} else {
			log.Fatal().Err(err).Msg("Could not read configuration!")
		}
	}
}
