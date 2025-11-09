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
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/cisco-open/jalapeno/gobmp-arango/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/tools"
)

// Config holds the configuration for the IGP Graph processor
type Config struct {
	URL               string
	User              string
	Password          string
	Database          string
	LSPrefix          string
	LSLink            string
	LSSRv6SID         string
	LSNode            string
	IGPDomain         string
	IGPNode           string
	IGPv4Graph        string
	IGPv6Graph        string
	LSNodeEdge        string
	BatchSize         int
	ConcurrentWorkers int
	Notifier          kafkanotifier.Event
}

type arangoDB struct {
	dbclient.DB
	*ArangoConn
	config Config

	// Collections
	lsprefix   driver.Collection
	lslink     driver.Collection
	lssrv6sid  driver.Collection
	lsnode     driver.Collection
	igpDomain  driver.Collection
	igpNode    driver.Collection
	lsNodeEdge driver.Collection

	// Graphs
	igpv4Graph driver.Graph
	igpv6Graph driver.Graph

	// Performance components
	batchProcessor    *BatchProcessor
	updateCoordinator *UpdateCoordinator

	// Control
	stop     chan struct{}
	notifier kafkanotifier.Event
}

// NewDBSrvClient creates a new unified IGP Graph database client
func NewDBSrvClient(config Config) (dbclient.Srv, error) {
	if err := tools.URLAddrValidation(config.URL); err != nil {
		return nil, err
	}

	arangoConn, err := NewArango(ArangoConfig{
		URL:      config.URL,
		User:     config.User,
		Password: config.Password,
		Database: config.Database,
	})
	if err != nil {
		return nil, err
	}

	arango := &arangoDB{
		config:   config,
		stop:     make(chan struct{}),
		notifier: config.Notifier,
	}
	arango.DB = arango
	arango.ArangoConn = arangoConn

	// Initialize collections
	if err := arango.initializeCollections(); err != nil {
		return nil, fmt.Errorf("failed to initialize collections: %w", err)
	}

	// Initialize graphs
	if err := arango.initializeGraphs(); err != nil {
		return nil, fmt.Errorf("failed to initialize graphs: %w", err)
	}

	// Initialize performance components
	arango.batchProcessor = NewBatchProcessor(config.BatchSize, config.ConcurrentWorkers)
	arango.updateCoordinator = NewUpdateCoordinator(arango, config.BatchSize)

	glog.Infof("IGP Graph processor initialized with %d workers, batch size %d",
		config.ConcurrentWorkers, config.BatchSize)

	return arango, nil
}

func (a *arangoDB) initializeCollections() error {
	ctx := context.TODO()
	var err error

	// Check if base link state collections exist
	a.lsprefix, err = a.db.Collection(ctx, a.config.LSPrefix)
	if err != nil {
		return fmt.Errorf("ls_prefix collection not found: %w", err)
	}

	a.lslink, err = a.db.Collection(ctx, a.config.LSLink)
	if err != nil {
		return fmt.Errorf("ls_link collection not found: %w", err)
	}

	a.lssrv6sid, err = a.db.Collection(ctx, a.config.LSSRv6SID)
	if err != nil {
		return fmt.Errorf("ls_srv6_sid collection not found: %w", err)
	}

	a.lsnode, err = a.db.Collection(ctx, a.config.LSNode)
	if err != nil {
		return fmt.Errorf("ls_node collection not found: %w", err)
	}

	// Initialize or create IGP collections
	if err := a.ensureCollection(a.config.IGPDomain, false); err != nil {
		return err
	}
	a.igpDomain, err = a.db.Collection(ctx, a.config.IGPDomain)
	if err != nil {
		return err
	}

	if err := a.ensureCollection(a.config.IGPNode, false); err != nil {
		return err
	}
	a.igpNode, err = a.db.Collection(ctx, a.config.IGPNode)
	if err != nil {
		return err
	}

	// Create ls_node_edge collection for backward compatibility
	if err := a.ensureCollection(a.config.LSNodeEdge, true); err != nil {
		return err
	}
	a.lsNodeEdge, err = a.db.Collection(ctx, a.config.LSNodeEdge)
	if err != nil {
		return err
	}

	return nil
}

