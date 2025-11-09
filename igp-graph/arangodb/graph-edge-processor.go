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

package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// IGPGraphObject represents an edge in the IGP graph (matching original lsGraphObject)
type IGPGraphObject struct {
	Key                   string      `json:"_key"`
	From                  string      `json:"_from"`
	To                    string      `json:"_to"`
	Link                  string      `json:"link"`
	ProtocolID            interface{} `json:"protocol_id"`
	DomainID              interface{} `json:"domain_id"`
	MTID                  uint16      `json:"mt_id"`
	AreaID                string      `json:"area_id"`
	Protocol              string      `json:"protocol"`
	LocalLinkID           uint32      `json:"local_link_id"`
	RemoteLinkID          uint32      `json:"remote_link_id"`
	LocalLinkIP           string      `json:"local_link_ip"`
	RemoteLinkIP          string      `json:"remote_link_ip"`
	LocalNodeASN          uint32      `json:"local_node_asn"`
	RemoteNodeASN         uint32      `json:"remote_node_asn"`
	PeerNodeSID           interface{} `json:"peer_node_sid,omitempty"`
	PeerAdjSID            interface{} `json:"peer_adj_sid,omitempty"`
	PeerSetSID            interface{} `json:"peer_set_sid,omitempty"`
	SRv6BGPPeerNodeSID    interface{} `json:"srv6_bgp_peer_node_sid,omitempty"`
	SRv6ENDXSID           interface{} `json:"srv6_endx_sid,omitempty"`
	LSAdjacencySID        interface{} `json:"ls_adjacency_sid,omitempty"`
	UnidirLinkDelay       uint32      `json:"unidir_link_delay"`
	UnidirLinkDelayMinMax []uint32    `json:"unidir_link_delay_min_max"`
	UnidirDelayVariation  uint32      `json:"unidir_delay_variation,omitempty"`
	UnidirPacketLoss      uint32      `json:"unidir_packet_loss,omitempty"`
	UnidirResidualBW      uint32      `json:"unidir_residual_bw,omitempty"`
	UnidirAvailableBW     uint32      `json:"unidir_available_bw,omitempty"`
	UnidirBWUtilization   uint32      `json:"unidir_bw_utilization,omitempty"`
	Prefix                string      `json:"prefix"`
	PrefixLen             int32       `json:"prefix_len"`
	PrefixMetric          uint32      `json:"prefix_metric"`
	PrefixAttrTLVs        interface{} `json:"prefix_attr_tlvs"`
}

// getIGPNode finds an IGP node matching the link's router information
// Handles Level-1-2 router connections by trying multiple lookup strategies
func (a *arangoDB) getIGPNode(ctx context.Context, link map[string]interface{}, local bool) (map[string]interface{}, error) {
	var routerID string
	if local {
		routerID, _ = link["igp_router_id"].(string)
	} else {
		routerID, _ = link["remote_igp_router_id"].(string)
	}

	if routerID == "" {
		return nil, fmt.Errorf("missing router ID in link data")
	}

	domainID := link["domain_id"]
	protocolID := link["protocol_id"]
	areaID, _ := link["area_id"].(string)

	// Strategy 1: Try to find node matching the link's protocol (exact match)
	node, err := a.findNodeByProtocol(ctx, routerID, domainID, protocolID, areaID)
	if err == nil {
		glog.V(8).Infof("Found exact protocol match for router %s", routerID)
		return node, nil
	}

	// Strategy 2: For Level-1 links, try to find Level-2 node (Level-1-2 routers)
	// After deduplication, Level-1-2 routers only exist as Level-2 nodes
	if proto, ok := protocolID.(float64); ok && proto == 1 {
		node, err := a.findNodeByProtocol(ctx, routerID, domainID, float64(2), areaID)
		if err == nil {
			glog.V(6).Infof("Found Level-2 node for Level-1 link, router %s (Level-1-2 router)", routerID)
			return node, nil
		}
	}

	// Strategy 3: Find any node with matching router_id and domain_id (protocol-agnostic)
	node, err = a.findNodeAnyProtocol(ctx, routerID, domainID, areaID)
	if err == nil {
		glog.V(8).Infof("Found protocol-agnostic match for router %s", routerID)
		return node, nil
	}

	return nil, fmt.Errorf("no IGP node found for router %s in domain %v", routerID, domainID)
}

// findNodeByProtocol finds a node with specific protocol
func (a *arangoDB) findNodeByProtocol(ctx context.Context, routerID string, domainID, protocolID interface{}, areaID string) (map[string]interface{}, error) {
	query := fmt.Sprintf("FOR d IN %s", a.config.IGPNode)
	query += " FILTER d.igp_router_id == @routerId"
	query += " FILTER d.domain_id == @domainId"
	query += " FILTER d.protocol_id == @protocolId"

	bindVars := map[string]interface{}{
		"routerId":   routerID,
		"domainId":   domainID,
		"protocolId": protocolID,
	}

	// For OSPF (protocol 3=OSPFv2, 6=OSPFv3), include area_id in query
	if proto, ok := protocolID.(float64); ok && (proto == 3 || proto == 6) {
		query += " FILTER d.area_id == @areaId"
		bindVars["areaId"] = areaID
	}

	query += " RETURN d"

	return a.executeNodeQuery(ctx, query, bindVars)
}

