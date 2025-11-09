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
	"net"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// IBGPSubnetProcessor handles iBGP-only nodes that should attach to subnet prefixes instead of RR
type IBGPSubnetProcessor struct {
	db *arangoDB
}

// NewIBGPSubnetProcessor creates a new iBGP subnet processor
func NewIBGPSubnetProcessor(db *arangoDB) *IBGPSubnetProcessor {
	return &IBGPSubnetProcessor{
		db: db,
	}
}

// ProcessIBGPSubnetAttachment handles iBGP-only nodes that should connect to subnet prefixes
func (isp *IBGPSubnetProcessor) ProcessIBGPSubnetAttachment(ctx context.Context) error {
	glog.Info("Processing iBGP-only node subnet attachments...")

	// Find iBGP-only BGP nodes (nodes that don't have corresponding IGP entries)
	ibgpOnlyNodes, err := isp.findIBGPOnlyNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to find iBGP-only nodes: %w", err)
	}

	processedCount := 0
	for _, bgpNode := range ibgpOnlyNodes {
		nodeID := getString(bgpNode, "_id")
		routerID := getString(bgpNode, "router_id")
		localIP := getString(bgpNode, "local_ip")
		remoteIP := getString(bgpNode, "remote_ip")

		glog.V(7).Infof("Processing iBGP-only node: %s (local_ip: %s, remote_ip: %s)", routerID, localIP, remoteIP)

		// Try to find subnet attachment for both local_ip and remote_ip
		attached := false

		// First try remote_ip (the Cilium node's IP)
		if remoteIP != "" {
			if err := isp.attachNodeToSubnet(ctx, nodeID, remoteIP, "remote"); err != nil {
				glog.V(8).Infof("No subnet attachment found for remote_ip %s: %v", remoteIP, err)
			} else {
				attached = true
			}
		}

		// Then try local_ip if remote_ip didn't work
		if !attached && localIP != "" {
			if err := isp.attachNodeToSubnet(ctx, nodeID, localIP, "local"); err != nil {
				glog.V(8).Infof("No subnet attachment found for local_ip %s: %v", localIP, err)
			} else {
				attached = true
			}
		}

		if attached {
			processedCount++
		} else {
			glog.V(7).Infof("No subnet attachment found for iBGP-only node %s", routerID)
		}
	}

	glog.Infof("Processed %d iBGP-only node subnet attachments", processedCount)
	return nil
}

// findIBGPOnlyNodes finds BGP nodes that are iBGP but don't have corresponding IGP entries
func (isp *IBGPSubnetProcessor) findIBGPOnlyNodes(ctx context.Context) ([]map[string]interface{}, error) {
	// Find BGP nodes that are iBGP (same ASN) but don't exist in IGP
	query := fmt.Sprintf(`
		FOR bgp IN %s
		FILTER bgp.asn != null
		// Check if this is an iBGP session by looking at peer data
		FOR peer IN peer
		FILTER peer.remote_bgp_id == bgp.router_id
		FILTER peer.remote_asn == peer.local_asn  // iBGP session
		// Ensure this node doesn't exist in IGP
		LET igp_exists = (
			FOR igp IN %s
			FILTER igp.router_id == bgp.router_id OR igp.bgp_router_id == bgp.router_id
			LIMIT 1
			RETURN true
		)
		FILTER LENGTH(igp_exists) == 0  // No IGP entry found
		RETURN MERGE(bgp, {
			local_ip: peer.local_ip,
			remote_ip: peer.remote_ip,
			local_asn: peer.local_asn,
			remote_asn: peer.remote_asn
		})
	`, isp.db.config.BGPNode, isp.db.config.IGPNode)

	cursor, err := isp.db.db.Query(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query iBGP-only nodes: %w", err)
	}
	defer cursor.Close()

	var ibgpOnlyNodes []map[string]interface{}
	for cursor.HasMore() {
		var node map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &node); err != nil {
			continue
		}
		ibgpOnlyNodes = append(ibgpOnlyNodes, node)
	}

	glog.V(6).Infof("Found %d iBGP-only nodes", len(ibgpOnlyNodes))
	return ibgpOnlyNodes, nil
}

