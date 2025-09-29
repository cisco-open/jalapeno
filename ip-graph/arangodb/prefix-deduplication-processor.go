package arangodb

import (
	"context"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// PrefixDeduplicationProcessor handles deduplication between IGP and BGP prefixes
type PrefixDeduplicationProcessor struct {
	db *arangoDB
}

// NewPrefixDeduplicationProcessor creates a new prefix deduplication processor
func NewPrefixDeduplicationProcessor(db *arangoDB) *PrefixDeduplicationProcessor {
	return &PrefixDeduplicationProcessor{
		db: db,
	}
}

// ProcessPrefixDeduplication performs IGP-BGP prefix deduplication and creates unified vertices
func (pdp *PrefixDeduplicationProcessor) ProcessPrefixDeduplication(ctx context.Context) error {
	glog.Info("Starting IGP-BGP prefix deduplication...")

	// Process IPv4 prefix deduplication
	if err := pdp.processIPv4PrefixDeduplication(ctx); err != nil {
		return fmt.Errorf("failed to process IPv4 prefix deduplication: %w", err)
	}

	// Process IPv6 prefix deduplication
	if err := pdp.processIPv6PrefixDeduplication(ctx); err != nil {
		return fmt.Errorf("failed to process IPv6 prefix deduplication: %w", err)
	}

	glog.Info("IGP-BGP prefix deduplication completed successfully")
	return nil
}

// processIPv4PrefixDeduplication handles IPv4 prefix conflicts between IGP and BGP
func (pdp *PrefixDeduplicationProcessor) processIPv4PrefixDeduplication(ctx context.Context) error {
	glog.V(6).Info("Processing IPv4 IGP-BGP prefix deduplication...")

	// Find IPv4 prefix conflicts using optimized hash-based approach
	// Look for BOTH directions: igp_node→ls_prefix AND ls_prefix→igp_node
	conflictQuery := `
		FOR edge IN ` + pdp.db.config.IPv4Graph + `
		FILTER (STARTS_WITH(edge._to, "ls_prefix/") OR STARTS_WITH(edge._from, "ls_prefix/"))
		FILTER edge.prefix != null AND edge.prefix_len != null
		
		// Use hash-based lookup instead of nested loop for performance
		LET prefix_key = CONCAT_SEPARATOR("_", edge.prefix, edge.prefix_len)
		LET bgp_match = DOCUMENT(` + pdp.db.config.BGPPrefixV4 + `, prefix_key)
		FILTER bgp_match != null
		
		// Get the corresponding ls_prefix data for metadata
		LET ls_prefix_key = STARTS_WITH(edge._to, "ls_prefix/") ? SPLIT(edge._to, "/")[1] : SPLIT(edge._from, "/")[1]
		LET ls_data = DOCUMENT("ls_prefix", ls_prefix_key)
		FILTER ls_data != null
		
		// Group by prefix to handle both directions together
		COLLECT prefix = edge.prefix, prefix_len = edge.prefix_len INTO edge_group
		
		RETURN {
			prefix: prefix,
			prefix_len: prefix_len,
			ls_data: FIRST(edge_group[*].ls_data),
			bgp_data: FIRST(edge_group[*].bgp_match),
			unified_key: CONCAT_SEPARATOR("_", prefix, prefix_len),
			conflict_type: "graph_edge_bidirectional",
			existing_edges: edge_group[*].edge
		}
	`

	cursor, err := pdp.db.db.Query(ctx, conflictQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to query IPv4 prefix conflicts: %w", err)
	}
	defer cursor.Close()

	conflictCount := 0
	for cursor.HasMore() {
		var conflict map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &conflict); err != nil {
			return fmt.Errorf("failed to read IPv4 conflict: %w", err)
		}

		if err := pdp.createUnifiedPrefixVertex(ctx, conflict, true); err != nil {
			glog.Warningf("Failed to create unified IPv4 prefix vertex: %v", err)
			continue
		}

		conflictCount++
	}

	glog.V(6).Infof("Processed %d IPv4 IGP-BGP prefix conflicts", conflictCount)
	return nil
}

