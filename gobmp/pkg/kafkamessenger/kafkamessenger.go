package kafkamessenger

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/tools"
	"github.com/sbezverk/topology/pkg/dbclient"
)

// Define constants for each topic name
const (
	peerTopic              = "gobmp.parsed.peer"
	unicastMessageTopic    = "gobmp.parsed.unicast_prefix"
	unicastMessageV4Topic  = "gobmp.parsed.unicast_prefix_v4"
	unicastMessageV6Topic  = "gobmp.parsed.unicast_prefix_v6"
	lsNodeMessageTopic     = "gobmp.parsed.ls_node"
	lsLinkMessageTopic     = "gobmp.parsed.ls_link"
	l3vpnMessageTopic      = "gobmp.parsed.l3vpn"
	l3vpnMessageV4Topic    = "gobmp.parsed.l3vpn_v4"
	l3vpnMessageV6Topic    = "gobmp.parsed.l3vpn_v6"
	lsPrefixMessageTopic   = "gobmp.parsed.ls_prefix"
	lsSRv6SIDMessageTopic  = "gobmp.parsed.ls_srv6_sid"
	evpnMessageTopic       = "gobmp.parsed.evpn"
	srPolicyMessageTopic   = "gobmp.parsed.sr_policy"
	srPolicyMessageV4Topic = "gobmp.parsed.sr_policy_v4"
	srPolicyMessageV6Topic = "gobmp.parsed.sr_policy_v6"
)

var (
	topics = map[string]dbclient.CollectionType{
		peerTopic:              bmp.PeerStateChangeMsg,
		unicastMessageTopic:    bmp.UnicastPrefixMsg,
		unicastMessageV4Topic:  bmp.UnicastPrefixV4Msg,
		unicastMessageV6Topic:  bmp.UnicastPrefixV6Msg,
		lsNodeMessageTopic:     bmp.LSNodeMsg,
		lsLinkMessageTopic:     bmp.LSLinkMsg,
		l3vpnMessageTopic:      bmp.L3VPNMsg,
		l3vpnMessageV4Topic:    bmp.L3VPNV4Msg,
		l3vpnMessageV6Topic:    bmp.L3VPNV6Msg,
		lsPrefixMessageTopic:   bmp.LSPrefixMsg,
		lsSRv6SIDMessageTopic:  bmp.LSSRv6SIDMsg,
		evpnMessageTopic:       bmp.EVPNMsg,
		srPolicyMessageTopic:   bmp.SRPolicyMsg,
		srPolicyMessageV4Topic: bmp.SRPolicyV4Msg,
		srPolicyMessageV6Topic: bmp.SRPolicyV6Msg,
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
	config.ClientID = "gobmp-consumer" + "_" + strconv.Itoa(rand.Intn(1000))
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
	for {
		// Loop until either a topic becomes available at the broker or stop signal is received
		partitions, err := k.master.Partitions(topicName)
		if nil != err {
			glog.Errorf("fail to get partitions for the topic %s with error: %+v", topicName, err)
			select {
			case <-ticker.C:
			case <-k.stopCh:
				return
			}
			continue
		}
		// Loop until either a topic's partition becomes consumable or stop signal is received
		consumer, err := k.master.ConsumePartition(topicName, partitions[0], sarama.OffsetOldest)
		if nil != err {
			glog.Errorf("fail to consume partition for the topic %s with error: %+v", topicName, err)
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