// attachNodeToSubnet finds and attaches a BGP node to the most specific subnet containing the IP
func (isp *IBGPSubnetProcessor) attachNodeToSubnet(ctx context.Context, nodeID, ipAddress, ipType string) error {
	// Parse the IP address
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddress)
	}

	isIPv4 := ip.To4() != nil

	// Find all prefixes that could contain this IP
	candidates, err := isp.findCandidateSubnets(ctx, ipAddress, isIPv4)
	if err != nil {
		return fmt.Errorf("failed to find candidate subnets: %w", err)
	}

	// Find the most specific (longest prefix) match
	bestMatch, err := isp.findBestSubnetMatch(ip, candidates, isIPv4)
	if err != nil {
		return fmt.Errorf("failed to find best subnet match: %w", err)
	}

	if bestMatch == nil {
		return fmt.Errorf("no subnet found containing IP %s", ipAddress)
	}

	// Create attachment to the subnet
	if err := isp.createSubnetAttachment(ctx, nodeID, bestMatch, ipAddress, ipType, isIPv4); err != nil {
		return fmt.Errorf("failed to create subnet attachment: %w", err)
	}

	prefix := getString(bestMatch, "prefix")
	prefixLen := getUint32FromInterface(bestMatch["prefix_len"])
	glog.V(7).Infof("Attached iBGP-only node to subnet %s/%d via %s IP %s", prefix, prefixLen, ipType, ipAddress)

	return nil
}

// findCandidateSubnets finds all prefixes that could potentially contain the given IP
func (isp *IBGPSubnetProcessor) findCandidateSubnets(ctx context.Context, ipAddress string, isIPv4 bool) ([]map[string]interface{}, error) {
	// Build query based on IP version - look in both ls_prefix and bgp_prefix collections
	var mtidFilter string
	var prefixCollections []string

	if isIPv4 {
		mtidFilter = "FILTER ls.mt_id_tlv == null OR ls.mt_id_tlv.mt_id == 0"
		prefixCollections = []string{"ls_prefix", isp.db.config.BGPPrefixV4}
	} else {
		mtidFilter = "FILTER ls.mt_id_tlv != null AND ls.mt_id_tlv.mt_id == 2"
		prefixCollections = []string{"ls_prefix", isp.db.config.BGPPrefixV6}
	}

	var allCandidates []map[string]interface{}

	// Search in ls_prefix (IGP prefixes)
	lsPrefixQuery := fmt.Sprintf(`
		FOR ls IN ls_prefix
		%s
		FILTER ls.prefix_len < %d  // Exclude host routes
		RETURN {
			prefix: ls.prefix,
			prefix_len: ls.prefix_len,
			source: "igp",
			_key: ls._key,
			collection: "ls_prefix"
		}
	`, mtidFilter, func() int {
		if isIPv4 {
			return 32
		}
		return 128
	}())

	cursor, err := isp.db.db.Query(ctx, lsPrefixQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query ls_prefix candidates: %w", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var candidate map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &candidate); err != nil {
			continue
		}
		allCandidates = append(allCandidates, candidate)
	}

	// Search in bgp_prefix collection
	bgpPrefixQuery := fmt.Sprintf(`
		FOR bgp IN %s
		FILTER bgp.prefix_len < %d  // Exclude host routes
		RETURN {
			prefix: bgp.prefix,
			prefix_len: bgp.prefix_len,
			source: "bgp",
			_key: bgp._key,
			collection: @bgpCollection
		}
	`, prefixCollections[1], func() int {
		if isIPv4 {
			return 32
		}
		return 128
	}())

	bindVars := map[string]interface{}{
		"bgpCollection": prefixCollections[1],
	}

	cursor2, err := isp.db.db.Query(ctx, bgpPrefixQuery, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query BGP prefix candidates: %w", err)
	}
	defer cursor2.Close()

	for cursor2.HasMore() {
		var candidate map[string]interface{}
		if _, err := cursor2.ReadDocument(ctx, &candidate); err != nil {
			continue
		}
		allCandidates = append(allCandidates, candidate)
	}

	glog.V(8).Infof("Found %d candidate subnets for IP %s", len(allCandidates), ipAddress)
	return allCandidates, nil
}

