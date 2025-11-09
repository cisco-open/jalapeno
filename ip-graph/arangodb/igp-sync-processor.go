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
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// IGPSyncProcessor handles syncing IGP topology changes from igpv4_graph/igpv6_graph
// to the full topology ipv4_graph/ipv6_graph collections
type IGPSyncProcessor struct {
	db       *arangoDB
	stopCh   chan struct{}
	interval time.Duration
}

// NewIGPSyncProcessor creates a new IGP sync processor
func NewIGPSyncProcessor(db *arangoDB) *IGPSyncProcessor {
	return &IGPSyncProcessor{
		db:       db,
		stopCh:   make(chan struct{}),
		interval: 5 * time.Second, // Default 10 second reconciliation interval
	}
}

// syncIGPNodeUpdate syncs an IGP node change to the full topology graphs
// This copies the node from igp_node to ipv4_graph/ipv6_graph if it has edges
func (isp *IGPSyncProcessor) syncIGPNodeUpdate(ctx context.Context, nodeKey, action string) error {
	glog.V(7).Infof("Syncing IGP node update: %s action: %s", nodeKey, action)

	switch action {
	case "del":
		// Node deletion is handled by edge deletions - IGP nodes in ipv4/ipv6 graphs
		// are referenced by edges, so no direct deletion needed here
		glog.V(8).Infof("IGP node %s deleted, edges will handle cleanup", nodeKey)
		return nil

	case "add", "update":
		// Node add/update doesn't require action here - edges reference nodes
		// and igp-graph maintains the igp_node collection
		glog.V(8).Infof("IGP node %s updated, no direct sync needed", nodeKey)
		return nil

	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// syncIGPLinkUpdate syncs an IGP link (edge) change from igpv4_graph/igpv6_graph to ipv4_graph/ipv6_graph
func (isp *IGPSyncProcessor) syncIGPLinkUpdate(ctx context.Context, linkKey, action string, isIPv4 bool) error {
	graphVersion := "IPv6"
	if isIPv4 {
		graphVersion = "IPv4"
	}
	glog.V(7).Infof("Syncing IGP %s link update: %s action: %s", graphVersion, linkKey, action)

	// Determine source and target collections
	var sourceCollection driver.Collection
	var targetCollection driver.Collection

	if isIPv4 {
		sourceCollection = isp.db.igpv4Graph
		targetCollection = isp.db.ipv4Graph
	} else {
		sourceCollection = isp.db.igpv6Graph
		targetCollection = isp.db.ipv6Graph
	}

	switch action {
	case "del":
		return isp.syncLinkDeletion(ctx, linkKey, targetCollection, graphVersion)

	case "add", "update":
		return isp.syncLinkAddUpdate(ctx, linkKey, sourceCollection, targetCollection, graphVersion)

	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// syncLinkDeletion removes an IGP link from the full topology graph
func (isp *IGPSyncProcessor) syncLinkDeletion(ctx context.Context, linkKey string, targetCollection driver.Collection, graphVersion string) error {
	glog.V(8).Infof("Deleting %s IGP link: %s", graphVersion, linkKey)

	// Try to remove the edge from the target collection
	_, err := targetCollection.RemoveDocument(ctx, linkKey)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			glog.V(8).Infof("Link %s not found in %s, already deleted", linkKey, targetCollection.Name())
			return nil
		}
		return fmt.Errorf("failed to remove link %s: %w", linkKey, err)
	}

	glog.V(8).Infof("Successfully deleted %s IGP link: %s", graphVersion, linkKey)
	return nil
}

// syncLinkAddUpdate copies an IGP link from igpvX_graph to ipvX_graph
func (isp *IGPSyncProcessor) syncLinkAddUpdate(ctx context.Context, linkKey string, sourceCollection, targetCollection driver.Collection, graphVersion string) error {
	glog.V(8).Infof("Syncing %s IGP link add/update: %s", graphVersion, linkKey)

	// Read the edge data from the source IGP graph
	var edgeData map[string]interface{}
	_, err := sourceCollection.ReadDocument(ctx, linkKey, &edgeData)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			glog.V(6).Infof("Link %s not found in source %s, may have been deleted", linkKey, sourceCollection.Name())
			return nil
		}
		return fmt.Errorf("failed to read link %s from source: %w", linkKey, err)
	}

	// Remove ArangoDB metadata fields before copying
	delete(edgeData, "_id")
	delete(edgeData, "_rev")

	// Insert or update in the target collection
	_, err = targetCollection.CreateDocument(ctx, edgeData)
	if err != nil {
		if driver.IsConflict(err) {
			// Edge already exists, update it
			if _, err := targetCollection.UpdateDocument(ctx, linkKey, edgeData); err != nil {
				return fmt.Errorf("failed to update link %s in target: %w", linkKey, err)
			}
			glog.V(9).Infof("Updated %s IGP link: %s", graphVersion, linkKey)
		} else {
			return fmt.Errorf("failed to create link %s in target: %w", linkKey, err)
		}
	} else {
		glog.V(9).Infof("Created %s IGP link: %s", graphVersion, linkKey)
	}

	return nil
}