// processIPv6PrefixDeduplication handles IPv6 prefix conflicts between IGP and BGP
func (pdp *PrefixDeduplicationProcessor) processIPv6PrefixDeduplication(ctx context.Context) error {
	glog.V(6).Info("Processing IPv6 IGP-BGP prefix deduplication...")

	// Find IPv6 prefix conflicts using optimized hash-based approach
	// Look for BOTH directions: igp_node→ls_prefix AND ls_prefix→igp_node
	conflictQuery := `
		FOR edge IN ` + pdp.db.config.IPv6Graph + `
		FILTER (STARTS_WITH(edge._to, "ls_prefix/") OR STARTS_WITH(edge._from, "ls_prefix/"))
		FILTER edge.prefix != null AND edge.prefix_len != null
		FILTER edge.mt_id == 2  // IPv6 topology
		
		// Use hash-based lookup instead of nested loop for performance
		LET prefix_key = CONCAT_SEPARATOR("_", edge.prefix, edge.prefix_len)
		LET bgp_match = DOCUMENT(` + pdp.db.config.BGPPrefixV6 + `, prefix_key)
		FILTER bgp_match != null
		
		// Get the corresponding ls_prefix data for metadata
		LET ls_prefix_key = STARTS_WITH(edge._to, "ls_prefix/") ? SPLIT(edge._to, "/")[1] : SPLIT(edge._from, "/")[1]
		LET ls_data = DOCUMENT("ls_prefix", ls_prefix_key)
		FILTER ls_data != null
		
		// Group by prefix to handle both directions together
		COLLECT prefix = edge.prefix, prefix_len = edge.prefix_len INTO edge_group
		
		RETURN {
			prefix: prefix,
			prefix_len: prefix_len,
			ls_data: FIRST(edge_group[*].ls_data),
			bgp_data: FIRST(edge_group[*].bgp_match),
			unified_key: CONCAT_SEPARATOR("_", prefix, prefix_len),
			conflict_type: "graph_edge_bidirectional",
			existing_edges: edge_group[*].edge
		}
	`

	cursor, err := pdp.db.db.Query(ctx, conflictQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to query IPv6 prefix conflicts: %w", err)
	}
	defer cursor.Close()

	conflictCount := 0
	for cursor.HasMore() {
		var conflict map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &conflict); err != nil {
			return fmt.Errorf("failed to read IPv6 conflict: %w", err)
		}

		if err := pdp.createUnifiedPrefixVertex(ctx, conflict, false); err != nil {
			glog.Warningf("Failed to create unified IPv6 prefix vertex: %v", err)
			continue
		}

		conflictCount++
	}

	glog.V(6).Infof("Processed %d IPv6 IGP-BGP prefix conflicts", conflictCount)
	return nil
}

