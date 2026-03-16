// Copyright 2024 Deutsche Telekom IT GmbH
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"vortex/service/config"
	"vortex/service/utils"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const Namespace = "vortex"

var (
	messagesConsumedTotal prometheus.Counter
	metadataConsumedTotal prometheus.Counter

	upsertedTotal prometheus.Counter

	registry *prometheus.Registry

	enabled *bool
)

func init() {
	registry = prometheus.NewRegistry()

	messagesConsumedTotal = createCounter("messages_consumed_total", "The total amount of consumed messages")
	metadataConsumedTotal = createCounter("metadata_consumed_total", "The total amount of consumed metadata")
	registry.MustRegister(messagesConsumedTotal, metadataConsumedTotal)

	upsertedTotal = createCounter("upserted_total", "The total amount of upserted datasets")
	registry.MustRegister(upsertedTotal)
}

func RecordConsumption(message *sarama.ConsumerMessage) {
	if !isEnabled() {
		return
	}

	var messageType = utils.GetHeader(message.Headers, "type")
	switch strings.ToLower(messageType) {

	case "message":
		messagesConsumedTotal.Inc()

	case "metadata":
		metadataConsumedTotal.Inc()

	default:
		var fields = map[string]any{
			"type": messageType,
			"key":  string(message.Key),
		}
		log.Warn().Fields(fields).Msg("Unknown message type. Will not be recorded in metrics.")

	}
}

func RecordUpserts(datasetCount float64) {
	if !isEnabled() {
		return
	}
	upsertedTotal.Add(float64(datasetCount))
}

func ExposeMetrics() {
	http.HandleFunc("/livez", healthHandler("livez"))
	http.HandleFunc("/readyz", healthHandler("readyz"))

	var metricsEnabled = isEnabled()
	if metricsEnabled {
		http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			Timeout: 15 * time.Second,
		}))
	}

	go func(port int) {
		var addr = fmt.Sprintf(":%d", port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Panic().Err(err).Msg("Could not start health/metrics server!")
		}
	}(config.Current.Metrics.Port)

	if metricsEnabled {
		log.Info().Msgf("Serving /metrics, /livez and /readyz on port %d", config.Current.Metrics.Port)
	} else {
		log.Info().Msgf("Serving /livez and /readyz on port %d (metrics disabled)", config.Current.Metrics.Port)
	}
}

func createCounter(name string, help string) prometheus.Counter {
	return promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      name,
		Help:      help,
	})
}

func isEnabled() bool {
	if enabled == nil {
		enabled = &config.Current.Metrics.Enabled
	}
	return *enabled
}

func healthHandler(endpoint string) http.HandlerFunc {
	var body = []byte("ok")
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write(body); err != nil {
			log.Error().Err(err).Msgf("Could not write %s response", endpoint)
		}
	}
}