func (a *arangoDB) initializeGraphs() error {
	var err error

	// Initialize IGPv4 graph
	a.igpv4Graph, err = a.ensureGraph(a.config.IGPv4Graph, a.config.IGPNode)
	if err != nil {
		return fmt.Errorf("failed to initialize IGPv4 graph: %w", err)
	}

	// Initialize IGPv6 graph
	a.igpv6Graph, err = a.ensureGraph(a.config.IGPv6Graph, a.config.IGPNode)
	if err != nil {
		return fmt.Errorf("failed to initialize IGPv6 graph: %w", err)
	}

	return nil
}

func (a *arangoDB) ensureCollection(name string, isEdge bool) error {
	ctx := context.TODO()

	found, err := a.db.CollectionExists(ctx, name)
	if err != nil {
		return err
	}

	if !found {
		options := &driver.CreateCollectionOptions{}
		if isEdge {
			options.Type = driver.CollectionTypeEdge
		}

		glog.V(5).Infof("Creating collection: %s", name)
		_, err = a.db.CreateCollection(ctx, name, options)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *arangoDB) ensureGraph(graphName, vertexCollection string) (driver.Graph, error) {
	ctx := context.TODO()

	found, err := a.db.GraphExists(ctx, graphName)
	if err != nil {
		return nil, err
	}

	if found {
		graph, err := a.db.Graph(ctx, graphName)
		if err != nil {
			return nil, err
		}
		glog.V(5).Infof("Found existing graph: %s", graphName)
		return graph, nil
	}

	// Create edge collection for the graph
	edgeCollectionName := graphName
	if err := a.ensureCollection(edgeCollectionName, true); err != nil {
		return nil, err
	}

	// Create graph with edge definition
	var edgeDefinition driver.EdgeDefinition
	edgeDefinition.Collection = edgeCollectionName
	edgeDefinition.From = []string{vertexCollection}
	edgeDefinition.To = []string{vertexCollection}

	var options driver.CreateGraphOptions
	options.OrphanVertexCollections = []string{a.config.LSPrefix, a.config.LSSRv6SID}
	options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}

	graph, err := a.db.CreateGraphV2(ctx, graphName, &options)
	if err != nil {
		return nil, err
	}

	glog.V(5).Infof("Created new graph: %s", graphName)
	return graph, nil
}

func (a *arangoDB) Start() error {
	if err := a.loadInitialData(); err != nil {
		return fmt.Errorf("failed to load initial data: %w", err)
	}

	glog.Info("Starting IGP Graph processor components...")

	// Start batch processor
	if err := a.batchProcessor.Start(); err != nil {
		return fmt.Errorf("failed to start batch processor: %w", err)
	}

	// Start update coordinator
	if err := a.updateCoordinator.Start(); err != nil {
		return fmt.Errorf("failed to start update coordinator: %w", err)
	}

	glog.Info("IGP Graph processor started successfully")
	go a.monitor()

	return nil
}

func (a *arangoDB) Stop() error {
	glog.Info("Stopping IGP Graph processor...")

	close(a.stop)

	// Stop components
	if a.updateCoordinator != nil {
		a.updateCoordinator.Stop()
	}

	if a.batchProcessor != nil {
		a.batchProcessor.Stop()
	}

	glog.Info("IGP Graph processor stopped")
	return nil
}

func (a *arangoDB) GetInterface() dbclient.DB {
	return a.DB
}

func (a *arangoDB) GetArangoDBInterface() *ArangoConn {
	return a.ArangoConn
}

func (a *arangoDB) StoreMessage(msgType dbclient.CollectionType, msg []byte) error {
	return a.updateCoordinator.ProcessMessage(msgType, msg)
}

func (a *arangoDB) loadInitialData() error {
	glog.Info("Loading initial IGP topology data...")
	ctx := context.TODO()

	// Load initial nodes
	if err := a.loadInitialNodes(ctx); err != nil {
		return fmt.Errorf("failed to load initial nodes: %w", err)
	}

	// Run deduplication BEFORE processing links to avoid orphaned references
	if err := a.runDeduplication(); err != nil {
		return fmt.Errorf("failed to deduplicate IGP nodes: %w", err)
	}

	// Clean up any existing orphaned graph edges from previous runs
	if err := a.cleanupOrphanedEdges(ctx); err != nil {
		glog.Warningf("Failed to cleanup orphaned edges: %v", err)
	}

	// Load initial links (after deduplication)
	if err := a.loadInitialLinks(ctx); err != nil {
		return fmt.Errorf("failed to load initial links: %w", err)
	}

	// Load initial SRv6 SIDs
	if err := a.loadInitialSRv6SIDs(ctx); err != nil {
		return fmt.Errorf("failed to load initial SRv6 SIDs: %w", err)
	}

	// Load initial prefixes
	if err := a.loadInitialPrefixes(ctx); err != nil {
		return fmt.Errorf("failed to load initial prefixes: %w", err)
	}

	glog.Info("Initial IGP topology data loaded successfully")
	return nil
}