// createUnifiedPrefixVertex creates a unified prefix vertex with both IGP and BGP metadata
func (pdp *PrefixDeduplicationProcessor) createUnifiedPrefixVertex(ctx context.Context, conflict map[string]interface{}, isIPv4 bool) error {
	prefix := getString(conflict, "prefix")
	prefixLen := getUint32FromInterface(conflict["prefix_len"])
	unifiedKey := getString(conflict, "unified_key")
	conflictType := getString(conflict, "conflict_type")

	// Extract IGP and BGP data
	lsData := conflict["ls_data"].(map[string]interface{})
	bgpData := conflict["bgp_data"].(map[string]interface{})

	// Handle missing conflict_type for backward compatibility
	if conflictType == "" {
		conflictType = "collection" // Default for older logic
	}

	glog.V(7).Infof("Creating unified prefix vertex for %s/%d (conflict type: %s)", prefix, prefixLen, conflictType)

	// Create unified prefix document with metadata from both sources
	unifiedPrefix := map[string]interface{}{
		"_key":       unifiedKey,
		"prefix":     prefix,
		"prefix_len": int32(prefixLen),
		"is_unified": true,
		"sources":    []string{"igp", "bgp"},

		// IGP metadata
		"igp_metric":       getUint32FromInterface(lsData["prefix_metric"]),
		"igp_router_id":    getString(lsData, "igp_router_id"),
		"igp_protocol":     getString(lsData, "protocol"),
		"igp_protocol_id":  getUint32FromInterface(lsData["protocol_id"]),
		"igp_area_id":      getString(lsData, "area_id"),
		"prefix_attr_tlvs": lsData["prefix_attr_tlvs"],

		// BGP metadata
		"bgp_origin_as":   getUint32FromInterface(bgpData["origin_as"]),
		"bgp_peer_asn":    getUint32FromInterface(bgpData["peer_asn"]),
		"bgp_prefix_type": getString(bgpData, "prefix_type"),
		"bgp_nexthop":     getString(bgpData, "nexthop"),
		"bgp_router_id":   getString(bgpData, "router_id"),

		// Common metadata
		"timestamp": getString(lsData, "timestamp"), // Use IGP timestamp as primary
		"is_host":   (isIPv4 && prefixLen == 32) || (!isIPv4 && prefixLen == 128),
	}

	// Add BGP-specific fields if available
	if localPref, exists := bgpData["local_pref"]; exists {
		unifiedPrefix["bgp_local_pref"] = localPref
	}

	// Determine target collection
	var targetCollection string
	if isIPv4 {
		targetCollection = pdp.db.config.BGPPrefixV4
	} else {
		targetCollection = pdp.db.config.BGPPrefixV6
	}

	// Update or create the unified prefix vertex
	updateQuery := fmt.Sprintf(`
		UPSERT { _key: @key }
		INSERT @doc
		UPDATE @doc
		IN %s
	`, targetCollection)

	bindVars := map[string]interface{}{
		"key": unifiedKey,
		"doc": unifiedPrefix,
	}

	cursor, err := pdp.db.db.Query(ctx, updateQuery, bindVars)
	if err != nil {
		return fmt.Errorf("failed to create unified prefix vertex: %w", err)
	}
	defer cursor.Close()

	// Handle existing IGP edges if this is a graph_edge conflict
	if conflictType == "graph_edge_bidirectional" {
		if existingEdges, exists := conflict["existing_edges"]; exists {
			edges := existingEdges.([]interface{})
			for _, edgeInterface := range edges {
				if edge, ok := edgeInterface.(map[string]interface{}); ok {
					if err := pdp.handleExistingIGPEdges(ctx, edge, unifiedKey, isIPv4); err != nil {
						glog.Warningf("Failed to handle existing IGP edge %s for %s: %v", getString(edge, "_key"), unifiedKey, err)
					}
				}
			}
		}
	} else if conflictType == "graph_edge" {
		// Handle single edge for backward compatibility
		if existingEdge, exists := conflict["existing_edge"]; exists {
			if err := pdp.handleExistingIGPEdges(ctx, existingEdge.(map[string]interface{}), unifiedKey, isIPv4); err != nil {
				glog.Warningf("Failed to handle existing IGP edges for %s: %v", unifiedKey, err)
			}
		}
	}

	// Remove any duplicate BGP edges for this prefix
	if err := pdp.removeDuplicateBGPEdges(ctx, unifiedKey, isIPv4); err != nil {
		glog.Warningf("Failed to remove duplicate BGP edges for %s: %v", unifiedKey, err)
	}

	// Create edges to both IGP and BGP nodes
	if err := pdp.createUnifiedPrefixEdges(ctx, unifiedPrefix, lsData, bgpData, isIPv4); err != nil {
		return fmt.Errorf("failed to create unified prefix edges: %w", err)
	}

	return nil
}

