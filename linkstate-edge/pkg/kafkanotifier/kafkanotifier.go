package kafkanotifier

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/jalapeno/topology/pkg/kafkanotifier"
)

const LSNodeEdgeTopic = "jalapeno.ls_node_edge_events"

var (
	brockerConnectTimeout = 10 * time.Second
	topicCreateTimeout    = 1 * time.Second
	// topic Retention for events is 5 minutes
	topicRetention = "300000"
)

type Notifier struct {
	broker   *sarama.Broker
	config   *sarama.Config
	producer sarama.SyncProducer
}

func NewKafkaNotifier(kafkaSrv string) (*Notifier, error) {
	glog.Infof("Initializing Kafka events producer client")
	if err := validator(kafkaSrv); err != nil {
		glog.Errorf("Failed to validate Kafka server address %s with error: %+v", kafkaSrv, err)
		return &Notifier{}, err
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0

	br := sarama.NewBroker(kafkaSrv)
	if err := br.Open(config); err != nil {
		if err != sarama.ErrAlreadyConnected {
			return &Notifier{}, err
		}
	}
	if err := waitForBrokerConnection(br, brockerConnectTimeout); err != nil {
		glog.Errorf("failed to open connection to the broker with error: %+v\n", err)
		return &Notifier{}, err
	}
	glog.V(5).Infof("Connected to broker: %s id: %d\n", br.Addr(), br.ID())

	if err := ensureTopic(br, topicCreateTimeout, LSNodeEdgeTopic); err != nil {
		return &Notifier{}, err
	}
	producer, err := sarama.NewSyncProducer([]string{kafkaSrv}, config)
	if err != nil {
		return &Notifier{}, err
	}
	glog.V(5).Infof("Initialized Kafka Sync producer")

	return &Notifier{
		broker:   br,
		config:   config,
		producer: producer,
	}, nil
}

func TriggerNotification(n *Notifier, topic string, msg *kafkanotifier.EventMessage) error {
	k := sarama.ByteEncoder{}
	k = []byte(msg.Key)
	m := sarama.ByteEncoder{}
	m, _ = json.Marshal(msg)
	_, _, err := n.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   k,
		Value: m,
	})

	return err
}

func validator(addr string) error {
	host, port, _ := net.SplitHostPort(addr)
	if host == "" || port == "" {
		return fmt.Errorf("host or port cannot be ''")
	}
	// Try to resolve if the hostname was used in the address
	if ip, err := net.LookupIP(host); err != nil || ip == nil {
		// Check if IP address was used in address instead of a host name
		if net.ParseIP(host) == nil {
			return fmt.Errorf("fail to parse host part of address")
		}
	}
	np, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("fail to parse port with error: %w", err)
	}
	if np == 0 || np > math.MaxUint16 {
		return fmt.Errorf("the value of port is invalid")
	}
	return nil
}

func ensureTopic(br *sarama.Broker, timeout time.Duration, topicName string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	tout := time.NewTimer(timeout)
	topic := &sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topicName: {
				NumPartitions:     1,
				ReplicationFactor: 1,
				ConfigEntries: map[string]*string{
					"retention.ms": &topicRetention,
				},
			},
		},
	}

	for {
		t, err := br.CreateTopics(topic)
		if err != nil {
			return err
		}
		if e, ok := t.TopicErrors[topicName]; ok {
			if e.Err == sarama.ErrTopicAlreadyExists || e.Err == sarama.ErrNoError {
				return nil
			}
			if e.Err != sarama.ErrRequestTimedOut {
				return e
			}
		}
		select {
		case <-ticker.C:
			continue
		case <-tout.C:
			return fmt.Errorf("timeout waiting for topic %s", topicName)
		}
	}
}

func waitForBrokerConnection(br *sarama.Broker, timeout time.Duration) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	tout := time.NewTimer(timeout)
	for {
		ok, err := br.Connected()
		if ok {
			return nil
		}
		if err != nil {
			return err
		}
		select {
		case <-ticker.C:
			continue
		case <-tout.C:
			return fmt.Errorf("timeout waiting for the connection to the broker %s", br.Addr())
		}
	}

}