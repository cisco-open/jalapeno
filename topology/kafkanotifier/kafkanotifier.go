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

package kafkanotifier

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cisco-open/jalapeno/topology/dbclient"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// Define constants for each topic name
const (
	PeerEventTopic            = "gobmp.parsed.peer_events"
	UnicastPrefixEventTopic   = "gobmp.parsed.unicast_prefix_events"
	UnicastPrefixV4EventTopic = "gobmp.parsed.unicast_prefix_v4_events"
	UnicastPrefixV6EventTopic = "gobmp.parsed.unicast_prefix_v6_events"
	LSNodeEventTopic          = "gobmp.parsed.ls_node_events"
	LSLinkEventTopic          = "gobmp.parsed.ls_link_events"
	L3VPNEventTopic           = "gobmp.parsed.l3vpn_events"
	L3VPNV4EventTopic         = "gobmp.parsed.l3vpn_v4_events"
	L3VPNV6EventTopic         = "gobmp.parsed.l3vpn_v6_events"
	LSPrefixEventTopic        = "gobmp.parsed.ls_prefix_events"
	LSSRv6SIDEventTopic       = "gobmp.parsed.ls_srv6_sid_events"
	EVPNEventTopic            = "gobmp.parsed.evpn_events"
	SRPolicyEventTopic        = "gobmp.parsed.sr_policy_events"
	SRPolicyV4EventTopic      = "gobmp.parsed.sr_policy_v4_events"
	SRPolicyV6EventTopic      = "gobmp.parsed.sr_policy_v6_events"
	FlowspecEventTopic        = "gobmp.parsed.flowspec_events"
	FlowspecV4EventTopic      = "gobmp.parsed.flowspec_v4_events"
	FlowspecV6EventTopic      = "gobmp.parsed.flowspec_v6_events"
)

var (
	brockerConnectTimeout = 10 * time.Second
	topicCreateTimeout    = 1 * time.Second
	// topic Retention for events is 5 minutes
	topicRetention = "300000"
)

var (
	// topics defines a list of topic to initialize and connect,
	// initialization is done as a part of NewKafkaPublisher func.
	topicNames = []string{
		PeerEventTopic,
		UnicastPrefixEventTopic,
		UnicastPrefixV4EventTopic,
		UnicastPrefixV6EventTopic,
		LSNodeEventTopic,
		LSLinkEventTopic,
		L3VPNEventTopic,
		L3VPNV4EventTopic,
		L3VPNV6EventTopic,
		LSPrefixEventTopic,
		LSSRv6SIDEventTopic,
		EVPNEventTopic,
		SRPolicyEventTopic,
		SRPolicyV4EventTopic,
		SRPolicyV6EventTopic,
		FlowspecEventTopic,
		FlowspecV4EventTopic,
		FlowspecV6EventTopic,
	}
)

type EventMessage struct {
	TopicType dbclient.CollectionType
	Key       string `json:"_key"`
	ID        string `json:"_id"`
	Action    string `json:"action"`
}

type Event interface {
	EventNotification(*EventMessage) error
}

type notifier struct {
	broker   *sarama.Broker
	config   *sarama.Config
	producer sarama.SyncProducer
}

func (n *notifier) EventNotification(msg *EventMessage) error {
	switch msg.TopicType {
	case bmp.PeerStateChangeMsg:
		return n.triggerNotification(PeerEventTopic, msg)
	case bmp.UnicastPrefixMsg:
		return n.triggerNotification(UnicastPrefixEventTopic, msg)
	case bmp.UnicastPrefixV4Msg:
		return n.triggerNotification(UnicastPrefixV4EventTopic, msg)
	case bmp.UnicastPrefixV6Msg:
		return n.triggerNotification(UnicastPrefixV6EventTopic, msg)
	case bmp.LSNodeMsg:
		return n.triggerNotification(LSNodeEventTopic, msg)
	case bmp.LSLinkMsg:
		return n.triggerNotification(LSLinkEventTopic, msg)
	case bmp.L3VPNMsg:
		return n.triggerNotification(L3VPNEventTopic, msg)
	case bmp.L3VPNV4Msg:
		return n.triggerNotification(L3VPNV4EventTopic, msg)
	case bmp.L3VPNV6Msg:
		return n.triggerNotification(L3VPNV6EventTopic, msg)
	case bmp.LSPrefixMsg:
		return n.triggerNotification(LSPrefixEventTopic, msg)
	case bmp.LSSRv6SIDMsg:
		return n.triggerNotification(LSSRv6SIDEventTopic, msg)
	case bmp.EVPNMsg:
		return n.triggerNotification(EVPNEventTopic, msg)
	case bmp.SRPolicyMsg:
		return n.triggerNotification(SRPolicyEventTopic, msg)
	case bmp.SRPolicyV4Msg:
		return n.triggerNotification(SRPolicyV4EventTopic, msg)
	case bmp.SRPolicyV6Msg:
		return n.triggerNotification(SRPolicyV6EventTopic, msg)
	case bmp.FlowspecMsg:
		return n.triggerNotification(SRPolicyEventTopic, msg)
	case bmp.FlowspecV4Msg:
		return n.triggerNotification(SRPolicyV4EventTopic, msg)
	case bmp.FlowspecV6Msg:
		return n.triggerNotification(SRPolicyV6EventTopic, msg)
	}

	return fmt.Errorf("unknown topic type %d", msg.TopicType)
}