// createUnifiedPrefixEdges creates edges from the unified prefix to both IGP and BGP nodes
func (pdp *PrefixDeduplicationProcessor) createUnifiedPrefixEdges(ctx context.Context, unifiedPrefix, lsData, bgpData map[string]interface{}, isIPv4 bool) error {
	prefixKey := getString(unifiedPrefix, "_key")
	prefix := getString(unifiedPrefix, "prefix")
	prefixLen := getUint32FromInterface(unifiedPrefix["prefix_len"])

	// Determine collections
	var prefixCollection, graphCollection string
	if isIPv4 {
		prefixCollection = pdp.db.config.BGPPrefixV4
		graphCollection = pdp.db.config.IPv4Graph
	} else {
		prefixCollection = pdp.db.config.BGPPrefixV6
		graphCollection = pdp.db.config.IPv6Graph
	}

	prefixVertexID := fmt.Sprintf("%s/%s", prefixCollection, prefixKey)

	// 1. Create edge to IGP node (based on ls_prefix data) - ONLY for appropriate prefix types
	// For eBGP private prefixes where origin_as == peer_asn, don't create IGP edges
	// The prefix originates from the BGP peer, not from IGP redistribution
	bgpPrefixType := getString(bgpData, "prefix_type")
	bgpOriginAS := getUint32FromInterface(bgpData["origin_as"])
	bgpPeerASN := getUint32FromInterface(bgpData["peer_asn"])

	shouldCreateIGPEdge := false
	switch bgpPrefixType {
	case "ibgp":
		// iBGP prefixes can have IGP edges (internal redistribution)
		shouldCreateIGPEdge = true
	case "ebgp_private", "ebgp_private_4byte":
		// eBGP private: only create IGP edge if origin_as != peer_asn (redistribution case)
		// If origin_as == peer_asn, the prefix originates from the BGP peer directly
		shouldCreateIGPEdge = (bgpOriginAS != bgpPeerASN)
	case "ebgp_public":
		// Internet prefixes: don't create IGP edges, they connect to BGP peers
		shouldCreateIGPEdge = false
	default:
		// Unknown type: be conservative and allow IGP edge
		shouldCreateIGPEdge = true
	}

	if shouldCreateIGPEdge {
		igpRouterID := getString(lsData, "igp_router_id")
		igpNodeID, err := pdp.findIGPNodeByRouterID(ctx, igpRouterID)
		if err != nil {
			glog.Warningf("Failed to find IGP node for router %s: %v", igpRouterID, err)
		} else if igpNodeID != "" {
			// Create bidirectional edges to IGP node
			if err := pdp.createBidirectionalEdges(ctx, prefixVertexID, igpNodeID, prefixKey, prefix, int32(prefixLen), "IGP_unified", graphCollection); err != nil {
				glog.Warningf("Failed to create IGP edges for unified prefix %s: %v", prefixKey, err)
			}
		}
	} else {
		glog.V(7).Infof("Skipping IGP edge for %s prefix %s (origin_as=%d, peer_asn=%d)", bgpPrefixType, prefixKey, bgpOriginAS, bgpPeerASN)
	}

	// 2. Create edge to BGP node (based on bgp_prefix data)
	bgpRouterID := getString(bgpData, "router_id")
	// bgpOriginAS and bgpPrefixType already declared above

	var bgpNodeID string
	var err error
	switch bgpPrefixType {
	case "ibgp":
		// iBGP: Look for IGP node
		bgpNodeID, err = pdp.findIGPNodeByRouterIDAndASN(ctx, bgpRouterID, bgpOriginAS)
	case "ebgp_private", "ebgp_private_4byte", "ebgp_public":
		// eBGP: Look for BGP node, or find BGP peers for public prefixes
		if bgpPrefixType == "ebgp_public" {
			// Internet prefixes connect to all public BGP peers
			return pdp.createInternetPrefixEdges(ctx, prefixVertexID, prefixKey, prefix, int32(prefixLen), graphCollection)
		} else {
			// Private eBGP: specific BGP node
			bgpNodeID, err = pdp.findBGPNodeByRouterIDAndASN(ctx, bgpRouterID, bgpOriginAS)
		}
	}

	if err != nil {
		glog.Warningf("Failed to find BGP node for router %s AS %d: %v", bgpRouterID, bgpOriginAS, err)
	} else if bgpNodeID != "" {
		// Create bidirectional edges to BGP node
		protocol := fmt.Sprintf("BGP_%s_unified", bgpPrefixType)
		if err := pdp.createBidirectionalEdges(ctx, prefixVertexID, bgpNodeID, prefixKey, prefix, int32(prefixLen), protocol, graphCollection); err != nil {
			glog.Warningf("Failed to create BGP edges for unified prefix %s: %v", prefixKey, err)
		}
	}

	return nil
}

