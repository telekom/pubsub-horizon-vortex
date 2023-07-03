package metrics

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
	"vortex/service/config"
	"vortex/service/utils"
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
	if !IsEnabled() {
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
	if !IsEnabled() {
		return
	}
	upsertedTotal.Add(float64(datasetCount))
}

func ExposeMetrics() {
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Timeout: 15 * time.Second,
	}))

	go func() {
		var addr = fmt.Sprintf(":%d", config.Current.Metrics.Port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Panic().Err(err).Msg("Could not start serving metrics!")
		}
	}()

	log.Info().Msgf("Serving metrics on port %d", config.Current.Metrics.Port)
}

func createCounter(name string, help string) prometheus.Counter {
	return promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      name,
		Help:      help,
	})
}

func IsEnabled() bool {
	if enabled == nil {
		enabled = &config.Current.Metrics.Enabled
	}
	return *enabled
}
