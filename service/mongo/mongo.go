package mongo

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"sync"
	"time"
	"vortex/service/config"
	"vortex/service/kafka"
	"vortex/service/metrics"
	"vortex/service/transforms"
	"vortex/service/utils"
)

type Connection struct {
	client            *mongo.Client
	config            *config.Mongo
	connectionContext context.Context
	connectionCancel  context.CancelFunc
	source            *kafka.Consumer
	updateOptions     *options.UpdateOptions
	bulk              []mongo.WriteModel
	mutex             sync.Mutex
}

func NewConnection(config *config.Mongo, source *kafka.Consumer) (*Connection, error) {
	var clientOpts = options.Client()
	var ctx, cancel = context.WithCancel(context.Background())

	var writeConcern = &writeconcern.WriteConcern{
		W:       config.WriteConcern.Writes,
		Journal: &config.WriteConcern.Journal,
	}

	clientOpts.ApplyURI(config.Url)
	clientOpts.SetWriteConcern(writeConcern)

	var client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		cancel()
		return nil, err
	}

	var updateOptions = options.Update()
	var enableUpsert = true
	updateOptions.Upsert = &enableUpsert

	return &Connection{
		client:            client,
		config:            config,
		source:            source,
		connectionContext: ctx,
		connectionCancel:  cancel,
		updateOptions:     updateOptions,
		bulk:              make([]mongo.WriteModel, 0),
	}, nil
}

func (c *Connection) Start(processGroup *sync.WaitGroup) {
	if err := c.client.Ping(context.TODO(), nil); err != nil {
		log.Fatal().Err(err).Msg("Could not connect to database")
	}
	log.Info().Msg("Database connection established")
	go c.flushWithInterval(time.Duration(c.config.FlushIntervalSec) * time.Second)

	defer processGroup.Done()
	for {
		select {

		case message := <-c.source.GetOutput():
			if err := c.upsert(message); err != nil {
				var fields = utils.GetFieldsFromMessage(message)
				log.Fatal().Fields(fields).Err(err).Msg("Could not perform update in database")
			}

		case <-c.connectionContext.Done():
			c.flush()
			return

		default:
			if c.connectionContext.Err() != nil {
				if err := c.client.Disconnect(c.connectionContext); err != nil {
					log.Fatal().Err(err).Msg("Could no disconnect from database")
				}
			}

		}
	}
}

func (c *Connection) Stop() {
	c.connectionCancel()
}

func (c *Connection) upsert(message *sarama.ConsumerMessage) error {
	var document map[string]any
	var filter = bson.M{"_id": string(message.Key)}

	if message.Value == nil {
		return nil
	}

	if err := json.Unmarshal(message.Value, &document); err != nil {
		return err
	}
	delete(document, "_id")

	if castedEvent, ok := document["event"].(map[string]any); ok {
		if castedId, ok := castedEvent["id"]; ok {
			filter["event.id"] = castedId
		} else {
			log.Warn().Fields(map[string]any{
				"partition": message.Partition,
				"offset":    message.Offset,
			}).Msg("Detected faulty message. Skipping!")
			return nil
		}
	} else {
		log.Warn().Fields(map[string]any{
			"partition": message.Partition,
			"offset":    message.Offset,
		}).Msg("Detected faulty message. Skipping!")
		return nil
	}

	document["topic"] = message.Topic
	var transformedDoc, err = transforms.GlobalRegistry.ApplyTransforms(document)
	if err != nil {
		log.Fatal().Fields(utils.GetFieldsFromMessage(message)).Err(err).Msg("Could not apply transformations to document")
	}

	var messageType = utils.GetHeader(message.Headers, "type")
	if messageType == "MESSAGE" {
		transformedDoc["coordinates"] = map[string]any{"partition": message.Partition, "offset": message.Offset}
		transformedDoc["timestamp"] = message.Timestamp
	}

	var update = bson.M{"$set": transformedDoc}

	c.mutex.Lock()
	c.bulk = append(c.bulk, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	c.mutex.Unlock()

	if len(c.bulk) >= c.config.BulkSize {
		c.flush()
	}

	return nil
}

func (c *Connection) flush() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.bulk) == 0 {
		return
	}

	var opts = options.BulkWrite().SetOrdered(false)
	var database = c.client.Database(c.config.Database)
	var collection = database.Collection(c.config.Collection)

	result, err := collection.BulkWrite(c.connectionContext, c.bulk, opts)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not perform bulk-write")
	}

	var fields = map[string]any{
		"upserted": result.UpsertedCount,
		"inserted": result.InsertedCount,
		"modified": result.ModifiedCount,
	}
	log.Debug().Fields(fields).Msgf("Completed bulk-write")
	metrics.RecordUpserts(float64(len(c.bulk)))

	c.bulk = make([]mongo.WriteModel, 0)
	c.source.CommitOffsets()
}

func (c *Connection) flushWithInterval(interval time.Duration) {
	for {
		time.Sleep(interval)

		if c.connectionContext.Err() != nil {
			return
		}

		if len(c.bulk) > 0 {
			c.flush()
		}
	}
}