// createInternetPrefixEdges creates edges from unified Internet prefix to all public BGP peers
func (pdp *PrefixDeduplicationProcessor) createInternetPrefixEdges(ctx context.Context, prefixVertexID, prefixKey, prefix string, prefixLen int32, graphCollection string) error {
	// Query all BGP nodes with public ASNs
	query := fmt.Sprintf(`
		FOR node IN %s 
		FILTER node.asn NOT IN 64512..65535 
		FILTER node.asn NOT IN 4200000000..4294967294 
		RETURN node._id
	`, pdp.db.config.BGPNode)

	cursor, err := pdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query public BGP nodes: %w", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var bgpNodeID string
		if _, err := cursor.ReadDocument(ctx, &bgpNodeID); err != nil {
			continue
		}

		if err := pdp.createBidirectionalEdges(ctx, prefixVertexID, bgpNodeID, prefixKey, prefix, prefixLen, "BGP_ebgp_public_unified", graphCollection); err != nil {
			glog.Warningf("Failed to create Internet edge for unified prefix %s to node %s: %v", prefixKey, bgpNodeID, err)
		}
	}

	return nil
}

// createBidirectionalEdges creates bidirectional edges between prefix and node
func (pdp *PrefixDeduplicationProcessor) createBidirectionalEdges(ctx context.Context, prefixVertexID, nodeID, prefixKey, prefix string, prefixLen int32, protocol, graphCollection string) error {
	// Extract node key from node ID
	nodeKey := extractKeyFromID(nodeID)

	edges := []*IPGraphObject{
		{
			Key:       fmt.Sprintf("%s_%s", nodeKey, prefixKey),
			From:      nodeID,
			To:        prefixVertexID,
			Protocol:  protocol,
			Link:      prefixKey,
			Prefix:    prefix,
			PrefixLen: prefixLen,
		},
		{
			Key:       fmt.Sprintf("%s_%s", prefixKey, nodeKey),
			From:      prefixVertexID,
			To:        nodeID,
			Protocol:  protocol,
			Link:      prefixKey,
			Prefix:    prefix,
			PrefixLen: prefixLen,
		},
	}

	// Get target collection
	var targetCollection driver.Collection
	if graphCollection == pdp.db.config.IPv4Graph {
		targetCollection = pdp.db.ipv4Graph
	} else {
		targetCollection = pdp.db.ipv6Graph
	}

	// Create both edges
	for _, edge := range edges {
		if _, err := targetCollection.CreateDocument(ctx, edge); err != nil {
			if !driver.IsConflict(err) {
				return fmt.Errorf("failed to create edge %s: %w", edge.Key, err)
			}
			// Update existing edge
			if _, err := targetCollection.UpdateDocument(ctx, edge.Key, edge); err != nil {
				return fmt.Errorf("failed to update edge %s: %w", edge.Key, err)
			}
		}
	}

	return nil
}

