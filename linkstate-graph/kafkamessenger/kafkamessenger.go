package kafkamessenger

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/cisco-open/jalapeno/topology/dbclient"
	"github.com/cisco-open/jalapeno/topology/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/tools"
)

var (
	topics = map[string]dbclient.CollectionType{
		kafkanotifier.LSSRv6SIDEventTopic: bmp.LSSRv6SIDMsg,
		kafkanotifier.LSNodeEventTopic:    bmp.LSNodeMsg,
		kafkanotifier.LSPrefixEventTopic:  bmp.LSPrefixMsg,
		kafkanotifier.LSLinkEventTopic:    bmp.LSLinkMsg,
	}
)

// Srv defines required method of a processor server
type Srv interface {
	Start() error
	Stop() error
}

type kafka struct {
	stopCh  chan struct{}
	brokers []string
	db      dbclient.DB
	config  *sarama.Config
	master  sarama.Consumer
}

// NewKafkaMessenger returns an instance of a kafka consumer acting as a messenger server
func NewKafkaMessenger(kafkaSrv string, db dbclient.DB) (Srv, error) {
	glog.Infof("ls events kafka reader")
	if err := tools.HostAddrValidator(kafkaSrv); err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.ClientID = "ls-node-collection"
	config.Consumer.Return.Errors = true
	config.Version = sarama.V0_11_0_0

	brokers := []string{kafkaSrv}

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}
	k := &kafka{
		stopCh: make(chan struct{}),
		config: config,
		master: master,
		db:     db,
	}

	return k, nil
}

func (k *kafka) Start() error {
	// Starting readers for each topic name and type defined in topics map
	for topicName, topicType := range topics {
		go k.topicReader(topicType, topicName)
	}

	return nil
}

func (k *kafka) Stop() error {
	close(k.stopCh)
	k.master.Close()
	return nil
}

func (k *kafka) topicReader(topicType dbclient.CollectionType, topicName string) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		partitions, err := k.master.Partitions(topicName)
		if err != nil {
			glog.Errorf("Failed to get partitions for topic %s: %v", topicName, err)
			select {
			case <-ticker.C:
				continue
			case <-k.stopCh:
				return
			}
		}

		// Consider consuming from all partitions
		for _, partition := range partitions {
			consumer, err := k.master.ConsumePartition(topicName, partition, sarama.OffsetOldest)
			if err != nil {
				glog.Errorf("Failed to create consumer for topic %s partition %d: %v",
					topicName, partition, err)
				continue
			}

			go k.handlePartition(consumer, topicType, topicName)
		}

		// Wait for stop signal
		<-k.stopCh
		return
	}
}

func (k *kafka) handlePartition(consumer sarama.PartitionConsumer, topicType dbclient.CollectionType, topicName string) {
	defer consumer.Close()

	for {
		select {
		case msg := <-consumer.Messages():
			if msg == nil {
				continue
			}
			if err := k.db.StoreMessage(topicType, msg.Value); err != nil {
				glog.Errorf("Failed to process message from topic %s: %v", topicName, err)
			}
		case err := <-consumer.Errors():
			if err != nil {
				glog.Errorf("Consumer error for topic %s: %v", topicName, err)
			}
		case <-k.stopCh:
			return
		}
	}
}
