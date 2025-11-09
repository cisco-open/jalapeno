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
	"github.com/sbezverk/gobmp/pkg/srv6"
)

// loadInitialSRv6SIDs loads and processes all existing SRv6 SIDs during initialization
func (a *arangoDB) loadInitialSRv6SIDs(ctx context.Context) error {
	query := `FOR doc IN @@collection RETURN doc`
	bindVars := map[string]interface{}{
		"@collection": a.config.LSSRv6SID,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to query ls_srv6_sid collection: %w", err)
	}
	defer cursor.Close()

	count := 0
	for {
		var srv6Data map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &srv6Data)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return fmt.Errorf("failed to read SRv6 SID document: %w", err)
		}

		// Process SRv6 SID (simplified for now)
		if err := a.processInitialSRv6SID(ctx, srv6Data); err != nil {
			glog.Warningf("Failed to process initial SRv6 SID %v: %v", srv6Data["_key"], err)
			continue
		}
		count++

		if count%1000 == 0 {
			glog.V(3).Infof("Loaded %d SRv6 SIDs...", count)
		}
	}

	glog.Infof("Loaded %d initial SRv6 SIDs", count)
	return nil
}

// processInitialSRv6SID processes a SRv6 SID during initial loading
func (a *arangoDB) processInitialSRv6SID(ctx context.Context, srv6Data map[string]interface{}) error {
	// Extract router ID for finding the corresponding IGP node
	routerID, ok := srv6Data["igp_router_id"].(string)
	if !ok || routerID == "" {
		return fmt.Errorf("missing or invalid igp_router_id in SRv6 SID data")
	}

	domainID, ok := srv6Data["domain_id"]
	if !ok {
		return fmt.Errorf("missing domain_id in SRv6 SID data")
	}

	// Create SID object from raw data
	sid := SID{
		SRv6SID: getString(srv6Data, "srv6_sid"),
	}

	// Process endpoint behavior if present
	if epBehavior, exists := srv6Data["srv6_endpoint_behavior"]; exists {
		if epMap, ok := epBehavior.(map[string]interface{}); ok {
			endpointBehavior := extractEndpointBehavior(epMap)
			if endpointBehavior != nil {
				sid.SRv6EndpointBehavior = endpointBehavior
				glog.V(8).Infof("Added endpoint behavior to SID %s: algo=%d, behavior=%d",
					sid.SRv6SID, endpointBehavior.Algorithm, endpointBehavior.EndpointBehavior)
			}
		}
	}

	// Process SID structure if present
	if sidStruct, exists := srv6Data["srv6_sid_structure"]; exists {
		if structMap, ok := sidStruct.(map[string]interface{}); ok {
			sidStructure := extractSIDStructure(structMap)
			if sidStructure != nil {
				sid.SRv6SIDStructure = sidStructure
				glog.V(8).Infof("Added SID structure to SID %s: locator_block=%d, locator_node=%d, function=%d, argument=%d",
					sid.SRv6SID, sidStructure.LBLength, sidStructure.LNLength,
					sidStructure.FunLength, sidStructure.ArgLength)
			}
		}
	}

	// Find and update the corresponding IGP node
	return a.addSRv6SIDToIGPNode(ctx, routerID, domainID, sid)
}

// addSRv6SIDToIGPNode finds the IGP node and adds the SRv6 SID to it
func (a *arangoDB) addSRv6SIDToIGPNode(ctx context.Context, routerID string, domainID interface{}, sid SID) error {
	// Query to find the IGP node
	query := `
		FOR node IN @@collection
		FILTER node.igp_router_id == @routerId
		FILTER node.domain_id == @domainId
		RETURN node`

	bindVars := map[string]interface{}{
		"@collection": a.config.IGPNode,
		"routerId":    routerID,
		"domainId":    domainID,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to execute IGP node query: %w", err)
	}
	defer cursor.Close()

	var igpNode IGPNode
	meta, err := cursor.ReadDocument(ctx, &igpNode)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			glog.V(6).Infof("No IGP node found for router %s, domain %v", routerID, domainID)
			return nil
		}
		return fmt.Errorf("error reading IGP node document: %w", err)
	}

	glog.V(6).Infof("Processing IGP node %s with SRv6 SID %s", meta.Key, sid.SRv6SID)

	// Check if SID already exists by comparing SRv6SID field
	for _, existingSID := range igpNode.SIDS {
		if existingSID.SRv6SID == sid.SRv6SID {
			glog.V(6).Infof("SRv6 SID %s already exists in IGP node document", sid.SRv6SID)
			return nil
		}
	}

	// Append new SID and update document
	igpNode.SIDS = append(igpNode.SIDS, sid)
	if _, err := a.igpNode.UpdateDocument(ctx, meta.Key, igpNode); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to update IGP node: %w", err)
		}
		glog.V(6).Infof("Conflict while updating IGP node %s", meta.Key)
	}

	glog.V(7).Infof("Added SRv6 SID %s to IGP node %s", sid.SRv6SID, meta.Key)
	return nil
}

// processSRv6SIDUpdate handles real-time SRv6 SID updates
func (a *arangoDB) processSRv6SIDUpdate(ctx context.Context, action, key string, srv6Data map[string]interface{}) error {
	switch action {
	case "del":
		return a.removeSRv6SIDFromIGPNode(ctx, key, srv6Data)
	case "add", "update":
		return a.processInitialSRv6SID(ctx, srv6Data)
	default:
		glog.V(5).Infof("Unknown SRv6 SID action: %s for key: %s", action, key)
		return nil
	}
}