// syncIGPPrefixUpdate syncs an IGP prefix change (if needed for full topology)
func (isp *IGPSyncProcessor) syncIGPPrefixUpdate(ctx context.Context, prefixKey, action string) error {
	glog.V(8).Infof("Syncing IGP prefix update: %s action: %s", prefixKey, action)

	// IGP prefixes from ls_prefix are typically represented as edges in the graph
	// The edge processing already handles this via link updates
	// Prefix vertices are primarily for BGP prefixes

	// For now, no direct action needed for prefix updates
	// If specific prefix handling is needed, implement here

	return nil
}

// syncIGPSRv6Update syncs an IGP SRv6 SID change (if needed for full topology)
func (isp *IGPSyncProcessor) syncIGPSRv6Update(ctx context.Context, srv6Key, action string) error {
	glog.V(8).Infof("Syncing IGP SRv6 update: %s action: %s", srv6Key, action)

	// SRv6 SIDs are typically stored as metadata on IGP nodes
	// The igp_node collection is shared, so SRv6 data is already available
	// No direct sync needed unless SRv6 creates separate vertices

	return nil
}

// InitialIGPSync performs initial bulk sync of IGP topology to full topology graphs
// This is called during startup to populate ipv4_graph/ipv6_graph with IGP data
func (isp *IGPSyncProcessor) InitialIGPSync(ctx context.Context) error {
	glog.Info("Performing initial IGP topology sync...")

	// Sync IPv4 IGP edges
	if err := isp.syncAllIGPEdges(ctx, true); err != nil {
		return fmt.Errorf("failed to sync IPv4 IGP edges: %w", err)
	}

	// Sync IPv6 IGP edges
	if err := isp.syncAllIGPEdges(ctx, false); err != nil {
		return fmt.Errorf("failed to sync IPv6 IGP edges: %w", err)
	}

	glog.Info("Initial IGP topology sync completed")
	return nil
}

// syncAllIGPEdges copies all edges from igpvX_graph to ipvX_graph
func (isp *IGPSyncProcessor) syncAllIGPEdges(ctx context.Context, isIPv4 bool) error {
	graphVersion := "IPv6"
	var sourceCollection driver.Collection
	var targetCollection driver.Collection

	if isIPv4 {
		graphVersion = "IPv4"
		sourceCollection = isp.db.igpv4Graph
		targetCollection = isp.db.ipv4Graph
	} else {
		sourceCollection = isp.db.igpv6Graph
		targetCollection = isp.db.ipv6Graph
	}

	glog.V(6).Infof("Syncing all %s IGP edges from %s to %s...",
		graphVersion, sourceCollection.Name(), targetCollection.Name())

	// Use AQL query to copy all edges efficiently
	// Filter out any BGP edges (safety check - igpv4/v6_graph should only have IGP edges)
	query := fmt.Sprintf(`
		FOR edge IN %s
		FILTER edge.protocol_id != null OR edge.protocol NOT LIKE "BGP_%%"
		INSERT UNSET(edge, "_id", "_rev") INTO %s
		OPTIONS { overwriteMode: "update" }
	`, sourceCollection.Name(), targetCollection.Name())

	cursor, err := isp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to sync %s IGP edges: %w", graphVersion, err)
	}
	defer cursor.Close()

	glog.V(6).Infof("Successfully synced all %s IGP edges", graphVersion)
	return nil
}

