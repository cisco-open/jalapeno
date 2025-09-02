package kafkamessenger

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// Srv defines required method of a processor server
type Srv interface {
	Start() error
	Stop() error
}

// MessageHandler handles Kafka messages
type MessageHandler struct {
	dbSrv dbclient.Srv
}

// Setup implements sarama.ConsumerGroupHandler
func (h *MessageHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler
func (h *MessageHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler
func (h *MessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := h.processMessage(message); err != nil {
			glog.Errorf("Failed to process message from topic %s: %v", message.Topic, err)
			continue
		}
		session.MarkMessage(message, "")
	}
	return nil
}

func (h *MessageHandler) processMessage(message *sarama.ConsumerMessage) error {
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
	case "gobmp.parsed.peer":
		msgType = bmp.PeerStateChangeMsg
	case "gobmp.parsed.unicast_prefix_v4":
		msgType = bmp.UnicastPrefixV4Msg
	case "gobmp.parsed.unicast_prefix_v6":
		msgType = bmp.UnicastPrefixV6Msg
	default:
		glog.V(5).Infof("Ignoring message from unsupported topic: %s", message.Topic)
		return nil
	}

	// Parse the raw message data
	var bmpData map[string]interface{}
	if err := json.Unmarshal(message.Value, &bmpData); err != nil {
		return err
	}

	// Add Kafka metadata
	bmpData["_kafka_topic"] = message.Topic
	bmpData["_kafka_partition"] = message.Partition
	bmpData["_kafka_offset"] = message.Offset
	bmpData["_kafka_timestamp"] = message.Timestamp

	// Marshal back to JSON for processing
	processedData, err := json.Marshal(bmpData)
	if err != nil {
		return err
	}

	// Send to database processor
	return h.dbSrv.GetInterface().StoreMessage(msgType, processedData)
}

type kafka struct {
	stopCh   chan struct{}
	brokers  []string
	dbSrv    dbclient.Srv
	config   *sarama.Config
	consumer sarama.ConsumerGroup
	topics   []string
}

// NewKafkaMessenger returns an instance of a kafka consumer acting as a messenger server
func NewKafkaMessenger(kafkaConn string, dbSrv dbclient.Srv) (Srv, error) {
	glog.Info("Initializing IP Graph Kafka messenger")

	brokers := strings.Split(kafkaConn, ",")

	// Topics that the IP graph processor subscribes to - raw BMP data topics
	topics := []string{
		"gobmp.parsed.ls_node",           // IGP sync
		"gobmp.parsed.ls_link",           // IGP sync
		"gobmp.parsed.ls_prefix",         // IGP prefix sync (optional)
		"gobmp.parsed.ls_srv6_sid",       // IGP SRv6 sync (optional)
		"gobmp.parsed.peer",              // BGP peer sessions
		"gobmp.parsed.unicast_prefix_v4", // BGP IPv4 prefixes
		"gobmp.parsed.unicast_prefix_v6", // BGP IPv6 prefixes
	}

	config := sarama.NewConfig()
	config.ClientID = "ip-graph-processor"
	config.Consumer.Group.InstanceId = "ip-graph-processor"
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Version = sarama.V2_6_0_0

	consumer, err := sarama.NewConsumerGroup(brokers, "ip-graph-processor", config)
	if err != nil {
		return nil, err
	}

	k := &kafka{
		stopCh:   make(chan struct{}),
		brokers:  brokers,
		dbSrv:    dbSrv,
		config:   config,
		consumer: consumer,
		topics:   topics,
	}

	return k, nil
}

func (k *kafka) Start() error {
	glog.Infof("Starting IP Graph Kafka messenger, group: %s, topics: %v",
		"ip-graph-processor", k.topics)

	go func() {
		for {
			select {
			case <-k.stopCh:
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

	// Monitor consumer errors
	go func() {
		for err := range k.consumer.Errors() {
			glog.Errorf("Kafka consumer error: %v", err)
		}
	}()

	glog.Info("IP Graph Kafka messenger started successfully")
	return nil
}

func (k *kafka) Stop() error {
	glog.Info("Stopping IP Graph Kafka messenger...")
	close(k.stopCh)

	if err := k.consumer.Close(); err != nil {
		glog.Errorf("Error closing Kafka consumer: %v", err)
		return err
	}

	glog.Info("IP Graph Kafka messenger stopped")
	return nil
}
