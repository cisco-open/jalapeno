package kafkamessenger

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/jalapeno-sdn/jalapeno/lsv4_edge/pkg/dbclient"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/tools"
)

// Define constants for each topic name
const (
	ls_linkMessageTopic = "gobmp.parsed.ls_link_events"
)

var (
	topics = map[string]int{
		ls_linkMessageTopic: bmp.LSLINKMsg,
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
	glog.Infof("LSv4 Edge kafka reader")
	if err := tools.HostAddrValidator(kafkaSrv); err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.ClientID = "lsv4_edge"
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

func (k *kafka) topicReader(topicType int, topicName string) {
	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		partitions, _ := k.master.Partitions(topicName)
		// this only consumes partition no 1, you would probably want to consume all partitions
		consumer, err := k.master.ConsumePartition(topicName, partitions[0], sarama.OffsetOldest)
		if nil != err {
			glog.Infof("Consumer error: %+v", err)
			select {
			case <-ticker.C:
			case <-k.stopCh:
				return
			}
			continue
		}
		glog.Infof("Starting Kafka reader for topic: %s", topicName)
		for {
			select {
			case msg := <-consumer.Messages():
				if msg == nil {
					continue
				}
				k.db.ProcessMessage(topicType, msg.Value)
			case consumerError := <-consumer.Errors():
				if consumerError == nil {
					break
				}
				glog.Errorf("error %+v for topic: %s, partition: %s ", consumerError.Err, string(consumerError.Topic), string(consumerError.Partition))
			case <-k.stopCh:
				return
			}
		}
	}
}