// findNodeAnyProtocol finds a node regardless of protocol (fallback)
func (a *arangoDB) findNodeAnyProtocol(ctx context.Context, routerID string, domainID interface{}, areaID string) (map[string]interface{}, error) {
	query := fmt.Sprintf("FOR d IN %s", a.config.IGPNode)
	query += " FILTER d.igp_router_id == @routerId"
	query += " FILTER d.domain_id == @domainId"
	query += " RETURN d"

	bindVars := map[string]interface{}{
		"routerId": routerID,
		"domainId": domainID,
	}

	return a.executeNodeQuery(ctx, query, bindVars)
}

// executeNodeQuery executes a node lookup query and returns the result
func (a *arangoDB) executeNodeQuery(ctx context.Context, query string, bindVars map[string]interface{}) (map[string]interface{}, error) {
	glog.V(9).Infof("Node query: %s, vars: %+v", query, bindVars)

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute node query: %w", err)
	}
	defer cursor.Close()

	var node map[string]interface{}
	count := 0
	for {
		_, err := cursor.ReadDocument(ctx, &node)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return nil, fmt.Errorf("error reading node document: %w", err)
		}
		count++
		if count > 1 {
			// If multiple nodes found, prefer Level-2 (protocol_id = 2)
			var currentNode map[string]interface{}
			currentNode = node
			if proto, ok := currentNode["protocol_id"].(float64); ok && proto == 2 {
				// Keep this Level-2 node and stop
				break
			}
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("query returned 0 results")
	}

	return node, nil
}

// createIGPv4EdgeObject creates an IPv4 IGP graph edge (mirrors original createv4EdgeObject)
func (a *arangoDB) createIGPv4EdgeObject(ctx context.Context, link map[string]interface{}, localNode, remoteNode map[string]interface{}) error {
	key, _ := link["_key"].(string)

	// Extract MTID
	var mtid uint16 = 0
	if mtidTLV, exists := link["mt_id_tlv"]; exists {
		if mtidArray, ok := mtidTLV.([]interface{}); ok && len(mtidArray) > 0 {
			if mtObj, ok := mtidArray[0].(map[string]interface{}); ok {
				if mt, ok := mtObj["mt_id"].(float64); ok {
					mtid = uint16(mt)
				}
			}
		} else if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			if mt, ok := mtidObj["mt_id"].(float64); ok {
				mtid = uint16(mt)
			}
		}
	}

	// Helper function for safe uint32 conversion
	getUint32 := func(v interface{}) uint32 {
		switch val := v.(type) {
		case float64:
			return uint32(val)
		case uint32:
			return val
		case int:
			return uint32(val)
		default:
			return 0
		}
	}

	// Create edge object matching original lsGraphObject structure
	edge := IGPGraphObject{
		Key:                   key,
		From:                  fmt.Sprintf("%s", localNode["_id"]),
		To:                    fmt.Sprintf("%s", remoteNode["_id"]),
		Link:                  key,
		ProtocolID:            link["protocol_id"],
		DomainID:              link["domain_id"],
		MTID:                  mtid,
		AreaID:                getString(link, "area_id"),
		Protocol:              getString(remoteNode, "protocol"),
		LocalLinkID:           getUint32(link["local_link_id"]),
		RemoteLinkID:          getUint32(link["remote_link_id"]),
		LocalLinkIP:           getString(link, "local_link_ip"),
		RemoteLinkIP:          getString(link, "remote_link_ip"),
		LocalNodeASN:          getUint32(link["local_node_asn"]),
		RemoteNodeASN:         getUint32(link["remote_node_asn"]),
		PeerNodeSID:           link["peer_node_sid"],
		PeerAdjSID:            link["peer_adj_sid"],
		PeerSetSID:            link["peer_set_sid"],
		SRv6BGPPeerNodeSID:    link["srv6_bgp_peer_node_sid"],
		SRv6ENDXSID:           link["srv6_endx_sid"],
		LSAdjacencySID:        link["ls_adjacency_sid"],
		UnidirLinkDelay:       getUint32(link["unidir_link_delay"]),
		UnidirLinkDelayMinMax: getUint32Array(link["unidir_link_delay_min_max"]),
		UnidirDelayVariation:  getUint32(link["unidir_delay_variation"]),
		UnidirPacketLoss:      getUint32(link["unidir_packet_loss"]),
		UnidirResidualBW:      getUint32(link["unidir_residual_bw"]),
		UnidirAvailableBW:     getUint32(link["unidir_available_bw"]),
		UnidirBWUtilization:   getUint32(link["unidir_bw_utilization"]),
		Prefix:                "",
		PrefixLen:             0,
		PrefixMetric:          0,
		PrefixAttrTLVs:        nil,
	}

	// Get IPv4 graph collection
	igpv4Collection, err := a.db.Collection(ctx, a.config.IGPv4Graph)
	if err != nil {
		return fmt.Errorf("failed to get IGPv4 collection: %w", err)
	}

	// Create or update the edge document
	if _, err := igpv4Collection.CreateDocument(ctx, &edge); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create IGPv4 edge: %w", err)
		}
		// Document exists, update it
		if _, err := igpv4Collection.UpdateDocument(ctx, edge.Key, &edge); err != nil {
			return fmt.Errorf("failed to update IGPv4 edge: %w", err)
		}
	}

	glog.V(7).Infof("Created/updated IGPv4 edge: %s", key)
	return nil
}

