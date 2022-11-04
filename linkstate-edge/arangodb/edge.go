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
	"strconv"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/cisco-open/jalapeno/linkstate-edge/kafkanotifier"
	notifier "github.com/cisco-open/jalapeno/topology/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/message"
)

const LSNodeEdgeCollection = "ls_node_edge"

func (a *arangoDB) lsLinkHandler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	// Check if Collection encoded in ID exists
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.edge.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.edge.Name(), c)
	}
	glog.V(5).Infof("Processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)
	var o message.LSLink
	_, err := a.edge.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		// In case of a ls_link removal notification, reading it will return Not Found error
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
		}
		// If operation matches to "del" then it is confirmed delete operation, otherwise return error
		if obj.Action != "del" {
			return fmt.Errorf("document %s not found but Action is not \"del\", possible stale event", obj.Key)
		}
		return a.processEdgeRemoval(ctx, obj.Key, obj.Action)
		//return nil
	}
	switch obj.Action {
	case "add":
		fallthrough
	case "update":
		if err := a.processEdge(ctx, obj.Key, &o); err != nil {
			return fmt.Errorf("failed to process action %s for edge %s with error: %+v", obj.Action, obj.Key, err)
		}
	}
	glog.V(5).Infof("Complete processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)

	//TODO Write back into ls_node_edge_events topic
	glog.V(5).Infof("Writing into events topic")
	a.notifier.EventNotification((*kafkanotifier.EventMessage)(obj))

	return nil
}

func (a *arangoDB) lsNodeHandler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	// Check if Collection encoded in ID exists
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.vertex.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.vertex.Name(), c)
	}
	glog.V(5).Infof("Processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)
	var o message.LSNode
	_, err := a.vertex.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		// In case of a ls_node removal notification, reading it will return Not Found error
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
		}
		// If operation matches to "del" then it is confirmed delete operation, otherwise return error
		if obj.Action != "del" {
			return fmt.Errorf("document %s not found but Action is not \"del\", possible stale event", obj.Key)
		}
		// return a.processVertexRemoval(ctx, obj.Key, obj.Action)
		return nil
	}
	switch obj.Action {
	case "add":
		fallthrough
	case "update":
		if err := a.processVertex(ctx, obj.Key, &o); err != nil {
			return fmt.Errorf("failed to process action %s for vertex %s with error: %+v", obj.Action, obj.Key, err)
		}
	}
	glog.V(5).Infof("Complete processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)

	return nil
}

type lsNodeEdgeObject struct {
	Key           string       `json:"_key"`
	From          string       `json:"_from"`
	To            string       `json:"_to"`
	Link          string       `json:"link"`
	ProtocolID    base.ProtoID `json:"protocol_id"`
	DomainID      int64        `json:"domain_id"`
	MTID          uint16       `json:"mt_id"`
	AreaID        string       `json:"area_id"`
	LocalLinkID   uint32       `json:"local_link_id"`
	RemoteLinkID  uint32       `json:"remote_link_id"`
	LocalLinkIP   string       `json:"local_link_ip"`
	RemoteLinkIP  string       `json:"remote_link_ip"`
	LocalNodeASN  uint32       `json:"local_node_asn"`
	RemoteNodeASN uint32       `json:"remote_node_asn"`
}

// processEdge processes a single ls_link connection which is a unidirectional edge between two nodes (vertices).
func (a *arangoDB) processEdge(ctx context.Context, key string, l *message.LSLink) error {
	if l.ProtocolID == base.BGP {
		return nil
	}
	glog.V(9).Infof("processEdge processing lslink: %s", l.ID)
	ln, err := a.getNode(ctx, l, true)
	if err != nil {
		glog.Errorf("processEdge failed to get local lsnode %s for link: %s with error: %+v", l.IGPRouterID, l.ID, err)
		return err
	}

	rn, err := a.getNode(ctx, l, false)
	if err != nil {
		glog.Errorf("processEdge failed to get remote lsnode %s for link: %s with error: %+v", l.RemoteIGPRouterID, l.ID, err)
		return err
	}
	// glog.V(6).Infof("Local node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
	// 	ln.ProtocolID, ln.DomainID, ln.IGPRouterID)
	// glog.V(6).Infof("Remote node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
	// 	rn.ProtocolID, rn.DomainID, rn.IGPRouterID)
	if err := a.createEdgeObject(ctx, l, ln, rn); err != nil {
		glog.Errorf("processEdge failed to create Edge object with error: %+v", err)
		glog.Errorf("Local node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
			ln.ProtocolID, ln.DomainID, ln.IGPRouterID)
		glog.Errorf("Remote node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
			rn.ProtocolID, rn.DomainID, rn.IGPRouterID)
		return err
	}
	glog.V(9).Infof("processEdge completed processing lslink: %s for ls nodes: %s - %s", l.ID, ln.ID, rn.ID)

	return nil
}

