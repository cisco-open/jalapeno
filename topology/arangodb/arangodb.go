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

package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/cisco-open/jalapeno/topology/dbclient"
	"github.com/cisco-open/jalapeno/topology/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/tools"
)

const (
	concurrentWorkers = 1024
)

var (
	collections = map[dbclient.CollectionType]*collectionProperties{
		dbclient.PeerStateChange: {name: "peer", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.LSLink:          {name: "ls_link", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.LSNode:          {name: "ls_node", isVertex: true, options: &driver.CreateCollectionOptions{}},
		dbclient.LSPrefix:        {name: "ls_prefix", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.LSSRv6SID:       {name: "ls_srv6_sid", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.L3VPN:           {name: "l3vpn_prefix", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.L3VPNV4:         {name: "l3vpn_v4_prefix", isVertex: true, options: &driver.CreateCollectionOptions{}},
		dbclient.L3VPNV6:         {name: "l3vpn_v6_prefix", isVertex: true, options: &driver.CreateCollectionOptions{}},
		dbclient.UnicastPrefix:   {name: "unicast_prefix", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.UnicastPrefixV4: {name: "unicast_prefix_v4", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.UnicastPrefixV6: {name: "unicast_prefix_v6", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.SRPolicy:        {name: "sr_policy", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.SRPolicyV4:      {name: "sr_policy_v4", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.SRPolicyV6:      {name: "sr_policy_v6", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.Flowspec:        {name: "flowspec", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.FlowspecV4:      {name: "flowspec_v4", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.FlowspecV6:      {name: "flowspec_v6", isVertex: false, options: &driver.CreateCollectionOptions{}},
	}
)

// collectionProperties defines a collection specific properties
// TODO this information should be configurable without recompiling code.
type collectionProperties struct {
	name     string
	isVertex bool
	options  *driver.CreateCollectionOptions
}

type arangoDB struct {
	dbclient.DB
	*ArangoConn
	stop             chan struct{}
	collections      map[dbclient.CollectionType]*collection
	notifyCompletion bool
	notifier         kafkanotifier.Event
}

// NewDBSrvClient returns an instance of a DB server client process
func NewDBSrvClient(arangoSrv, user, pass, dbname string, notifier kafkanotifier.Event) (dbclient.Srv, error) {
	if err := tools.URLAddrValidation(arangoSrv); err != nil {
		return nil, err
	}
	arangoConn, err := NewArango(ArangoConfig{
		URL:      arangoSrv,
		User:     user,
		Password: pass,
		Database: dbname,
	})
	if err != nil {
		return nil, err
	}
	arango := &arangoDB{
		stop:        make(chan struct{}),
		collections: make(map[dbclient.CollectionType]*collection),
	}
	arango.DB = arango
	arango.ArangoConn = arangoConn
	if notifier != nil {
		arango.notifyCompletion = true
		arango.notifier = notifier
	}
	// Init collections
	for t, n := range collections {
		if err := arango.ensureCollection(n, t); err != nil {
			return nil, err
		}
	}

	// Create additional graphs: igpv4_graph and igpv6_graph
	if err := arango.ensureAdditionalGraphs(); err != nil {
		return nil, err
	}

	return arango, nil
}

func (a *arangoDB) ensureCollection(p *collectionProperties, collectionType dbclient.CollectionType) error {
	if _, ok := a.collections[collectionType]; !ok {
		a.collections[collectionType] = &collection{
			queue:          make(chan *queueMsg),
			stats:          &stats{},
			stop:           a.stop,
			arango:         a,
			collectionType: collectionType,
			properties:     p,
		}
		switch collectionType {
		case bmp.PeerStateChangeMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.LSLinkMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.LSNodeMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.LSPrefixMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.LSSRv6SIDMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.L3VPNMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.L3VPNV4Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.L3VPNV6Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.UnicastPrefixMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.UnicastPrefixV4Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.UnicastPrefixV6Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.SRPolicyMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.SRPolicyV4Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.SRPolicyV6Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.FlowspecMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.FlowspecV4Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.FlowspecV6Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		default:
			return fmt.Errorf("unknown collection type %d", collectionType)
		}
	}
	var ci driver.Collection
	var err error
	// There are two possible collection types, base type and edge type
	// for Edge type a collection must be created as a Vertex collection
	if a.collections[collectionType].properties.isVertex {
		graph, err := a.ensureGraph(a.collections[collectionType].properties.name)
		if err != nil {
			return err
		}
		// Check if the vertex collection already exists
		ci, err = graph.VertexCollection(context.TODO(), a.collections[collectionType].properties.name)
		if err != nil {
			if !driver.IsArangoErrorWithErrorNum(err, driver.ErrArangoDataSourceNotFound) {
				return err
			}
			// Collection does not exist, attempting to create it
			ci, err = graph.CreateVertexCollection(context.TODO(), a.collections[collectionType].properties.name)
			if err != nil {
				return err
			}
		}
		a.collections[collectionType].topicCollection = ci
		return nil
	}
	ci, err = a.db.Collection(context.TODO(), a.collections[collectionType].properties.name)
	if err != nil {
		if !driver.IsArangoErrorWithErrorNum(err, driver.ErrArangoDataSourceNotFound) {
			return err
		}
		ci, err = a.db.CreateCollection(context.TODO(), a.collections[collectionType].properties.name, a.collections[collectionType].properties.options)
		if err != nil {
			return err
		}
	}
	a.collections[collectionType].topicCollection = ci

	return nil
}

func (a *arangoDB) ensureGraph(name string) (driver.Graph, error) {
	var edgeDefinition driver.EdgeDefinition
	edgeDefinition.Collection = name + "_edge"
	edgeDefinition.From = []string{name}
	edgeDefinition.To = []string{name}

	var options driver.CreateGraphOptions
	options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}
	graph, err := a.db.Graph(context.TODO(), name)
	if err == nil {
		graph.Remove(context.TODO())
		return a.db.CreateGraphV2(context.TODO(), name, &options)
	}
	if !driver.IsArangoErrorWithErrorNum(err, 1924) {
		return nil, err
	}

	return a.db.CreateGraphV2(context.TODO(), name, &options)
}

// ensureAdditionalGraphs creates the igpv4_graph and igpv6_graph graphs
func (a *arangoDB) ensureAdditionalGraphs() error {
	additionalGraphs := []string{"igpv4_graph", "igpv6_graph"}

	for _, graphName := range additionalGraphs {
		if err := a.ensureSpecificGraph(graphName); err != nil {
			return fmt.Errorf("failed to create %s: %w", graphName, err)
		}
	}

	return nil
}

// ensureSpecificGraph creates a specific graph with the given name
func (a *arangoDB) ensureSpecificGraph(graphName string) error {
	var edgeDefinition driver.EdgeDefinition
	edgeDefinition.Collection = graphName + "_edge"
	edgeDefinition.From = []string{graphName}
	edgeDefinition.To = []string{graphName}

	var options driver.CreateGraphOptions
	options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}

	graph, err := a.db.Graph(context.TODO(), graphName)
	if err == nil {
		graph.Remove(context.TODO())
		_, err = a.db.CreateGraphV2(context.TODO(), graphName, &options)
		return err
	}

	if !driver.IsArangoErrorWithErrorNum(err, 1924) {
		return err
	}

	_, err = a.db.CreateGraphV2(context.TODO(), graphName, &options)
	return err
}

func (a *arangoDB) Start() error {
	glog.Infof("Connected to arango database, starting monitor")
	go a.monitor()
	for _, c := range a.collections {
		go c.handler()
	}
	return nil
}

func (a *arangoDB) Stop() error {
	close(a.stop)

	return nil
}

func (a *arangoDB) GetInterface() dbclient.DB {
	return a.DB
}

func (a *arangoDB) GetArangoDBInterface() *ArangoConn {
	return a.ArangoConn
}

func (a *arangoDB) StoreMessage(msgType dbclient.CollectionType, msg []byte) error {
	if t, ok := a.collections[msgType]; ok {
		t.queue <- &queueMsg{
			msgType: msgType,
			msgData: msg,
		}
	}

	return nil
}

func (a *arangoDB) monitor() {
	for {
		select {
		case <-a.stop:
			// TODO Add clean up of connection with Arango DB
			return
		}
	}
}