// StartReconciliation starts periodic reconciliation of IGP topology
// This ensures IGP changes are eventually synced even if individual updates are missed
func (isp *IGPSyncProcessor) StartReconciliation() {
	glog.Infof("Starting IGP topology reconciliation (interval: %v)", isp.interval)

	go func() {
		ticker := time.NewTicker(isp.interval)
		defer ticker.Stop()

		for {
			select {
			case <-isp.stopCh:
				glog.Info("Stopping IGP topology reconciliation")
				return
			case <-ticker.C:
				if err := isp.reconcile(); err != nil {
					glog.Errorf("IGP reconciliation failed: %v", err)
				}
			}
		}
	}()
}

// StopReconciliation stops the periodic reconciliation
func (isp *IGPSyncProcessor) StopReconciliation() {
	close(isp.stopCh)
}

// reconcile performs a full reconciliation of IGP topology
func (isp *IGPSyncProcessor) reconcile() error {
	ctx := context.Background()

	glog.V(7).Info("Starting IGP topology reconciliation cycle...")

	// Step 1: Add missing edges (edges in igpvX_graph but not in ipvX_graph)
	if err := isp.reconcileIPv4Edges(ctx); err != nil {
		glog.Errorf("Failed to reconcile IPv4 IGP edges: %v", err)
	}

	if err := isp.reconcileIPv6Edges(ctx); err != nil {
		glog.Errorf("Failed to reconcile IPv6 IGP edges: %v", err)
	}

	// Step 2: Remove stale edges (IGP edges in ipvX_graph but not in igpvX_graph)
	if err := isp.removeStaleIPv4Edges(ctx); err != nil {
		glog.Errorf("Failed to remove stale IPv4 IGP edges: %v", err)
	}

	if err := isp.removeStaleIPv6Edges(ctx); err != nil {
		glog.Errorf("Failed to remove stale IPv6 IGP edges: %v", err)
	}

	glog.V(7).Info("IGP topology reconciliation cycle completed")
	return nil
}

// reconcileIPv4Edges ensures all edges from igpv4_graph exist in ipv4_graph
func (isp *IGPSyncProcessor) reconcileIPv4Edges(ctx context.Context) error {
	// Query to find edges in igpv4_graph that don't exist in ipv4_graph
	// and copy them over (excluding any BGP edges as a safety check)
	query := fmt.Sprintf(`
		FOR igp_edge IN %s
		FILTER igp_edge.protocol_id != null OR igp_edge.protocol NOT LIKE "BGP_%%"
		LET exists = (
			FOR ip_edge IN %s
			FILTER ip_edge._key == igp_edge._key
			LIMIT 1
			RETURN 1
		)
		FILTER LENGTH(exists) == 0
		INSERT UNSET(igp_edge, "_id", "_rev") INTO %s
		OPTIONS { overwriteMode: "update" }
		RETURN NEW._key
	`, isp.db.igpv4Graph.Name(), isp.db.ipv4Graph.Name(), isp.db.ipv4Graph.Name())

	cursor, err := isp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to reconcile IPv4 IGP edges: %w", err)
	}
	defer cursor.Close()

	// Count synced edges
	syncedCount := 0
	for cursor.HasMore() {
		var key string
		if _, err := cursor.ReadDocument(ctx, &key); err == nil {
			syncedCount++
		}
	}

	if syncedCount > 0 {
		glog.V(6).Infof("Reconciled %d missing IPv4 IGP edges", syncedCount)
	}

	return nil
}