// removeSRv6SIDFromIGPNode removes a SRv6 SID from the corresponding IGP node
func (a *arangoDB) removeSRv6SIDFromIGPNode(ctx context.Context, key string, srv6Data map[string]interface{}) error {
	routerID, ok := srv6Data["igp_router_id"].(string)
	if !ok || routerID == "" {
		return fmt.Errorf("missing or invalid igp_router_id in SRv6 SID data")
	}

	domainID, ok := srv6Data["domain_id"]
	if !ok {
		return fmt.Errorf("missing domain_id in SRv6 SID data")
	}

	srv6SID := getString(srv6Data, "srv6_sid")
	if srv6SID == "" {
		return fmt.Errorf("missing srv6_sid in SRv6 SID data")
	}

	// Query to find the IGP node
	query := `
		FOR node IN @@collection
		FILTER node.igp_router_id == @routerId
		FILTER node.domain_id == @domainId
		RETURN node`

	bindVars := map[string]interface{}{
		"@collection": a.config.IGPNode,
		"routerId":    routerID,
		"domainId":    domainID,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to execute IGP node query: %w", err)
	}
	defer cursor.Close()

	var igpNode IGPNode
	meta, err := cursor.ReadDocument(ctx, &igpNode)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			glog.V(6).Infof("No IGP node found for router %s, domain %v", routerID, domainID)
			return nil
		}
		return fmt.Errorf("error reading IGP node document: %w", err)
	}

	// Remove the SID from the SIDS array
	var updatedSIDs []SID
	found := false
	for _, existingSID := range igpNode.SIDS {
		if existingSID.SRv6SID != srv6SID {
			updatedSIDs = append(updatedSIDs, existingSID)
		} else {
			found = true
		}
	}

	if !found {
		glog.V(6).Infof("SRv6 SID %s not found in IGP node %s", srv6SID, meta.Key)
		return nil
	}

	// Update the IGP node with the modified SIDS array
	igpNode.SIDS = updatedSIDs
	if _, err := a.igpNode.UpdateDocument(ctx, meta.Key, igpNode); err != nil {
		return fmt.Errorf("failed to update IGP node after SID removal: %w", err)
	}

	glog.V(6).Infof("Removed SRv6 SID %s from IGP node %s", srv6SID, meta.Key)
	return nil
}

// findAndProcessSRv6SIDsForNode finds all SRv6 SIDs for a given node during node processing
func (a *arangoDB) findAndProcessSRv6SIDsForNode(ctx context.Context, routerID string, domainID interface{}) error {
	query := `
		FOR sid IN @@collection
		FILTER sid.igp_router_id == @routerId
		FILTER sid.domain_id == @domainId
		FILTER sid.srv6_sid != null
		RETURN sid`

	bindVars := map[string]interface{}{
		"@collection": a.config.LSSRv6SID,
		"routerId":    routerID,
		"domainId":    domainID,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to execute SRv6 SID query: %w", err)
	}
	defer cursor.Close()

	for {
		var srv6Data map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &srv6Data)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return fmt.Errorf("error reading SRv6 SID document: %w", err)
		}

		// Process each SRv6 SID found for this node
		if err := a.processInitialSRv6SID(ctx, srv6Data); err != nil {
			glog.Warningf("Failed to process SRv6 SID for node %s: %v", routerID, err)
		}
	}

	return nil
}

// extractEndpointBehavior converts raw endpoint behavior data to srv6.EndpointBehavior
func extractEndpointBehavior(data map[string]interface{}) *srv6.EndpointBehavior {
	if data == nil {
		return nil
	}

	epb := &srv6.EndpointBehavior{}

	// Extract algorithm (this is the key field the user wants)
	if algo, ok := data["algo"].(float64); ok {
		epb.Algorithm = uint8(algo)
	} else if algo, ok := data["algo"].(int); ok {
		epb.Algorithm = uint8(algo)
	}

	// Extract endpoint behavior code
	if behavior, ok := data["endpoint_behavior"].(float64); ok {
		epb.EndpointBehavior = uint16(behavior)
	} else if behavior, ok := data["endpoint_behavior"].(int); ok {
		epb.EndpointBehavior = uint16(behavior)
	}

	// Extract flags
	if flag, ok := data["flag"].(float64); ok {
		epb.Flag = uint8(flag)
	} else if flag, ok := data["flag"].(int); ok {
		epb.Flag = uint8(flag)
	}

	return epb
}

// extractSIDStructure converts raw SID structure data to srv6.SIDStructure
func extractSIDStructure(data map[string]interface{}) *srv6.SIDStructure {
	if data == nil {
		return nil
	}

	sidStruct := &srv6.SIDStructure{}

	// Extract locator block length
	if val, ok := data["locator_block_length"].(float64); ok {
		sidStruct.LBLength = uint8(val)
	} else if val, ok := data["locator_block_length"].(int); ok {
		sidStruct.LBLength = uint8(val)
	}

	// Extract locator node length
	if val, ok := data["locator_node_length"].(float64); ok {
		sidStruct.LNLength = uint8(val)
	} else if val, ok := data["locator_node_length"].(int); ok {
		sidStruct.LNLength = uint8(val)
	}

	// Extract function length
	if val, ok := data["function_length"].(float64); ok {
		sidStruct.FunLength = uint8(val)
	} else if val, ok := data["function_length"].(int); ok {
		sidStruct.FunLength = uint8(val)
	}

	// Extract argument length
	if val, ok := data["argument_length"].(float64); ok {
		sidStruct.ArgLength = uint8(val)
	} else if val, ok := data["argument_length"].(int); ok {
		sidStruct.ArgLength = uint8(val)
	}

	return sidStruct
}

// Helper function to safely get string values from map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}