func (a *arangoDB) loadInitialNodes(ctx context.Context) error {
	// Query all ls_node documents and process them
	query := fmt.Sprintf("FOR doc IN %s RETURN doc", a.config.LSNode)
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()

	count := 0
	for {
		var node map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &node)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return err
		}

		// Process node (simplified for now)
		if err := a.processInitialNode(ctx, node); err != nil {
			glog.Warningf("Failed to process initial node %v: %v", node["_key"], err)
			continue
		}
		count++

		if count%1000 == 0 {
			glog.V(3).Infof("Loaded %d nodes...", count)
		}
	}

	glog.Infof("Loaded %d initial nodes", count)
	return nil
}

func (a *arangoDB) loadInitialLinks(ctx context.Context) error {
	// Query all ls_link documents and process them
	query := fmt.Sprintf("FOR doc IN %s RETURN doc", a.config.LSLink)
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()

	count := 0
	for {
		var link map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &link)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return err
		}

		// Process link (simplified for now)
		if err := a.processInitialLink(ctx, link); err != nil {
			glog.Warningf("Failed to process initial link %v: %v", link["_key"], err)
			continue
		}
		count++

		if count%1000 == 0 {
			glog.V(3).Infof("Loaded %d links...", count)
		}
	}

	glog.Infof("Loaded %d initial links", count)
	return nil
}

// cleanupOrphanedEdges removes graph edges that reference non-existent nodes
func (a *arangoDB) cleanupOrphanedEdges(ctx context.Context) error {
	glog.V(6).Info("Cleaning up orphaned graph edges...")

	// Clean up IPv4 graph edges
	v4Query := fmt.Sprintf(`
		FOR edge IN %s
		LET localExists = LENGTH(FOR n IN %s FILTER n._id == edge._from RETURN 1) > 0
		LET remoteExists = LENGTH(FOR n IN %s FILTER n._id == edge._to RETURN 1) > 0
		FILTER !localExists OR !remoteExists
		REMOVE edge IN %s
		RETURN OLD._key
	`, a.config.IGPv4Graph, a.config.IGPNode, a.config.IGPNode, a.config.IGPv4Graph)

	cursor, err := a.db.Query(ctx, v4Query, nil)
	if err != nil {
		return fmt.Errorf("failed to cleanup IPv4 orphaned edges: %w", err)
	}
	v4Count := 0
	for cursor.HasMore() {
		var key string
		_, err := cursor.ReadDocument(ctx, &key)
		if err != nil {
			break
		}
		v4Count++
	}
	cursor.Close()

	// Clean up IPv6 graph edges
	v6Query := fmt.Sprintf(`
		FOR edge IN %s
		LET localExists = LENGTH(FOR n IN %s FILTER n._id == edge._from RETURN 1) > 0
		LET remoteExists = LENGTH(FOR n IN %s FILTER n._id == edge._to RETURN 1) > 0
		FILTER !localExists OR !remoteExists
		REMOVE edge IN %s
		RETURN OLD._key
	`, a.config.IGPv6Graph, a.config.IGPNode, a.config.IGPNode, a.config.IGPv6Graph)

	cursor, err = a.db.Query(ctx, v6Query, nil)
	if err != nil {
		return fmt.Errorf("failed to cleanup IPv6 orphaned edges: %w", err)
	}
	v6Count := 0
	for cursor.HasMore() {
		var key string
		_, err := cursor.ReadDocument(ctx, &key)
		if err != nil {
			break
		}
		v6Count++
	}
	cursor.Close()

	if v4Count > 0 || v6Count > 0 {
		glog.Infof("Cleaned up %d IPv4 and %d IPv6 orphaned graph edges", v4Count, v6Count)
	}

	return nil
}

