package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
	"vortex/service/config"
	"vortex/service/metrics"
	"vortex/service/utils"
)

type Consumer struct {
	consumer         sarama.ConsumerGroup
	config           *config.Kafka
	dataChannel      chan *sarama.ConsumerMessage
	commitChannel    chan bool
	reBalanceChannel chan bool
	consumerCtx      context.Context
	consumerCancel   context.CancelFunc
}

func NewConsumer(config *config.Kafka) (*Consumer, error) {
	var consumerConfig = sarama.NewConfig()

	consumerConfig.Consumer.Group.Session.Timeout = time.Duration(config.SessionTimeoutSec) * time.Second
	consumerConfig.Net.ReadTimeout = consumerConfig.Consumer.Group.Session.Timeout + (5 * time.Second)
	consumerConfig.Consumer.Offsets.AutoCommit.Enable = false

	var client, err = sarama.NewConsumerGroup(config.Brokers, config.GroupName, consumerConfig)
	if err != nil {
		return nil, err
	}

	var ctx, cancel = context.WithCancel(context.Background())
	return &Consumer{
		consumer:       client,
		config:         config,
		dataChannel:    make(chan *sarama.ConsumerMessage),
		commitChannel:  make(chan bool),
		consumerCtx:    ctx,
		consumerCancel: cancel,
	}, nil
}

func (c *Consumer) Start(processGroup *sync.WaitGroup) {
	defer processGroup.Done()
	for {
		if err := c.consumer.Consume(c.consumerCtx, c.config.Topics, c); err != nil {
			log.Fatal().Err(err).Msg("A consumer error has occurred")
		}

		if c.consumerCtx.Err() != nil {
			return
		}
	}
}

func (c *Consumer) Stop() {
	c.consumerCancel()
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	var fields = utils.GetFieldsFromClaims(session.Claims())
	log.Info().Fields(fields).Msg("Received assignment from Kafka")
	return c.seekToLastCommittedOffset(session)
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	select {

	case <-session.Context().Done():
		if err := session.Context().Err(); err != nil {
			log.Err(err).Msg("Session has ended unexpectedly")
		}

	default:
		log.Info().Msg("Re-balance is about to happen. Committing offsets...")
		session.Commit()
		c.reBalanceChannel <- true

	}
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {

		case message := <-claim.Messages():
			if message != nil {
				c.dataChannel <- message
				log.Debug().Fields(utils.GetFieldsFromMessage(message)).Msg("Consumed message")
				session.MarkMessage(message, "")
				metrics.RecordConsumption(message)
			}

		case doCommit := <-c.commitChannel:
			if doCommit {
				session.Commit()
				log.Debug().Msg("Committed offsets")
			}

		case <-session.Context().Done():
			log.Debug().Msg("Session has ended. Awaiting re-balance...")
			return nil

		}
	}
}

func (c *Consumer) GetOutput() <-chan *sarama.ConsumerMessage {
	return c.dataChannel
}

func (c *Consumer) CommitOffsets() {
	go func() {
		c.commitChannel <- true
	}()
}

func (c *Consumer) AwaitReBalance() {
	<-c.reBalanceChannel
}

func (c *Consumer) seekToLastCommittedOffset(session sarama.ConsumerGroupSession) error {
	var consumerGroup = config.Current.Kafka.GroupName
	var claims = session.Claims()
	var fields = make(map[string]any)

	log.Debug().Msgf("Requesting coordinator of consumer group %s", consumerGroup)
	client, err := sarama.NewClient(config.Current.Kafka.Brokers, sarama.NewConfig())
	if err != nil {
		return err
	}

	coordinator, err := client.Coordinator(consumerGroup)
	if err != nil {
		return err
	}

	log.Info().Msgf("Restoring offsets from coordinator %d", coordinator.ID())
	request := new(sarama.OffsetFetchRequest)
	request.Version = 1
	request.ConsumerGroup = consumerGroup

	for topic, partitions := range claims {
		for _, partition := range partitions {
			request.AddPartition(topic, partition)
		}
	}

	blocks, err := coordinator.FetchOffset(request)
	if err != nil {
		return err
	}

	for topic, partitions := range claims {
		for _, partition := range partitions {
			block := blocks.GetBlock(topic, partition)
			fields[fmt.Sprintf("%s.%d", topic, partition)] = block.Offset
			session.ResetOffset(topic, partition, block.Offset, "")
		}
	}

	log.Info().Fields(fields).Msg("Restored last seen offsets from coordinator")
	return nil
}
