package kafkamessenger

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/tools"
	"github.com/jalapeno-sdn/topology/pkg/dbclient"
)

// Define constants for each topic name
const (
	peerTopic             = "gobmp.parsed.peer"
	unicastMessageTopic   = "gobmp.parsed.unicast_prefix"
	lsNodeMessageTopic    = "gobmp.parsed.ls_node"
	lsLinkMessageTopic    = "gobmp.parsed.ls_link"
	l3vpnMessageTopic     = "gobmp.parsed.l3vpn"
	lsPrefixMessageTopic  = "gobmp.parsed.ls_prefix"
	lsSRv6SIDMessageTopic = "gobmp.parsed.ls_srv6_sid"
	evpnMessageTopic      = "gobmp.parsed.evpn"
)

var (
	topics = map[string]int{
		peerTopic:             bmp.PeerStateChangeMsg,
		unicastMessageTopic:   bmp.UnicastPrefixMsg,
		lsNodeMessageTopic:    bmp.LSNodeMsg,
		lsLinkMessageTopic:    bmp.LSLinkMsg,
		l3vpnMessageTopic:     bmp.L3VPNMsg,
		lsPrefixMessageTopic:  bmp.LSPrefixMsg,
		lsSRv6SIDMessageTopic: bmp.LSSRv6SIDMsg,
		evpnMessageTopic:      bmp.EVPNMsg,
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
	glog.Infof("NewKafkaMessenger")
	if err := tools.HostAddrValidator(kafkaSrv); err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.ClientID = "gobmp-consumer"
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
				k.db.StoreMessage(topicType, msg.Value)
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