func NewKafkaNotifier(kafkaSrv string) (Event, error) {
	glog.Infof("Initializing Kafka events producer client")
	if err := validator(kafkaSrv); err != nil {
		glog.Errorf("Failed to validate Kafka server address %s with error: %+v", kafkaSrv, err)
		return nil, err
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0

	br := sarama.NewBroker(kafkaSrv)
	if err := br.Open(config); err != nil {
		if err != sarama.ErrAlreadyConnected {
			return nil, err
		}
	}
	if err := waitForBrokerConnection(br, brockerConnectTimeout); err != nil {
		glog.Errorf("failed to open connection to the broker with error: %+v\n", err)
		return nil, err
	}
	glog.V(5).Infof("Connected to broker: %s id: %d\n", br.Addr(), br.ID())

	for _, t := range topicNames {
		if err := ensureTopic(br, topicCreateTimeout, t); err != nil {
			return nil, err
		}
	}
	producer, err := sarama.NewSyncProducer([]string{kafkaSrv}, config)
	if err != nil {
		return nil, err
	}
	glog.V(5).Infof("Initialized Kafka Sync producer")

	return &notifier{
		broker:   br,
		config:   config,
		producer: producer,
	}, nil
}

func (n *notifier) triggerNotification(topic string, msg *EventMessage) error {
	k := sarama.ByteEncoder{}
	k = []byte(msg.Key)
	m := sarama.ByteEncoder{}
	m, _ = json.Marshal(msg)
	_, _, err := n.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   k,
		Value: m,
	})

	return err
}

func validator(addr string) error {
	host, port, _ := net.SplitHostPort(addr)
	if host == "" || port == "" {
		return fmt.Errorf("host or port cannot be ''")
	}
	// Try to resolve if the hostname was used in the address
	if ip, err := net.LookupIP(host); err != nil || ip == nil {
		// Check if IP address was used in address instead of a host name
		if net.ParseIP(host) == nil {
			return fmt.Errorf("fail to parse host part of address")
		}
	}
	np, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("fail to parse port with error: %w", err)
	}
	if np == 0 || np > math.MaxUint16 {
		return fmt.Errorf("the value of port is invalid")
	}
	return nil
}

func ensureTopic(br *sarama.Broker, timeout time.Duration, topicName string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	tout := time.NewTimer(timeout)
	topic := &sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topicName: {
				NumPartitions:     1,
				ReplicationFactor: 1,
				ConfigEntries: map[string]*string{
					"retention.ms": &topicRetention,
				},
			},
		},
		Timeout: timeout,
	}

	for {
		t, err := br.CreateTopics(topic)
		if err != nil {
			return err
		}
		if e, ok := t.TopicErrors[topicName]; ok {
			if e.Err == sarama.ErrTopicAlreadyExists || e.Err == sarama.ErrNoError {
				return nil
			}
			if e.Err != sarama.ErrRequestTimedOut {
				return e
			}
		}
		select {
		case <-ticker.C:
			continue
		case <-tout.C:
			return fmt.Errorf("timeout waiting for topic %s", topicName)
		}
	}
}

func waitForBrokerConnection(br *sarama.Broker, timeout time.Duration) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	tout := time.NewTimer(timeout)
	for {
		ok, err := br.Connected()
		if ok {
			return nil
		}
		if err != nil {
			return err
		}
		select {
		case <-ticker.C:
			continue
		case <-tout.C:
			return fmt.Errorf("timeout waiting for the connection to the broker %s", br.Addr())
		}
	}

}