// createIGPv6EdgeObject creates an IPv6 IGP graph edge (mirrors original createv6EdgeObject)
func (a *arangoDB) createIGPv6EdgeObject(ctx context.Context, link map[string]interface{}, localNode, remoteNode map[string]interface{}) error {
	key, _ := link["_key"].(string)

	// Extract MTID
	var mtid uint16 = 0
	if mtidTLV, exists := link["mt_id_tlv"]; exists {
		if mtidArray, ok := mtidTLV.([]interface{}); ok && len(mtidArray) > 0 {
			if mtObj, ok := mtidArray[0].(map[string]interface{}); ok {
				if mt, ok := mtObj["mt_id"].(float64); ok {
					mtid = uint16(mt)
				}
			}
		} else if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			if mt, ok := mtidObj["mt_id"].(float64); ok {
				mtid = uint16(mt)
			}
		}
	}

	// Helper function for safe uint32 conversion
	getUint32 := func(v interface{}) uint32 {
		switch val := v.(type) {
		case float64:
			return uint32(val)
		case uint32:
			return val
		case int:
			return uint32(val)
		default:
			return 0
		}
	}

	// Create edge object matching original lsGraphObject structure
	edge := IGPGraphObject{
		Key:                   key,
		From:                  fmt.Sprintf("%s", localNode["_id"]),
		To:                    fmt.Sprintf("%s", remoteNode["_id"]),
		Link:                  key,
		ProtocolID:            link["protocol_id"],
		DomainID:              link["domain_id"],
		MTID:                  mtid,
		AreaID:                getString(link, "area_id"),
		Protocol:              getString(remoteNode, "protocol"),
		LocalLinkID:           getUint32(link["local_link_id"]),
		RemoteLinkID:          getUint32(link["remote_link_id"]),
		LocalLinkIP:           getString(link, "local_link_ip"),
		RemoteLinkIP:          getString(link, "remote_link_ip"),
		LocalNodeASN:          getUint32(link["local_node_asn"]),
		RemoteNodeASN:         getUint32(link["remote_node_asn"]),
		PeerNodeSID:           link["peer_node_sid"],
		PeerAdjSID:            link["peer_adj_sid"],
		PeerSetSID:            link["peer_set_sid"],
		SRv6BGPPeerNodeSID:    link["srv6_bgp_peer_node_sid"],
		SRv6ENDXSID:           link["srv6_endx_sid"],
		LSAdjacencySID:        link["ls_adjacency_sid"],
		UnidirLinkDelay:       getUint32(link["unidir_link_delay"]),
		UnidirLinkDelayMinMax: getUint32Array(link["unidir_link_delay_min_max"]),
		UnidirDelayVariation:  getUint32(link["unidir_delay_variation"]),
		UnidirPacketLoss:      getUint32(link["unidir_packet_loss"]),
		UnidirResidualBW:      getUint32(link["unidir_residual_bw"]),
		UnidirAvailableBW:     getUint32(link["unidir_available_bw"]),
		UnidirBWUtilization:   getUint32(link["unidir_bw_utilization"]),
		Prefix:                "",
		PrefixLen:             0,
		PrefixMetric:          0,
		PrefixAttrTLVs:        nil,
	}

	// Get IPv6 graph collection
	igpv6Collection, err := a.db.Collection(ctx, a.config.IGPv6Graph)
	if err != nil {
		return fmt.Errorf("failed to get IGPv6 collection: %w", err)
	}

	// Create or update the edge document
	if _, err := igpv6Collection.CreateDocument(ctx, &edge); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create IGPv6 edge: %w", err)
		}
		// Document exists, update it
		if _, err := igpv6Collection.UpdateDocument(ctx, edge.Key, &edge); err != nil {
			return fmt.Errorf("failed to update IGPv6 edge: %w", err)
		}
	}

	glog.V(7).Infof("Created/updated IGPv6 edge: %s", key)
	return nil
}

// Helper function to safely convert to uint32 array
func getUint32Array(v interface{}) []uint32 {
	if arr, ok := v.([]interface{}); ok {
		result := make([]uint32, len(arr))
		for i, item := range arr {
			switch val := item.(type) {
			case float64:
				result[i] = uint32(val)
			case uint32:
				result[i] = val
			case int:
				result[i] = uint32(val)
			default:
				result[i] = 0
			}
		}
		return result
	}
	return nil
}