// Helper functions for node lookups
func (pdp *PrefixDeduplicationProcessor) findIGPNodeByRouterID(ctx context.Context, routerID string) (string, error) {
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.igp_router_id == @routerId
		LIMIT 1
		RETURN node._id
	`, pdp.db.config.IGPNode)

	bindVars := map[string]interface{}{
		"routerId": routerID,
	}

	cursor, err := pdp.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return "", err
	}
	defer cursor.Close()

	if cursor.HasMore() {
		var nodeID string
		if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
			return "", err
		}
		return nodeID, nil
	}

	return "", nil
}

func (pdp *PrefixDeduplicationProcessor) findIGPNodeByRouterIDAndASN(ctx context.Context, routerID string, asn uint32) (string, error) {
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.router_id == @routerId AND node.asn == @asn
		LIMIT 1
		RETURN node._id
	`, pdp.db.config.IGPNode)

	bindVars := map[string]interface{}{
		"routerId": routerID,
		"asn":      asn,
	}

	cursor, err := pdp.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return "", err
	}
	defer cursor.Close()

	if cursor.HasMore() {
		var nodeID string
		if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
			return "", err
		}
		return nodeID, nil
	}

	return "", nil
}

func (pdp *PrefixDeduplicationProcessor) findBGPNodeByRouterIDAndASN(ctx context.Context, routerID string, asn uint32) (string, error) {
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.router_id == @routerId AND node.asn == @asn
		LIMIT 1
		RETURN node._id
	`, pdp.db.config.BGPNode)

	bindVars := map[string]interface{}{
		"routerId": routerID,
		"asn":      asn,
	}

	cursor, err := pdp.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return "", err
	}
	defer cursor.Close()

	if cursor.HasMore() {
		var nodeID string
		if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
			return "", err
		}
		return nodeID, nil
	}

	return "", nil
}

// extractKeyFromID extracts the key portion from an ArangoDB document ID
func extractKeyFromID(id string) string {
	// ArangoDB IDs are in format "collection/key"
	if idx := strings.LastIndex(id, "/"); idx != -1 {
		return id[idx+1:]
	}
	return id // fallback if no slash found
}

// handleExistingIGPEdges removes existing IGP edges pointing to ls_prefix and updates them to point to unified vertex
func (pdp *PrefixDeduplicationProcessor) handleExistingIGPEdges(ctx context.Context, existingEdge map[string]interface{}, unifiedKey string, isIPv4 bool) error {
	edgeKey := getString(existingEdge, "_key")
	fromNode := getString(existingEdge, "_from")
	oldToVertex := getString(existingEdge, "_to") // ls_prefix/xxx

	glog.V(8).Infof("Handling existing IGP edge: %s from %s to %s", edgeKey, fromNode, oldToVertex)

	// Determine target collections
	var targetCollection driver.Collection
	var newPrefixCollection string
	if isIPv4 {
		targetCollection = pdp.db.ipv4Graph
		newPrefixCollection = pdp.db.config.BGPPrefixV4
	} else {
		targetCollection = pdp.db.ipv6Graph
		newPrefixCollection = pdp.db.config.BGPPrefixV6
	}

	// Create new unified vertex ID
	newToVertex := fmt.Sprintf("%s/%s", newPrefixCollection, unifiedKey)

	// Remove the old IGP edge
	if _, err := targetCollection.RemoveDocument(ctx, edgeKey); err != nil {
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to remove old IGP edge %s: %w", edgeKey, err)
		}
	}

	// Create new edge pointing to unified vertex with preserved IGP metadata
	newEdge := &IPGraphObject{
		Key:            edgeKey, // Keep same key for consistency
		From:           fromNode,
		To:             newToVertex,
		Protocol:       "IGP_unified", // Mark as unified
		Link:           unifiedKey,
		ProtocolID:     existingEdge["protocol_id"],
		DomainID:       existingEdge["domain_id"],
		MTID:           uint16(getUint32FromInterface(existingEdge["mt_id"])),
		AreaID:         getString(existingEdge, "area_id"),
		LocalNodeASN:   getUint32FromInterface(existingEdge["local_node_asn"]),
		RemoteNodeASN:  getUint32FromInterface(existingEdge["remote_node_asn"]),
		IGPMetric:      getUint32FromInterface(existingEdge["igp_metric"]),
		Prefix:         getString(existingEdge, "prefix"),
		PrefixLen:      int32(getUint32FromInterface(existingEdge["prefix_len"])),
		PrefixMetric:   getUint32FromInterface(existingEdge["prefix_metric"]),
		PrefixAttrTLVs: existingEdge["prefix_attr_tlvs"],
	}

	// Create the updated edge
	if _, err := targetCollection.CreateDocument(ctx, newEdge); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create updated IGP edge %s: %w", edgeKey, err)
		}
		// Update if it already exists
		if _, err := targetCollection.UpdateDocument(ctx, edgeKey, newEdge); err != nil {
			return fmt.Errorf("failed to update IGP edge %s: %w", edgeKey, err)
		}
	}

	glog.V(8).Infof("Updated IGP edge %s to point to unified vertex %s", edgeKey, newToVertex)
	return nil
}

// removeDuplicateBGPEdges removes duplicate BGP edges for the same prefix
func (pdp *PrefixDeduplicationProcessor) removeDuplicateBGPEdges(ctx context.Context, unifiedKey string, isIPv4 bool) error {
	// Determine target collections
	var targetCollection driver.Collection
	var bgpPrefixCollection string
	if isIPv4 {
		targetCollection = pdp.db.ipv4Graph
		bgpPrefixCollection = pdp.db.config.BGPPrefixV4
	} else {
		targetCollection = pdp.db.ipv6Graph
		bgpPrefixCollection = pdp.db.config.BGPPrefixV6
	}

	bgpVertexID := fmt.Sprintf("%s/%s", bgpPrefixCollection, unifiedKey)

	// Find all edges pointing to or from this BGP prefix vertex
	// Keep only the one with the best metadata (most complete IGP data)
	query := fmt.Sprintf(`
		FOR edge IN %s
		FILTER edge._to == @bgp_vertex OR edge._from == @bgp_vertex
		FILTER edge.protocol == "IGP_unified" OR edge.protocol_id == null
		SORT edge.protocol_id DESC, edge.igp_metric DESC  // Prefer edges with IGP metadata
		RETURN edge
	`, targetCollection.Name())

	bindVars := map[string]interface{}{
		"bgp_vertex": bgpVertexID,
	}

	cursor, err := pdp.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to query duplicate BGP edges: %w", err)
	}
	defer cursor.Close()

	var edgesToRemove []string
	var bestEdgeKey string
	edgeCount := 0

	for cursor.HasMore() {
		var edge map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &edge); err != nil {
			return fmt.Errorf("failed to read duplicate edge: %w", err)
		}

		edgeKey := getString(edge, "_key")
		if edgeCount == 0 {
			// Keep the first (best) edge
			bestEdgeKey = edgeKey
		} else {
			// Mark others for removal
			edgesToRemove = append(edgesToRemove, edgeKey)
		}
		edgeCount++
	}

	// Remove duplicate edges
	for _, edgeKey := range edgesToRemove {
		if _, err := targetCollection.RemoveDocument(ctx, edgeKey); err != nil {
			if !driver.IsNotFound(err) {
				glog.Warningf("Failed to remove duplicate BGP edge %s: %v", edgeKey, err)
			}
		} else {
			glog.V(8).Infof("Removed duplicate BGP edge %s, keeping %s", edgeKey, bestEdgeKey)
		}
	}

	if len(edgesToRemove) > 0 {
		glog.V(7).Infof("Removed %d duplicate BGP edges for prefix %s, kept %s", len(edgesToRemove), unifiedKey, bestEdgeKey)
	}

	return nil
}
