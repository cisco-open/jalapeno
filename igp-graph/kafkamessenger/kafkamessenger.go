package kafkamessenger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// KafkaMessenger handles Kafka message consumption for the IGP graph processor
type KafkaMessenger struct {
	consumer sarama.ConsumerGroup
	dbSrv    dbclient.DB
	stop     chan struct{}
	topics   []string
	brokers  []string
	groupID  string
}

// NewKafkaMessenger creates a new Kafka messenger for IGP graph processing
func NewKafkaMessenger(kafkaConn string, dbSrv dbclient.DB) (*KafkaMessenger, error) {
	if err := validateConnection(kafkaConn); err != nil {
		return nil, err
	}

	brokers := strings.Split(kafkaConn, ",")

	// Topics that the IGP graph processor subscribes to
	topics := []string{
		"gobmp.parsed.ls_node",
		"gobmp.parsed.ls_link",
		"gobmp.parsed.ls_prefix",
		"gobmp.parsed.ls_srv6_sid",
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.MaxProcessingTime = 1 * time.Second
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Retry.Max = 3

	groupID := "igp-graph-processor"
	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaMessenger{
		consumer: consumer,
		dbSrv:    dbSrv,
		stop:     make(chan struct{}),
		topics:   topics,
		brokers:  brokers,
		groupID:  groupID,
	}, nil
}

// Start begins consuming Kafka messages
func (k *KafkaMessenger) Start() {
	glog.Infof("Starting Kafka messenger for IGP graph processor, group: %s, topics: %v",
		k.groupID, k.topics)

	go func() {
		defer k.consumer.Close()

		for {
			select {
			case <-k.stop:
				glog.Info("Kafka messenger stopping...")
				return
			default:
				ctx := context.Background()
				handler := &MessageHandler{dbSrv: k.dbSrv}

				if err := k.consumer.Consume(ctx, k.topics, handler); err != nil {
					glog.Errorf("Error consuming from Kafka: %v", err)
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()

	// Error handling goroutine
	go func() {
		for err := range k.consumer.Errors() {
			glog.Errorf("Kafka consumer error: %v", err)
		}
	}()
}

// Stop stops the Kafka messenger
func (k *KafkaMessenger) Stop() {
	glog.Info("Stopping Kafka messenger...")
	close(k.stop)
	time.Sleep(100 * time.Millisecond) // Give time for graceful shutdown
}

// MessageHandler implements sarama.ConsumerGroupHandler
type MessageHandler struct {
	dbSrv dbclient.DB
}

// Setup is called when a consumer group session starts
func (h *MessageHandler) Setup(sarama.ConsumerGroupSession) error {
	glog.V(5).Info("IGP graph consumer group session setup")
	return nil
}

// Cleanup is called when a consumer group session ends
func (h *MessageHandler) Cleanup(sarama.ConsumerGroupSession) error {
	glog.V(5).Info("IGP graph consumer group session cleanup")
	return nil
}

// ConsumeClaim processes messages from a partition
func (h *MessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			if err := h.processMessage(message); err != nil {
				glog.Errorf("Failed to process message from topic %s, partition %d, offset %d: %v",
					message.Topic, message.Partition, message.Offset, err)
				// Continue processing other messages even if one fails
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *MessageHandler) processMessage(message *sarama.ConsumerMessage) error {
	glog.V(9).Infof("Processing message from topic: %s, partition: %d, offset: %d",
		message.Topic, message.Partition, message.Offset)

	// Determine message type based on topic
	var msgType dbclient.CollectionType
	switch message.Topic {
	case "gobmp.parsed.ls_node":
		msgType = bmp.LSNodeMsg
	case "gobmp.parsed.ls_link":
		msgType = bmp.LSLinkMsg
	case "gobmp.parsed.ls_prefix":
		msgType = bmp.LSPrefixMsg
	case "gobmp.parsed.ls_srv6_sid":
		msgType = bmp.LSSRv6SIDMsg
	default:
		glog.V(5).Infof("Ignoring message from unsupported topic: %s", message.Topic)
		return nil
	}

	// Validate that the message is valid JSON
	var temp interface{}
	if err := json.Unmarshal(message.Value, &temp); err != nil {
		return err
	}

	// Store the message for processing
	return h.dbSrv.StoreMessage(msgType, message.Value)
}

func validateConnection(kafkaConn string) error {
	if kafkaConn == "" {
		return fmt.Errorf("kafka connection string cannot be empty")
	}
	return nil
}