// findBestSubnetMatch finds the most specific subnet that contains the given IP
func (isp *IBGPSubnetProcessor) findBestSubnetMatch(targetIP net.IP, candidates []map[string]interface{}, isIPv4 bool) (map[string]interface{}, error) {
	var bestMatch map[string]interface{}
	var longestPrefixLen uint32 = 0

	for _, candidate := range candidates {
		prefix := getString(candidate, "prefix")
		prefixLen := getUint32FromInterface(candidate["prefix_len"])

		// Parse the candidate prefix
		_, candidateNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", prefix, prefixLen))
		if err != nil {
			glog.V(9).Infof("Failed to parse candidate prefix %s/%d: %v", prefix, prefixLen, err)
			continue
		}

		// Check if the target IP is within this subnet
		if candidateNet.Contains(targetIP) {
			// This subnet contains the IP, check if it's more specific than current best
			if prefixLen > longestPrefixLen {
				longestPrefixLen = prefixLen
				bestMatch = candidate
				glog.V(9).Infof("New best match: %s/%d for IP %s", prefix, prefixLen, targetIP.String())
			}
		}
	}

	return bestMatch, nil
}

// createSubnetAttachment creates edges between the BGP node and the subnet vertex
func (isp *IBGPSubnetProcessor) createSubnetAttachment(ctx context.Context, nodeID string, subnet map[string]interface{}, ipAddress, ipType string, isIPv4 bool) error {
	prefix := getString(subnet, "prefix")
	prefixLen := getUint32FromInterface(subnet["prefix_len"])
	subnetKey := getString(subnet, "_key")
	collection := getString(subnet, "collection")

	// Construct the subnet vertex ID
	subnetVertexID := fmt.Sprintf("%s/%s", collection, subnetKey)

	// Extract node key from node ID
	nodeKey := nodeID
	if idx := strings.LastIndex(nodeID, "/"); idx != -1 {
		nodeKey = nodeID[idx+1:]
	}

	// Create bidirectional edges
	edges := []*IPGraphObject{
		{
			Key:       fmt.Sprintf("%s_subnet_%s", nodeKey, subnetKey),
			From:      nodeID,
			To:        subnetVertexID,
			Protocol:  fmt.Sprintf("iBGP_subnet_%s", ipType),
			Link:      fmt.Sprintf("subnet_%s", subnetKey),
			Prefix:    prefix,
			PrefixLen: int32(prefixLen),
		},
		{
			Key:       fmt.Sprintf("subnet_%s_%s", subnetKey, nodeKey),
			From:      subnetVertexID,
			To:        nodeID,
			Protocol:  fmt.Sprintf("iBGP_subnet_%s", ipType),
			Link:      fmt.Sprintf("subnet_%s", subnetKey),
			Prefix:    prefix,
			PrefixLen: int32(prefixLen),
		},
	}

	// Determine target graph collection
	var targetCollection driver.Collection
	if isIPv4 {
		targetCollection = isp.db.ipv4Graph
	} else {
		targetCollection = isp.db.ipv6Graph
	}

	// Create both edges
	for _, edge := range edges {
		if _, err := targetCollection.CreateDocument(ctx, edge); err != nil {
			if !driver.IsConflict(err) {
				return fmt.Errorf("failed to create subnet edge %s: %w", edge.Key, err)
			}
			// Update existing edge
			if _, err := targetCollection.UpdateDocument(ctx, edge.Key, edge); err != nil {
				return fmt.Errorf("failed to update subnet edge %s: %w", edge.Key, err)
			}
		}
	}

	return nil
}
