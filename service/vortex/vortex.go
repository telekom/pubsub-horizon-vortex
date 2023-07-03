package vortex

import (
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"vortex/service/config"
	"vortex/service/kafka"
	"vortex/service/metrics"
	"vortex/service/mongo"
)

var (
	source       *kafka.Consumer
	sink         *mongo.Connection
	processGroup *sync.WaitGroup
)

func StartPipeline(config config.Configuration) {
	processGroup = new(sync.WaitGroup)
	processGroup.Add(2)

	var err error

	var sourceCfg = config.Kafka
	source, err = kafka.NewConsumer(&sourceCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while creating consumer!")
	}

	var sinkCfg = config.Mongo
	sink, err = mongo.NewConnection(&sinkCfg, source)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not establish database connection!")
	}

	if metrics.IsEnabled() {
		go metrics.ExposeMetrics()
	}

	go sink.Start(processGroup)
	go source.Start(processGroup)
	go terminateOnSignal()

	processGroup.Wait()
}

func Terminate() {
	source.Stop()
	sink.Stop()
	os.Exit(0)
}

func terminateOnSignal() {
	var sigintChannel = make(chan os.Signal, 1)
	signal.Notify(sigintChannel, syscall.SIGINT, syscall.SIGTERM)
	for {
		<-sigintChannel
		Terminate()
	}
}