func (a *arangoDB) processInitialNode(ctx context.Context, node map[string]interface{}) error {
	// Convert map to LSNode-like structure for processing
	key, ok := node["_key"].(string)
	if !ok {
		return fmt.Errorf("invalid node key")
	}

	// Filter out BGP nodes (protocol_id = 7) as they're not part of IGP topology
	if protocolID, ok := node["protocol_id"].(float64); ok && protocolID == 7 {
		glog.V(7).Infof("Skipping BGP node (protocol_id=7): %s", key)
		return nil
	}

	// Ensure IGP domain exists for this node
	if err := a.ensureIGPDomain(ctx, node); err != nil {
		glog.Warningf("Failed to ensure IGP domain for node %s: %v", key, err)
	}

	// Create IGP node entry with enhanced metadata
	igpNodeDoc := map[string]interface{}{
		"_key": key,
		// "action":                     node["action"],
		// "router_hash":                node["router_hash"],
		"domain_id": node["domain_id"],
		// "router_ip":                  node["router_ip"],
		// "peer_hash":                  node["peer_hash"],
		"peer_ip":  node["peer_ip"],
		"peer_asn": node["peer_asn"],
		// "timestamp":                  node["timestamp"],
		"igp_router_id":         node["igp_router_id"],
		"router_id":             node["router_id"],
		"asn":                   node["asn"],
		"mt_id_tlv":             node["mt_id_tlv"],
		"area_id":               node["area_id"],
		"protocol":              node["protocol"],
		"protocol_id":           node["protocol_id"],
		"name":                  node["name"],
		"ls_sr_capabilities":    node["ls_sr_capabilities"],
		"sr_algorithm":          node["sr_algorithm"],
		"sr_local_block":        node["sr_local_block"],
		"srv6_capabilities_tlv": node["srv6_capabilities_tlv"],
		"node_msd":              node["node_msd"],
		"flex_algo_definition":  node["flex_algo_definition"],
		// "is_adj_rib_in_post_policy":  node["is_adj_rib_in_post_policy"],
		// "is_adj_rib_out_post_policy": node["is_adj_rib_out_post_policy"],
		// "is_loc_rib_filtered":        node["is_loc_rib_filtered"],
		"prefix_attr_tlvs": node["prefix_attr_tlvs"],
		// "is_prepolicy":               node["is_prepolicy"],
		// "is_adj_rib_in":              node["is_adj_rib_in"],
		"sids": []SID{}, // Initialize empty SIDs array for SRv6 metadata
	}

	// Try to create the document
	_, err := a.igpNode.CreateDocument(ctx, igpNodeDoc)
	if err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create igp_node document: %w", err)
		}
		// Document exists, update it
		if _, err := a.igpNode.UpdateDocument(ctx, key, igpNodeDoc); err != nil {
			return fmt.Errorf("failed to update igp_node document: %w", err)
		}
	}

	// After creating the IGP node, find and process any associated SRv6 SIDs
	routerID, _ := node["igp_router_id"].(string)
	domainID := node["domain_id"]
	if routerID != "" {
		if err := a.findAndProcessSRv6SIDsForNode(ctx, routerID, domainID); err != nil {
			glog.Warningf("Failed to process SRv6 SIDs for node %s: %v", routerID, err)
		}
	}

	glog.V(9).Infof("Processed initial node: %s", key)
	return nil
}

func (a *arangoDB) processInitialLink(ctx context.Context, link map[string]interface{}) error {
	// Convert map to LSLink-like structure for processing
	key, ok := link["_key"].(string)
	if !ok {
		return fmt.Errorf("invalid link key")
	}

	// Filter out BGP links (protocol_id = 7) as they're not part of IGP topology
	if protocolID, ok := link["protocol_id"].(float64); ok && protocolID == 7 {
		glog.V(7).Infof("Skipping BGP link (protocol_id=7): %s", key)
		return nil
	}

	// Create ls_node_edge entry for backward compatibility
	lsNodeEdgeDoc := map[string]interface{}{
		"_key":                  key,
		"_from":                 fmt.Sprintf("%s/%s", a.config.LSNode, link["igp_router_id"]),
		"_to":                   fmt.Sprintf("%s/%s", a.config.LSNode, link["remote_igp_router_id"]),
		"link":                  key,
		"protocol_id":           link["protocol_id"],
		"domain_id":             link["domain_id"],
		"area_id":               link["area_id"],
		"local_link_ip":         link["local_link_ip"],
		"remote_link_ip":        link["remote_link_ip"],
		"igp_metric":            link["igp_metric"],
		"local_node_asn":        link["local_node_asn"],
		"remote_node_asn":       link["remote_node_asn"],
		"max_link_bw":           link["max_link_bw"],
		"max_resv_bw":           link["max_resv_bw"],
		"te_default_metric":     link["te_default_metric"],
		"unidir_link_delay":     link["unidir_link_delay"],
		"unidir_packet_loss":    link["unidir_packet_loss"],
		"unidir_available_bw":   link["unidir_available_bw"],
		"unidir_bw_utilization": link["unidir_bw_utilization"],
	}

	// Create ls_node_edge document
	_, err := a.lsNodeEdge.CreateDocument(ctx, lsNodeEdgeDoc)
	if err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create ls_node_edge document: %w", err)
		}
		// Document exists, update it
		if _, err := a.lsNodeEdge.UpdateDocument(ctx, key, lsNodeEdgeDoc); err != nil {
			return fmt.Errorf("failed to update ls_node_edge document: %w", err)
		}
	}

	// Create IGP graph edges
	if err := a.createIGPGraphEdges(ctx, link); err != nil {
		return fmt.Errorf("failed to create IGP graph edges: %w", err)
	}

	glog.V(9).Infof("Processed initial link: %s", key)
	return nil
}