func (a *arangoDB) processVertex(ctx context.Context, key string, ln *message.LSNode) error {
	if ln.ProtocolID == 7 {
		return nil
	}
	// Check if there is an edge with matching to ls_node's e.IGPRouterID, e.AreaID, e.DomainID and e.ProtocolID
	query := "FOR d IN " + a.edge.Name() +
		" filter d.igp_router_id == " + "\"" + ln.IGPRouterID + "\"" +
		" OR d.remote_igp_router_id == " + "\"" + ln.IGPRouterID + "\"" +
		" filter d.domain_id == " + strconv.Itoa(int(ln.DomainID)) +
		" filter d.protocol_id == " + strconv.Itoa(int(ln.ProtocolID))
	// If OSPFv2 or OSPFv3, then query must include AreaID
	if ln.ProtocolID == base.OSPFv2 || ln.ProtocolID == base.OSPFv3 {
		query += " filter d.area_id == " + "\"" + ln.AreaID + "\""
	}
	query += " return d"
	lcursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer lcursor.Close()
	var l message.LSLink
	// Processing each ls_link
	i := 0
	for ; ; i++ {
		_, err := lcursor.ReadDocument(ctx, &l)
		if err != nil {
			if !driver.IsNoMoreDocuments(err) {
				return err
			}
			break
		}
		glog.V(9).Infof("processVertex processing lsnode: %s link: %s", ln.ID, l.ID)
		rn, err := a.getNode(ctx, &l, false)
		if err != nil {
			glog.Errorf("processVertex failed to get remote ls node for lsnode: %s link: %s remote node: %s", ln.ID, l.ID, l.RemoteIGPRouterID)
			continue
		}
		// glog.V(6).Infof("Local node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
		// 	ln.ProtocolID, ln.DomainID, ln.IGPRouterID)
		// glog.V(6).Infof("Remote node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
		// 	rn.ProtocolID, rn.DomainID, rn.IGPRouterID)

		if err := a.createEdgeObject(ctx, &l, ln, rn); err != nil {
			glog.Errorf("proicessVertex failed to create Edge object with error: %+v", err)
			glog.Errorf("Local node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
				ln.ProtocolID, ln.DomainID, ln.IGPRouterID)
			glog.Errorf("Remote node -> Protocol: %+v Domain ID: %+v IGP Router ID: %+v",
				rn.ProtocolID, rn.DomainID, rn.IGPRouterID)
			continue
		}
		glog.V(9).Infof("processVertex completed processing lsnode: %s link: %s remote node: %s", ln.ID, l.ID, rn.ID)
	}

	return nil
}

// processEdgeRemoval removes a record from Node's graph collection
// since the key matches in both collections (LS Links and Nodes' Graph) deleting the record directly.
func (a *arangoDB) processEdgeRemoval(ctx context.Context, key string, action string) error {
	if _, err := a.graph.RemoveDocument(ctx, key); err != nil {
		if !driver.IsNotFound(err) {
			return err
		}
		return nil
	}

	return nil
}

// processEdgeRemoval removes all documents where removed Vertix (ls_node) is referenced in "_to" or "_from"
func (a *arangoDB) processVertexRemoval(ctx context.Context, key string, action string) error {
	query := "FOR d IN " + a.graph.Name() +
		" filter d._to == " + "\"" + key + "\"" + " OR" + " d._from == " + "\"" + key + "\"" +
		" return d"
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()

	for {
		var p lsNodeEdgeObject
		_, err := cursor.ReadDocument(ctx, &p)
		if err != nil {
			if !driver.IsNoMoreDocuments(err) {
				return err
			}
			break
		}
		glog.V(6).Infof("Removing from %s object %s", a.graph.Name(), p.Key)
		if _, err = a.graph.RemoveDocument(ctx, p.Key); err != nil {
			if !driver.IsNotFound(err) {
				return err
			}
			return nil
		}
	}
	return nil
}

func (a *arangoDB) getNode(ctx context.Context, e *message.LSLink, local bool) (*message.LSNode, error) {
	// Need to find ls_node object matching ls_link's IGP Router ID
	query := "FOR d IN " + a.vertex.Name()
	if local {
		query += " filter d.igp_router_id == " + "\"" + e.IGPRouterID + "\""
	} else {
		query += " filter d.igp_router_id == " + "\"" + e.RemoteIGPRouterID + "\""
	}
	query += " filter d.domain_id == " + strconv.Itoa(int(e.DomainID)) +
		" filter d.protocol_id == " + strconv.Itoa(int(e.ProtocolID))
	// If OSPFv2 or OSPFv3, then query must include AreaID
	if e.ProtocolID == base.OSPFv2 || e.ProtocolID == base.OSPFv3 {
		query += " filter d.area_id == " + "\"" + e.AreaID + "\""
	}
	query += " return d"
	lcursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}
	defer lcursor.Close()
	var ln message.LSNode
	i := 0
	for ; ; i++ {
		_, err := lcursor.ReadDocument(ctx, &ln)
		if err != nil {
			if !driver.IsNoMoreDocuments(err) {
				return nil, err
			}
			break
		}
	}
	if i == 0 {
		return nil, fmt.Errorf("query %s returned 0 results", query)
	}
	if i > 1 {
		return nil, fmt.Errorf("query %s returned more than 1 result", query)
	}

	return &ln, nil
}

func (a *arangoDB) createEdgeObject(ctx context.Context, l *message.LSLink, ln, rn *message.LSNode) error {
	mtid := 0
	if l.MTID != nil {
		mtid = int(l.MTID.MTID)
	}
	ne := lsNodeEdgeObject{
		Key:           l.Key,
		From:          ln.ID,
		To:            rn.ID,
		Link:          l.Key,
		ProtocolID:    l.ProtocolID,
		DomainID:      l.DomainID,
		MTID:          uint16(mtid),
		AreaID:        l.AreaID,
		LocalLinkID:   l.LocalLinkID,
		RemoteLinkID:  l.RemoteLinkID,
		LocalLinkIP:   l.LocalLinkIP,
		RemoteLinkIP:  l.RemoteLinkIP,
		LocalNodeASN:  l.LocalNodeASN,
		RemoteNodeASN: l.RemoteNodeASN,
	}
	if _, err := a.graph.CreateDocument(ctx, &ne); err != nil {
		if !driver.IsConflict(err) {
			return err
		}
		// The document already exists, updating it with the latest info
		if _, err := a.graph.UpdateDocument(ctx, ne.Key, &ne); err != nil {
			return err
		}
	}

	return nil
}