// reconcileIPv6Edges ensures all edges from igpv6_graph exist in ipv6_graph
func (isp *IGPSyncProcessor) reconcileIPv6Edges(ctx context.Context) error {
	// Query to find edges in igpv6_graph that don't exist in ipv6_graph
	// and copy them over (excluding any BGP edges as a safety check)
	query := fmt.Sprintf(`
		FOR igp_edge IN %s
		FILTER igp_edge.protocol_id != null OR igp_edge.protocol NOT LIKE "BGP_%%"
		LET exists = (
			FOR ip_edge IN %s
			FILTER ip_edge._key == igp_edge._key
			LIMIT 1
			RETURN 1
		)
		FILTER LENGTH(exists) == 0
		INSERT UNSET(igp_edge, "_id", "_rev") INTO %s
		OPTIONS { overwriteMode: "update" }
		RETURN NEW._key
	`, isp.db.igpv6Graph.Name(), isp.db.ipv6Graph.Name(), isp.db.ipv6Graph.Name())

	cursor, err := isp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to reconcile IPv6 IGP edges: %w", err)
	}
	defer cursor.Close()

	// Count synced edges
	syncedCount := 0
	for cursor.HasMore() {
		var key string
		if _, err := cursor.ReadDocument(ctx, &key); err == nil {
			syncedCount++
		}
	}

	if syncedCount > 0 {
		glog.V(6).Infof("Reconciled %d missing IPv6 IGP edges", syncedCount)
	}

	return nil
}

// removeStaleIPv4Edges removes IGP edges from ipv4_graph that no longer exist in igpv4_graph
func (isp *IGPSyncProcessor) removeStaleIPv4Edges(ctx context.Context) error {
	// Query to find IGP edges in ipv4_graph that don't exist in igpv4_graph anymore
	// IGP edges have protocol_id field set (from BGP-LS), BGP edges don't
	query := fmt.Sprintf(`
		FOR ip_edge IN %s
		FILTER ip_edge.protocol_id != null  // Only IGP edges (from BGP-LS)
		LET exists = (
			FOR igp_edge IN %s
			FILTER igp_edge._key == ip_edge._key
			LIMIT 1
			RETURN 1
		)
		FILTER LENGTH(exists) == 0  // Not in igpv4_graph anymore
		REMOVE ip_edge IN %s
		RETURN OLD._key
	`, isp.db.ipv4Graph.Name(), isp.db.igpv4Graph.Name(), isp.db.ipv4Graph.Name())

	cursor, err := isp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to remove stale IPv4 IGP edges: %w", err)
	}
	defer cursor.Close()

	// Count removed edges
	removedCount := 0
	for cursor.HasMore() {
		var key string
		if _, err := cursor.ReadDocument(ctx, &key); err == nil {
			removedCount++
		}
	}

	if removedCount > 0 {
		glog.V(6).Infof("Removed %d stale IPv4 IGP edges", removedCount)
	}

	return nil
}

// removeStaleIPv6Edges removes IGP edges from ipv6_graph that no longer exist in igpv6_graph
func (isp *IGPSyncProcessor) removeStaleIPv6Edges(ctx context.Context) error {
	// Query to find IGP edges in ipv6_graph that don't exist in igpv6_graph anymore
	// IGP edges have protocol_id field set (from BGP-LS), BGP edges don't
	query := fmt.Sprintf(`
		FOR ip_edge IN %s
		FILTER ip_edge.protocol_id != null  // Only IGP edges (from BGP-LS)
		LET exists = (
			FOR igp_edge IN %s
			FILTER igp_edge._key == ip_edge._key
			LIMIT 1
			RETURN 1
		)
		FILTER LENGTH(exists) == 0  // Not in igpv6_graph anymore
		REMOVE ip_edge IN %s
		RETURN OLD._key
	`, isp.db.ipv6Graph.Name(), isp.db.igpv6Graph.Name(), isp.db.ipv6Graph.Name())

	cursor, err := isp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to remove stale IPv6 IGP edges: %w", err)
	}
	defer cursor.Close()

	// Count removed edges
	removedCount := 0
	for cursor.HasMore() {
		var key string
		if _, err := cursor.ReadDocument(ctx, &key); err == nil {
			removedCount++
		}
	}

	if removedCount > 0 {
		glog.V(6).Infof("Removed %d stale IPv6 IGP edges", removedCount)
	}

	return nil
}
