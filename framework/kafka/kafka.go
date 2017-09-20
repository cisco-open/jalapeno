package kafka

import (
	"errors"
	"log"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/kafka/handler"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/openbmp"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

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

func New(cfg Config) (*Consumer, error) {
	c := &Consumer{Config: cluster.NewConfig(), Topics: cfg.Topics, GroupName: cfg.ConsumerGroup, Brokers: cfg.Brokers}
	if len(c.Topics) == 0 {
		c.Topics = DefaultTopics()
	}
	if len(c.GroupName) == 0 {
		c.GroupName = "OpenBMPConsumerGroup"
	}
	if len(c.Brokers) == 0 {
		return nil, errors.New("A list of kafka brokers is required")
	}
	c.Config.Consumer.Return.Errors = true
	c.Config.Group.Return.Notifications = true
	c.Config.Group.PartitionStrategy = cluster.StrategyRoundRobin
	c.Config.Config.Consumer.Offsets.Initial = sarama.OffsetOldest
	c.Handler = handler.NewDefaultHandler()
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
				log.Printf("Error: %s\n", err.Error())
			}
		case ntf, more := <-consumer.Notifications():
			if more {
				log.Printf("Rebalanced: %+v\n", ntf)
			}
		case <-c.stop:
			err := c.Consumer.Close()
			return err
		}
	}
}

func (c *Consumer) Stop() {
	c.stop <- true
}
