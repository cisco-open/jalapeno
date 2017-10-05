package kafka

import (
	"errors"
	"math/rand"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/handler"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/openbmp"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Consumer struct {
	Config    *cluster.Config
	Topics    []string
	Brokers   []string
	GroupName string
	*cluster.Consumer
	stop    chan bool
	Handler handler.Handler
}

type Config struct {
	Brokers       []string `json:"brokers" desc:"List of Kafka Brokers"`
	Topics        []string `json:"topics" desc:"Optional subset of openbmp topics."`
	ConsumerGroup string   `desc:"Optional Consumer Group"`
}

func NewConfig() Config {
	return Config{}
}

func DefaultTopics() []string {
	return []string{"openbmp.parsed.collector", "openbmp.parsed.collector",
		"openbmp.parsed.peer", "openbmp.parsed.base_attribute",
		"openbmp.parsed.unicast_prefix", "openbmp.parsed.ls_node",
		"openbmp.parsed.ls_link", "openbmp.parsed.ls_prefix"}
}

func New(cfg Config, hndlr handler.Handler) (*Consumer, error) {
	c := &Consumer{Config: cluster.NewConfig(), Topics: cfg.Topics, GroupName: cfg.ConsumerGroup, Brokers: cfg.Brokers}
	if len(c.Topics) == 0 {
		c.Topics = DefaultTopics()
	}
	if len(c.GroupName) == 0 { //TODO: REMOVE IN PROD
		c.GroupName = "OpenBMPConsumerGroup" + randStringBytesMask(8)
	}
	if len(c.Brokers) == 0 {
		return nil, errors.New("A list of kafka brokers is required")
	}
	c.Config.Consumer.Return.Errors = true
	c.Config.Group.Return.Notifications = true
	c.Config.Group.PartitionStrategy = cluster.StrategyRoundRobin
	c.Config.Config.Consumer.Offsets.Initial = sarama.OffsetOldest
	c.Handler = hndlr
	return c, nil
}

func (c *Consumer) SetHandler(h handler.Handler) {
	c.Handler = h
}

func (c *Consumer) Start() error {
	consumer, err := cluster.NewConsumer(c.Brokers, c.GroupName, c.Topics, c.Config)
	if err != nil {
		return err
	}

	c.Consumer = consumer
	c.stop = make(chan bool)

	// TODO: strategy: We want to process Peers first...
	// Refactor to allow out of order processing
	for {
		select {
		case msg, more := <-consumer.Messages():
			// TODO: uncomment markOffset (when not in DEV)
			if more {
				omsg := openbmp.NewMessage(msg.Topic, msg.Value)
				if omsg == nil { // error
					//	consumer.MarkOffset(msg, "") // mark message as processed
					continue
				}
				//consumer.MarkOffset(msg, "") // mark message as processed
				c.Handler.Handle(omsg)
			}
		case err, more := <-consumer.Errors():
			// TODO: add error/notification channel.
			if more {
				log.Errorf("Error: %s\n", err.Error())
			}
		case ntf, more := <-consumer.Notifications():
			if more {
				log.Infof("Rebalanced: %+v\n", ntf)
			}
		case <-c.stop:
			err := c.Consumer.Close()
			log.Infof("Consumer Closed...")
			return err
		}
	}
}

func (c *Consumer) Stop() {
	c.stop <- true
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randStringBytesMask(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