// createIGPGraphEdges creates edges in appropriate IPv4 or IPv6 graphs based on MTID
// Following the original linkstate-graph pattern with proper node lookups
func (a *arangoDB) createIGPGraphEdges(ctx context.Context, link map[string]interface{}) error {
	key, _ := link["_key"].(string)

	// Determine if this is IPv4 or IPv6 based on MTID (matching original logic)
	// IPv4: no mt_id_tlv field or mt_id = 0
	// IPv6: mt_id_tlv contains mt_id = 2
	isIPv6 := false

	if mtidTLV, exists := link["mt_id_tlv"]; exists {
		// Handle both array format (from nodes) and object format (from SRv6)
		if mtidArray, ok := mtidTLV.([]interface{}); ok {
			// Array format: search for mt_id = 2
			for _, mtItem := range mtidArray {
				if mtObj, ok := mtItem.(map[string]interface{}); ok {
					if mtID, ok := mtObj["mt_id"].(float64); ok && mtID == 2 {
						isIPv6 = true
						break
					}
				}
			}
		} else if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			// Object format: direct check
			if mtID, ok := mtidObj["mt_id"].(float64); ok && mtID == 2 {
				isIPv6 = true
			}
		}
	}

	glog.V(8).Infof("Link %s: isIPv6 = %t", key, isIPv6)

	// Get local node from IGP node collection (matching original getv4Node/getv6Node)
	localNode, err := a.getIGPNode(ctx, link, true)
	if err != nil {
		glog.Errorf("Failed to get local IGP node %s for link %s: %v",
			link["igp_router_id"], key, err)
		return err
	}

	// Get remote node from IGP node collection
	remoteNode, err := a.getIGPNode(ctx, link, false)
	if err != nil {
		glog.Errorf("Failed to get remote IGP node %s for link %s: %v",
			link["remote_igp_router_id"], key, err)
		return err
	}

	glog.V(7).Infof("Local node -> Protocol: %v Domain ID: %v IGP Router ID: %v",
		localNode["protocol_id"], localNode["domain_id"], localNode["igp_router_id"])
	glog.V(7).Infof("Remote node -> Protocol: %v Domain ID: %v IGP Router ID: %v",
		remoteNode["protocol_id"], remoteNode["domain_id"], remoteNode["igp_router_id"])

	// Create edge in appropriate graph based on IP version (matching original pattern)
	if isIPv6 {
		// IPv6 graph (MTID = 2) - matches original processigpv6LinkEdge
		if err := a.createIGPv6EdgeObject(ctx, link, localNode, remoteNode); err != nil {
			glog.Errorf("Failed to create IPv6 edge object: %v", err)
			return err
		}
	} else {
		// IPv4 graph (MTID = nil or 0) - matches original processLSLinkEdge
		if err := a.createIGPv4EdgeObject(ctx, link, localNode, remoteNode); err != nil {
			glog.Errorf("Failed to create IPv4 edge object: %v", err)
			return err
		}
	}

	return nil
}

func (a *arangoDB) monitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stop:
			return
		case <-ticker.C:
			// Log performance statistics
			if a.batchProcessor != nil {
				stats := a.batchProcessor.GetStats()
				processedCount := stats.Processed.Load()
				pendingCount := stats.Pending.Load()
				glog.V(5).Infof("Batch processor stats: processed=%d, pending=%d",
					processedCount, pendingCount)
			}
		}
	}
}
