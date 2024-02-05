// Copyright 2024 Deutsche Telekom IT GmbH
//
// SPDX-License-Identifier: Apache-2.0

package vortex_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"testing"
	"time"
	"vortex/service/config"
	"vortex/service/vortex"
)

var (
	kafkaPort   string
	kafkaHost   string
	mongoPort   string
	mongoHost   string
	testMessage = createWorkingCopy(mustReadJson("../../testdata/kafka_msg.json"))
)

func TestMain(m *testing.M) {
	var resources = make([]*dockertest.Resource, 0)

	log.Println("Performing integration test...")
	log.Println("Starting required containers via docker...")

	// Docker pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		log.Fatalf("Could not ping docker: %s", err)
	}

	// Test network
	log.Println("Creating docker network...")
	network, err := pool.CreateNetwork("vortex")
	if err != nil {
		log.Fatalf("Could not create network: %s", err)
	}

	// MongoDB
	log.Println("Creating mongodb...")
	mongodbContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Hostname:     "mongo",
		Name:         "vortex-mongo",
		Repository:   envOrDefault("MONGO_IMAGE", "mongo"),
		Tag:          envOrDefault("MONGO_TAG", "7.0.5-rc0"),
		ExposedPorts: []string{"27017/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"27017/tcp": {{HostIP: "localhost", HostPort: "27017"}},
		},
	}, teardownContainer)
	_ = mongodbContainer.Expire(120)

	if err != nil {
		log.Fatalf("Could not start mongodb: %s", err)
	}
	_ = mongodbContainer.ConnectToNetwork(network)

	// Kafka
	log.Println("Creating kafka...")

	kafkaContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Hostname:   "kafka",
		Name:       "vortex-kafka",
		Repository: envOrDefault("KAFKA_IMAGE", "bitnami/kafka"),
		Tag:        envOrDefault("KAFKA_TAG", "3.4.1"),
		Env: []string{
			"KAFKA_ENABLE_KRAFT=yes",
			"KAFKA_CFG_PROCESS_ROLES=broker,controller",
			"KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER",
			"KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093,EXTERNAL://:9094",
			"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,EXTERNAL:PLAINTEXT",
			"KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://vortex-kafka:9092,EXTERNAL://" + envOrDefault("KAFKA_HOST", "localhost:9094"),
			"KAFKA_BROKER_ID=1",
			"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=1@localhost:9093",
			"ALLOW_PLAINTEXT_LISTENER=yes",
			"KAFKA_CFG_NODE_ID=1",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE=true",
			"BITNAMI_DEBUG=yes",
			"KAFKA_CFG_NUM_PARTITIONS=2",
		},
		ExposedPorts: []string{"9094/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9094/tcp": {{HostIP: "localhost", HostPort: "9094"}},
		},
	}, teardownContainer)
	_ = kafkaContainer.Expire(120)

	if err != nil {
		log.Fatalf("Could not start kafka: %s", err)
	}
	_ = kafkaContainer.ConnectToNetwork(network)

	// Add resources to clean up after testing
	resources = append(resources, mongodbContainer, kafkaContainer)

	// Container ports
	mongoPort = mongodbContainer.GetPort("27017/tcp")
	mongoHost = envOrDefault("MONGO_HOST", "localhost:27017")

	kafkaPort = kafkaContainer.GetPort("9094/tcp")
	kafkaHost = envOrDefault("KAFKA_HOST", "localhost:9094")

	// Retry for readiness
	if err := pool.Retry(func() error {

		// Ping mongodb
		mongoClient, err := mongo.Connect(context.TODO(), options.Client().
			ApplyURI(
				"mongodb://"+mongoHost,
			),
		)
		if err != nil {
			return err
		}

		if err := mongoClient.Ping(context.Background(), nil); err != nil {
			return err
		}

		// Wait for Kafka to be available
		brokers := []string{kafkaHost}
		kafkaClient, err := sarama.NewClient(brokers, sarama.NewConfig())
		if err != nil {
			return err
		}

		if len(kafkaClient.Brokers()) != 1 {
			return errors.New("expected 1 broker to be available")
		}

		return err
	}); err != nil {
		log.Fatalf("Could not reach pods for testing: %s", err)
	}

	code := m.Run()

	log.Println("Cleaning up...")
	for _, resource := range resources {
		log.Printf("Purging container '%s'...\n", resource.Container.Name)
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge container '%s': %s", resource.Container.Name, err)
		}
	}
	_ = pool.RemoveNetwork(network)
	os.Exit(code)
}

func TestStartPipeline(t *testing.T) {
	var assertions = assert.New(t)

	config.Current = config.Configuration{
		LogLevel: "debug",

		Kafka: config.Kafka{
			Brokers:           []string{kafkaHost},
			GroupName:         "vortex",
			Topics:            []string{"vortex"},
			SessionTimeoutSec: 40,
		},

		Mongo: config.Mongo{
			Url:              "mongodb://" + mongoHost,
			BulkSize:         1,
			Collection:       "vortex",
			Database:         "vortex",
			FlushIntervalSec: 5,
			WriteConcern: config.MongoWriteConcern{
				Writes:  1,
				Journal: false,
			},
		},

		Metrics: config.Metrics{
			Enabled: false,
			Port:    0,
		},
	}

	go vortex.StartPipeline(config.Current)
	log.Println("Waiting 30s for kafka to perform re-balance...")
	time.Sleep(30 * time.Second)

	// Produce sample event
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{envOrDefault("KAFKA_HOST", "localhost:"+kafkaPort)}, producerConfig)
	assertions.Nil(err, "expected no error when creating producer")

	jsonEvent, err := json.Marshal(testMessage)
	assertions.Nil(err, "expected no error when marshalling event")

	var dummyMessage = new(sarama.ProducerMessage)
	dummyMessage.Key = sarama.StringEncoder(testMessage["uuid"].(string))
	dummyMessage.Headers = []sarama.RecordHeader{
		{Key: []byte("type"), Value: []byte("MESSAGE")},
	}
	dummyMessage.Topic = "vortex"
	dummyMessage.Value = sarama.StringEncoder(jsonEvent)
	dummyMessage.Timestamp = time.Now()

	_, _, err = producer.SendMessage(dummyMessage)
	assertions.Nil(err, "expected no error when sending message")
	log.Println("Sent message to kafka")

	log.Println("Waiting 10s for message to be processed and flushed to database...")
	time.Sleep(10 * time.Second)

	// Check if message is in database
	mongoUrl, mongoDatabase, mongoCollections := config.Current.Mongo.Url, config.Current.Mongo.Database, config.Current.Mongo.Collection
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUrl))
	assertions.Nil(err)

	collection := mongoClient.Database(mongoDatabase).Collection(mongoCollections)
	result := collection.FindOne(context.TODO(), bson.D{{"_id", testMessage["uuid"].(string)}}, options.FindOne())
	assertions.Nil(result.Err(), "expected no error when querying database")

	var entry map[string]any
	assertions.Nil(result.Decode(&entry), "expected no error when decoding result")
	assertions.Equal(testMessage["uuid"].(string), entry["_id"])
	log.Println("Found test-message in database")
}

func envOrDefault(env string, fallback string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	}

	return fallback
}

func teardownContainer(config *docker.HostConfig) {
	config.AutoRemove = true
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

func createWorkingCopy(data map[string]any) map[string]any {
	var copy = make(map[string]any)
	for k, v := range data {
		copy[k] = v
	}

	return copy
}

func mustReadJson(filename string) map[string]any {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var data = make(map[string]any)
	if err := json.Unmarshal(bytes, &data); err != nil {
		panic(err)
	}

	return data
}
