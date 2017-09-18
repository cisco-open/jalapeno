package kafka

import (
	"log"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/kafka/handler"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/openbmp"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type Consumer struct {
	Config    *Config
	Topics    []string
	Brokers   []string
	GroupName string
	*cluster.Consumer
	stop    chan bool
	Handler handler.Handler
}

type Config struct {
	*cluster.Config
}

func NewConfig() *Config {
	return &Config{
		cluster.NewConfig(),
	}
}

func DefaultTopics() []string {
	return []string{"openbmp.parsed.collector", "openbmp.parsed.collector",
		"openbmp.parsed.peer", "openbmp.parsed.base_attribute",
		"openbmp.parsed.unicast_prefix", "openbmp.parsed.ls_node",
		"openbmp.parsed.ls_link", "openbmp.parsed.ls_prefix"}
}

func New(brokers []string, groupName string, topics ...string) *Consumer {
	c := &Consumer{Config: NewConfig(), Topics: topics, GroupName: groupName, Brokers: brokers}
	if len(topics) == 0 {
		c.Topics = DefaultTopics()
	}
	c.Config.Consumer.Return.Errors = true
	c.Config.Group.Return.Notifications = true
	c.Config.Group.PartitionStrategy = cluster.StrategyRoundRobin
	c.Config.Config.Consumer.Offsets.Initial = sarama.OffsetOldest
	c.Handler = handler.NewDefaultHandler()
	return c
}

func (c *Consumer) SetHandler(h handler.Handler) {
	c.Handler = h
}

func (c *Consumer) Start() error {
	consumer, err := cluster.NewConsumer(c.Brokers, c.GroupName, c.Topics, c.Config.Config)
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
