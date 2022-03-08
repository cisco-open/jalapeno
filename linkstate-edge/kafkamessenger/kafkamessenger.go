// Copyright (c) 2022 Cisco Systems, Inc. and its affiliates
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
		kafkanotifier.LSLinkEventTopic: bmp.LSLinkMsg,
		kafkanotifier.LSNodeEventTopic: bmp.LSNodeMsg,
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
	glog.Infof("LS Node Vertex kafka reader")
	if err := tools.HostAddrValidator(kafkaSrv); err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.ClientID = "lslinknode-edge-collection"
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
		glog.Infof("Starting Kafka reader for topic: %s topicType: %d", topicName, topicType)
		for {
			select {
			case msg := <-consumer.Messages():
				if msg == nil {
					continue
				}
				if err := k.db.StoreMessage(topicType, msg.Value); err != nil {
					glog.Errorf("failed to process a message from topic %s with error: %+v", topicName, err)
				}
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
