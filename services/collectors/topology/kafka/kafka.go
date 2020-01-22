package kafka

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
        "strings"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/handler"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/log"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/openbmp"
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
		"openbmp.parsed.router", "openbmp.parsed.peer", "openbmp.parsed.base_attribute",
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
	c.Config.Consumer.Offsets.Initial = sarama.OffsetOldest
	c.Handler = hndlr
	return c, nil
}

func (c *Consumer) SetHandler(h handler.Handler) {
	c.Handler = h
}

func (c *Consumer) Start() error {
	fmt.Println("Starting Consumer...")
	consumer, err := cluster.NewConsumer(c.Brokers, c.GroupName, c.Topics, c.Config)
	if err != nil {
		return err
	}

	c.Consumer = consumer
	c.stop = make(chan bool)
	fmt.Println(c.Topics)

	// TODO: strategy: We want to process Peers first...
	// Refactor to allow out of order processing
	for {
		select {
		case msg, more := <-consumer.Messages():
			// TODO: uncomment markOffset (when not in DEV)
                        if more {
                                fmt.Println("Topic of record to be added:")
                                fmt.Println(msg.Topic)

                                openbmp_msg := strings.Split(string(msg.Value), "\n\n")
                                if len(openbmp_msg) != 2 {
                                    fmt.Println("Processing OpenBMP message from Kafka failed: something is wrong with header / data splitting.")
                                    consumer.MarkOffset(msg, "") // mark message as processed
                                    continue
                                } else {
                                    // fmt.Println("The headers of our OpenBMP message are:")
                                    // fmt.Println(openbmp_msg[0])
                                    // fmt.Println("The data of our OpenBMP message is:")

                                    openbmp_msg_data := strings.Split(string(openbmp_msg[1]), "\n")
                                    for _, element := range openbmp_msg_data {
                                        // fmt.Println("The current record to be processed is:")
                                        current_openbmp_record := strings.Split(element, "\t")
                                        // fmt.Println(current_openbmp_record)

                                        omsg := openbmp.NewMessage(msg.Topic, current_openbmp_record)
                                        // fmt.Println("The message created by openbmp.go is:")
                                        // fmt.Println(omsg)

                                        if omsg == nil { // error
                                            fmt.Println("Something failed")
                                            consumer.MarkOffset(msg, "") // mark message as processed
                                            fmt.Println("=============================================================")
                                            continue
                                        }
                                        c.Handler.Handle(omsg)
                                        fmt.Println("=============================================================")
                                    }
                                }
                                consumer.MarkOffset(msg, "") // mark message as processed
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
