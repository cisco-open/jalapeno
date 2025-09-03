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

	// Find IPv4 prefix conflicts between ls_prefix and bgp_prefix_v4
	conflictQuery := `
		FOR ls IN ls_prefix
		FILTER ls.mt_id_tlv == null OR ls.mt_id_tlv.mt_id == 0  // IPv4 prefixes
		FOR bgp IN ` + pdp.db.config.BGPPrefixV4 + `
		FILTER ls.prefix == bgp.prefix AND ls.prefix_len == bgp.prefix_len
		RETURN {
			prefix: ls.prefix,
			prefix_len: ls.prefix_len,
			ls_data: ls,
			bgp_data: bgp,
			unified_key: CONCAT_SEPARATOR("_", ls.prefix, ls.prefix_len)
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

	// Find IPv6 prefix conflicts between ls_prefix and bgp_prefix_v6
	conflictQuery := `
		FOR ls IN ls_prefix
		FILTER ls.mt_id_tlv != null AND ls.mt_id_tlv.mt_id == 2  // IPv6 prefixes
		FOR bgp IN ` + pdp.db.config.BGPPrefixV6 + `
		FILTER ls.prefix == bgp.prefix AND ls.prefix_len == bgp.prefix_len
		RETURN {
			prefix: ls.prefix,
			prefix_len: ls.prefix_len,
			ls_data: ls,
			bgp_data: bgp,
			unified_key: CONCAT_SEPARATOR("_", ls.prefix, ls.prefix_len)
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

	// Extract IGP and BGP data
	lsData := conflict["ls_data"].(map[string]interface{})
	bgpData := conflict["bgp_data"].(map[string]interface{})

	glog.V(7).Infof("Creating unified prefix vertex for %s/%d", prefix, prefixLen)

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

	// 1. Create edge to IGP node (based on ls_prefix data)
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

	// 2. Create edge to BGP node (based on bgp_prefix data)
	bgpRouterID := getString(bgpData, "router_id")
	bgpOriginAS := getUint32FromInterface(bgpData["origin_as"])
	bgpPrefixType := getString(bgpData, "prefix_type")

	var bgpNodeID string
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
