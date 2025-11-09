// Copyright (c) 2022-2025 Cisco Systems, Inc. and its affiliates
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// The contents of this file are licensed under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with the
// License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package kafkamessenger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// KafkaMessenger handles Kafka message consumption for the IGP graph processor
type KafkaMessenger struct {
	consumer sarama.ConsumerGroup
	dbSrv    dbclient.DB
	stop     chan struct{}
	topics   []string
	brokers  []string
	groupID  string
}

// NewKafkaMessenger creates a new Kafka messenger for IGP graph processing
func NewKafkaMessenger(kafkaConn string, dbSrv dbclient.DB) (*KafkaMessenger, error) {
	if err := validateConnection(kafkaConn); err != nil {
		return nil, err
	}

	brokers := strings.Split(kafkaConn, ",")

	// Topics that the IGP graph processor subscribes to - raw BMP data topics
	topics := []string{
		"gobmp.parsed.ls_node",
		"gobmp.parsed.ls_link",
		"gobmp.parsed.ls_prefix",
		"gobmp.parsed.ls_srv6_sid",
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.MaxProcessingTime = 1 * time.Second
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Retry.Max = 3

	groupID := "igp-graph-processor"
	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaMessenger{
		consumer: consumer,
		dbSrv:    dbSrv,
		stop:     make(chan struct{}),
		topics:   topics,
		brokers:  brokers,
		groupID:  groupID,
	}, nil
}

// Start begins consuming Kafka messages
func (k *KafkaMessenger) Start() {
	glog.Infof("Starting Kafka messenger for IGP graph processor, group: %s, topics: %v",
		k.groupID, k.topics)

	go func() {
		defer k.consumer.Close()

		for {
			select {
			case <-k.stop:
				glog.Info("Kafka messenger stopping...")
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

	// Error handling goroutine
	go func() {
		for err := range k.consumer.Errors() {
			glog.Errorf("Kafka consumer error: %v", err)
		}
	}()
}

// Stop stops the Kafka messenger
func (k *KafkaMessenger) Stop() {
	glog.Info("Stopping Kafka messenger...")
	close(k.stop)
	time.Sleep(100 * time.Millisecond) // Give time for graceful shutdown
}

// MessageHandler implements sarama.ConsumerGroupHandler
type MessageHandler struct {
	dbSrv dbclient.DB
}

// Setup is called when a consumer group session starts
func (h *MessageHandler) Setup(sarama.ConsumerGroupSession) error {
	glog.V(5).Info("IGP graph consumer group session setup")
	return nil
}

// Cleanup is called when a consumer group session ends
func (h *MessageHandler) Cleanup(sarama.ConsumerGroupSession) error {
	glog.V(5).Info("IGP graph consumer group session cleanup")
	return nil
}

// ConsumeClaim processes messages from a partition
func (h *MessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			if err := h.processMessage(message); err != nil {
				glog.Errorf("Failed to process message from topic %s, partition %d, offset %d: %v",
					message.Topic, message.Partition, message.Offset, err)
				// Continue processing other messages even if one fails
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *MessageHandler) processMessage(message *sarama.ConsumerMessage) error {
	glog.V(9).Infof("Processing message from topic: %s, partition: %d, offset: %d",
		message.Topic, message.Partition, message.Offset)

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
	default:
		glog.V(5).Infof("Ignoring message from unsupported topic: %s", message.Topic)
		return nil
	}

	// Parse the raw BMP message
	var bmpData map[string]interface{}
	if err := json.Unmarshal(message.Value, &bmpData); err != nil {
		return fmt.Errorf("failed to parse BMP message: %w", err)
	}

	// Add message key if present
	if message.Key != nil {
		bmpData["_message_key"] = string(message.Key)
	}

	// Add topic information for processing context
	bmpData["_topic"] = message.Topic
	bmpData["_partition"] = message.Partition
	bmpData["_offset"] = message.Offset

	// Re-marshal for storage
	processedMessage, err := json.Marshal(bmpData)
	if err != nil {
		return fmt.Errorf("failed to marshal processed message: %w", err)
	}

	// Store the processed message
	return h.dbSrv.StoreMessage(msgType, processedMessage)
}

func validateConnection(kafkaConn string) error {
	if kafkaConn == "" {
		return fmt.Errorf("kafka connection string cannot be empty")
	}
	return nil
}
