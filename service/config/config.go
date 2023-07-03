package config

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

type Configuration struct {
	LogLevel string  `mapstructure:"logLevel"`
	Metrics  Metrics `mapstructure:"metrics"`
	Kafka    Kafka   `mapstructure:"kafka"`
	Mongo    Mongo   `mapstructure:"mongo"`
}

type Kafka struct {
	Brokers           []string `mapstructure:"brokers"`
	Topics            []string `mapstructure:"topics"`
	GroupName         string   `mapstructure:"groupName"`
	SessionTimeoutSec int      `mapstructure:"sessionTimeoutSec"`
}

type Mongo struct {
	Url              string            `mapstructure:"url"`
	Database         string            `mapstructure:"database"`
	Collection       string            `mapstructure:"collection"`
	BulkSize         int               `mapstructure:"bulkSize"`
	FlushIntervalSec int               `mapstructure:"flushIntervalSec"`
	WriteConcern     MongoWriteConcern `mapstructure:"writeConcern"`
}

type MongoWriteConcern struct {
	Writes  int  `mapstructure:"writes"`
	Journal bool `mapstructure:"journal"`
}

type Metrics struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

func (c *Configuration) ApplyLogLevel() {
	logLevel, err := zerolog.ParseLevel(c.LogLevel)
	if err != nil {
		panic(errors.New(fmt.Sprintf("unknown log level '%s'", c.LogLevel)))
	}

	if logLevel == zerolog.DebugLevel {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	log.Logger = log.Level(logLevel)
}
