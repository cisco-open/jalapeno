package arangodb

import (
	"context"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// processBGPPrefixUpdate processes BGP unicast prefix messages
func (uc *UpdateCoordinator) processBGPPrefixUpdate(msg *ProcessingMessage) error {
	glog.Infof("Processing BGP prefix update: %s action: %s", msg.Key, msg.Action)

	ctx := context.TODO()

	switch msg.Action {
	case "del":
		return uc.processPrefixWithdrawal(ctx, msg.Key, msg.Data)
	case "add", "update":
		return uc.processPrefixAdvertisement(ctx, msg.Key, msg.Data)
	default:
		glog.V(5).Infof("Unknown prefix action: %s for key: %s", msg.Action, msg.Key)
		return nil
	}
}

func (uc *UpdateCoordinator) processPrefixAdvertisement(ctx context.Context, key string, prefixData map[string]interface{}) error {
	// Extract prefix information
	prefix, _ := prefixData["prefix"].(string)
	prefixLen := getUint32FromInterface(prefixData["prefix_len"])
	originAS := getUint32FromInterface(prefixData["origin_as"])
	peerASN := getUint32FromInterface(prefixData["peer_asn"])
	isIPv4, _ := prefixData["is_ipv4"].(bool)

	if prefix == "" || originAS == 0 {
		return fmt.Errorf("invalid prefix data: missing prefix or origin_as")
	}

	// Use consistent key format (prefix_prefixlen) to match initial loading
	consistentKey := fmt.Sprintf("%s_%d", prefix, prefixLen)

	// Classify prefix type and determine processing strategy
	prefixType := uc.classifyBGPPrefix(prefix, prefixLen, originAS, peerASN, isIPv4)

	glog.V(8).Infof("Processing %s BGP prefix: %s/%d from AS%d via AS%d (key: %s)",
		prefixType, prefix, prefixLen, originAS, peerASN, consistentKey)

	// Determine if this should be node metadata or separate vertex
	if uc.shouldAttachAsNodeMetadata(prefixLen, isIPv4) {
		return uc.attachPrefixToOriginNode(ctx, prefix, prefixLen, originAS, prefixData, isIPv4)
	} else {
		return uc.createBGPPrefixVertex(ctx, consistentKey, prefixData, prefixType, isIPv4)
	}
}

func (uc *UpdateCoordinator) processPrefixWithdrawal(ctx context.Context, key string, prefixData map[string]interface{}) error {
	prefix, _ := prefixData["prefix"].(string)
	prefixLen := getUint32FromInterface(prefixData["prefix_len"])
	originAS := getUint32FromInterface(prefixData["origin_as"])
	isIPv4, _ := prefixData["is_ipv4"].(bool)

	// Use consistent key format (prefix_prefixlen) to match initial loading
	consistentKey := fmt.Sprintf("%s_%d", prefix, prefixLen)

	glog.Infof("Withdrawing BGP prefix: %s/%d from AS%d (BMP key: %s, consistent key: %s)", prefix, prefixLen, originAS, key, consistentKey)

	// Determine if this was node metadata or separate vertex
	if uc.shouldAttachAsNodeMetadata(prefixLen, isIPv4) {
		// For loopback prefixes (/32, /128) - remove from node metadata
		if originAS == 0 {
			glog.V(6).Infof("Origin AS missing for loopback withdrawal %s - skipping node metadata removal", consistentKey)
			return nil
		}
		return uc.removePrefixFromOriginNode(ctx, prefix, prefixLen, originAS, isIPv4)
	} else {
		// For transit prefixes - remove vertex and edges
		return uc.removeBGPPrefixVertex(ctx, consistentKey, prefixData, isIPv4)
	}
}

func (uc *UpdateCoordinator) classifyBGPPrefix(prefix string, prefixLen, originAS, peerASN uint32, isIPv4 bool) string {
	// Classify based on AS characteristics
	isOriginPrivate := uc.isPrivateASN(originAS)
	isPeerPrivate := uc.isPrivateASN(peerASN)

	if originAS == peerASN {
		return "ibgp"
	} else if isOriginPrivate && isPeerPrivate {
		return "ebgp_private"
	} else if !isOriginPrivate && !isPeerPrivate {
		return "ebgp_public"
	} else {
		// Mixed private/public - typically customer/provider relationships
		return "ebgp_hybrid"
	}
}

func (uc *UpdateCoordinator) isPrivateASN(asn uint32) bool {
	return (asn >= 64512 && asn <= 65535) || (asn >= 4200000000 && asn <= 4294967294)
}

func (uc *UpdateCoordinator) shouldAttachAsNodeMetadata(prefixLen uint32, isIPv4 bool) bool {
	// Following IGP-graph pattern: /32 (IPv4) and /128 (IPv6) loopbacks as node metadata
	if isIPv4 && prefixLen == 32 {
		return true
	}
	if !isIPv4 && prefixLen == 128 {
		return true
	}
	return false
}

func (uc *UpdateCoordinator) attachPrefixToOriginNode(ctx context.Context, prefix string, prefixLen, originAS uint32, prefixData map[string]interface{}, isIPv4 bool) error {
	// Find the origin node (could be IGP node or BGP node)
	originNodeID, err := uc.findOriginNode(ctx, originAS, prefixData)
	if err != nil {
		return fmt.Errorf("failed to find origin node for AS%d: %w", originAS, err)
	}

	if originNodeID == "" {
		// Origin node doesn't exist - create BGP node for this AS
		originNodeID, err = uc.createOriginBGPNode(ctx, originAS, prefixData)
		if err != nil {
			return fmt.Errorf("failed to create origin BGP node for AS%d: %w", originAS, err)
		}
	}

	// Add prefix to node's metadata
	return uc.addPrefixToNodeMetadata(ctx, originNodeID, prefix, prefixLen, prefixData)
}

func (uc *UpdateCoordinator) findOriginNode(ctx context.Context, originAS uint32, prefixData map[string]interface{}) (string, error) {
	// Strategy 1: Look for IGP nodes with matching ASN
	igpNodeID, err := uc.findIGPNodeByASN(ctx, originAS)
	if err != nil {
		return "", err
	}
	if igpNodeID != "" {
		return igpNodeID, nil
	}

	// Strategy 2: Look for existing BGP peer nodes with matching ASN
	bgpNodeID, err := uc.findBGPNodeByASN(ctx, originAS)
	if err != nil {
		return "", err
	}
	if bgpNodeID != "" {
		return bgpNodeID, nil
	}

	// Strategy 3: No existing node found - will need to create one
	return "", nil // Indicates node doesn't exist yet
}

func (uc *UpdateCoordinator) findIGPNodeByASN(ctx context.Context, asn uint32) (string, error) {
	// Query IGP nodes for matching ASN
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.asn == @asn
		LIMIT 1
		RETURN node._id
	`, uc.db.config.IGPNode)

	bindVars := map[string]interface{}{
		"asn": asn,
	}

	cursor, err := uc.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return "", fmt.Errorf("failed to query IGP nodes by ASN: %w", err)
	}
	defer cursor.Close()

	if cursor.HasMore() {
		var nodeID string
		if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
			return "", fmt.Errorf("failed to read IGP node ID: %w", err)
		}
		return nodeID, nil
	}

	return "", nil
}

func (uc *UpdateCoordinator) findBGPNodeByASN(ctx context.Context, asn uint32) (string, error) {
	// Query BGP nodes for matching ASN (prefer real peer nodes over origin nodes)
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.asn == @asn
		FILTER node.node_type == "bgp"  // Prefer real BGP peer nodes
		LIMIT 1
		RETURN node._id
	`, uc.db.config.BGPNode)

	bindVars := map[string]interface{}{
		"asn": asn,
	}

	cursor, err := uc.db.db.Query(ctx, query, bindVars)
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

func (uc *UpdateCoordinator) createOriginBGPNode(ctx context.Context, originAS uint32, prefixData map[string]interface{}) (string, error) {
	// Create a representative BGP node for this AS
	// Use origin AS as the router ID since we don't have specific router info
	bgpNodeKey := fmt.Sprintf("bgp_%d_origin", originAS)
	routerID := fmt.Sprintf("origin_as_%d", originAS)

	bgpNode := &BGPNode{
		Key:         bgpNodeKey,
		RouterID:    routerID,
		BGPRouterID: routerID,
		ASN:         originAS,
		NodeType:    "bgp_origin",
		Tier:        uc.determineBGPNodeTier(originAS),
	}

	// Create BGP node
	if _, err := uc.db.bgpNode.CreateDocument(ctx, bgpNode); err != nil {
		if !driver.IsConflict(err) {
			return "", fmt.Errorf("failed to create origin BGP node: %w", err)
		}
		// Node already exists, which is fine
	}

	nodeID := fmt.Sprintf("%s/%s", uc.db.config.BGPNode, bgpNodeKey)
	glog.V(8).Infof("Created origin BGP node: %s for AS%d", nodeID, originAS)
	return nodeID, nil
}

func (uc *UpdateCoordinator) addPrefixToNodeMetadata(ctx context.Context, nodeID, prefix string, prefixLen uint32, prefixData map[string]interface{}) error {
	// Create prefix metadata object
	prefixMetadata := map[string]interface{}{
		"prefix":     prefix,
		"prefix_len": prefixLen,
		"origin_as":  getUint32FromInterface(prefixData["origin_as"]),
		"peer_asn":   getUint32FromInterface(prefixData["peer_asn"]),
		"nexthop":    getStringFromData(prefixData, "nexthop"),
		"timestamp":  getStringFromData(prefixData, "timestamp"),
	}

	// Add AS path if available
	if asPath, ok := prefixData["base_attrs"].(map[string]interface{}); ok {
		if path, ok := asPath["as_path"].([]interface{}); ok {
			prefixMetadata["as_path"] = path
		}
	}

	// Update node to add prefix to metadata
	// This requires reading the node, updating the prefixes array, and writing it back
	updateQuery := fmt.Sprintf(`
		FOR node IN %s
		FILTER node._id == @nodeId
		LET currentPrefixes = node.prefixes || []
		LET newPrefixes = APPEND(currentPrefixes, @prefixData)
		UPDATE node WITH { prefixes: newPrefixes } IN %s
		RETURN NEW
	`, uc.getNodeCollectionFromID(nodeID), uc.getNodeCollectionFromID(nodeID))

	bindVars := map[string]interface{}{
		"nodeId":     nodeID,
		"prefixData": prefixMetadata,
	}

	cursor, err := uc.db.db.Query(ctx, updateQuery, bindVars)
	if err != nil {
		return fmt.Errorf("failed to add prefix to node metadata: %w", err)
	}
	defer cursor.Close()

	glog.V(8).Infof("Added prefix %s/%d to node %s metadata", prefix, prefixLen, nodeID)
	return nil
}

func (uc *UpdateCoordinator) getNodeCollectionFromID(nodeID string) string {
	// Extract collection name from node ID (format: "collection/key")
	if len(nodeID) > 0 {
		for i, char := range nodeID {
			if char == '/' {
				return nodeID[:i]
			}
		}
	}
	return uc.db.config.IGPNode // Default fallback
}

func (uc *UpdateCoordinator) createBGPPrefixVertex(ctx context.Context, key string, prefixData map[string]interface{}, prefixType string, isIPv4 bool) error {
	prefix, _ := prefixData["prefix"].(string)
	prefixLen := getUint32FromInterface(prefixData["prefix_len"])
	originAS := getUint32FromInterface(prefixData["origin_as"])
	peerASN := getUint32FromInterface(prefixData["peer_asn"])

	// Create BGP prefix vertex
	bgpPrefix := &BGPPrefix{
		Key:        key,
		Prefix:     prefix,
		PrefixLen:  int32(prefixLen),
		OriginAS:   int32(originAS),
		PeerASN:    peerASN,
		PrefixType: prefixType,
		Nexthop:    getStringFromData(prefixData, "nexthop"),
		IsHost:     uc.shouldAttachAsNodeMetadata(prefixLen, isIPv4),
	}

	// Add base attributes if available
	if baseAttrs, ok := prefixData["base_attrs"].(map[string]interface{}); ok {
		// Convert to BGP base attributes if needed
		// For now, we'll store the raw data - could be enhanced later
		_ = baseAttrs // Placeholder for base attributes processing
	}

	// Determine target collection based on prefix type
	var targetCollection driver.Collection
	if isIPv4 {
		targetCollection = uc.db.bgpPrefixV4
	} else {
		targetCollection = uc.db.bgpPrefixV6
	}

	// Create prefix vertex
	if _, err := targetCollection.CreateDocument(ctx, bgpPrefix); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create BGP prefix vertex: %w", err)
		}
		// Update existing prefix
		if _, err := targetCollection.UpdateDocument(ctx, key, bgpPrefix); err != nil {
			return fmt.Errorf("failed to update BGP prefix vertex: %w", err)
		}
	}

	// Create edge from origin node to prefix vertex
	if err := uc.createPrefixToOriginEdge(ctx, key, prefixData, isIPv4); err != nil {
		return fmt.Errorf("failed to create prefix-to-origin edge: %w", err)
	}

	glog.V(8).Infof("Created BGP prefix vertex: %s/%d (type: %s)", prefix, prefixLen, prefixType)
	return nil
}

func (uc *UpdateCoordinator) createPrefixToOriginEdge(ctx context.Context, prefixKey string, prefixData map[string]interface{}, isIPv4 bool) error {
	originAS := getUint32FromInterface(prefixData["origin_as"])
	peerASN := getUint32FromInterface(prefixData["peer_asn"])

	// For internet prefixes, attach to BGP peer nodes that advertise them
	// rather than creating artificial origin nodes
	peerNodeIDs, err := uc.findBGPPeerNodesForPrefix(ctx, originAS, peerASN, prefixData)
	if err != nil {
		return fmt.Errorf("failed to find BGP peer nodes: %w", err)
	}

	if len(peerNodeIDs) == 0 {
		glog.V(6).Infof("No BGP peer nodes found for prefix %s from AS%d via AS%d - skipping edge creation", prefixKey, originAS, peerASN)
		return nil // Don't create artificial origin nodes
	}

	// Determine target prefix collection
	var prefixCollection string
	if isIPv4 {
		prefixCollection = uc.db.config.BGPPrefixV4
	} else {
		prefixCollection = uc.db.config.BGPPrefixV6
	}

	// Create bidirectional edges between all peer nodes and prefix vertex
	for _, peerNodeID := range peerNodeIDs {
		// Get the peer node data
		peerNodeData, err := uc.getPeerNodeData(ctx, peerNodeID)
		if err != nil {
			glog.Warningf("Failed to get peer node data for %s: %v", peerNodeID, err)
			continue
		}

		if err := uc.db.createBidirectionalPrefixEdges(ctx, prefixData, peerNodeData, prefixCollection, isIPv4); err != nil {
			glog.Warningf("Failed to create bidirectional prefix edges for peer %s: %v", peerNodeID, err)
			continue
		}
		glog.V(8).Infof("Created prefix edges: %s ↔ %s", peerNodeID, fmt.Sprintf("%s/%s", prefixCollection, prefixKey))
	}

	glog.V(8).Infof("Created prefix-to-peer edges for: %s (%d peers)", prefixKey, len(peerNodeIDs))
	return nil
}

func (uc *UpdateCoordinator) removePrefixFromOriginNode(ctx context.Context, prefix string, prefixLen, originAS uint32, isIPv4 bool) error {
	// Find the origin node
	originNodeID, err := uc.findOriginNode(ctx, originAS, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to find origin node for prefix removal: %w", err)
	}

	if originNodeID == "" {
		glog.V(6).Infof("Origin node not found for AS%d during prefix removal", originAS)
		return nil // Node doesn't exist, nothing to remove
	}

	// Remove prefix from node's metadata
	removeQuery := fmt.Sprintf(`
		FOR node IN %s
		FILTER node._id == @nodeId
		LET currentPrefixes = node.prefixes || []
		LET filteredPrefixes = (
			FOR p IN currentPrefixes
			FILTER NOT (p.prefix == @prefix AND p.prefix_len == @prefixLen)
			RETURN p
		)
		UPDATE node WITH { prefixes: filteredPrefixes } IN %s
		RETURN NEW
	`, uc.getNodeCollectionFromID(originNodeID), uc.getNodeCollectionFromID(originNodeID))

	bindVars := map[string]interface{}{
		"nodeId":    originNodeID,
		"prefix":    prefix,
		"prefixLen": prefixLen,
	}

	cursor, err := uc.db.db.Query(ctx, removeQuery, bindVars)
	if err != nil {
		return fmt.Errorf("failed to remove prefix from node metadata: %w", err)
	}
	defer cursor.Close()

	glog.V(8).Infof("Removed prefix %s/%d from node %s metadata", prefix, prefixLen, originNodeID)
	return nil
}

func (uc *UpdateCoordinator) removeBGPPrefixVertex(ctx context.Context, key string, prefixData map[string]interface{}, isIPv4 bool) error {
	// Determine target collections
	var prefixCollection driver.Collection
	var graphCollection driver.Collection
	var prefixCollectionName string

	if isIPv4 {
		prefixCollection = uc.db.bgpPrefixV4
		graphCollection = uc.db.ipv4Graph
		prefixCollectionName = uc.db.config.BGPPrefixV4
	} else {
		prefixCollection = uc.db.bgpPrefixV6
		graphCollection = uc.db.ipv6Graph
		prefixCollectionName = uc.db.config.BGPPrefixV6
	}

	prefixVertexID := fmt.Sprintf("%s/%s", prefixCollectionName, key)

	// Find and remove all edges connected to this prefix vertex
	edgeQuery := fmt.Sprintf(`
		FOR edge IN %s
		FILTER edge._to == @prefixVertex OR edge._from == @prefixVertex
		RETURN edge._key
	`, graphCollection.Name())

	bindVars := map[string]interface{}{
		"prefixVertex": prefixVertexID,
	}

	cursor, err := uc.db.db.Query(ctx, edgeQuery, bindVars)
	if err != nil {
		glog.Warningf("Failed to query prefix edges for %s: %v", key, err)
	} else {
		defer cursor.Close()

		edgeCount := 0
		for cursor.HasMore() {
			var edgeKey string
			if _, err := cursor.ReadDocument(ctx, &edgeKey); err != nil {
				continue
			}

			if _, err := graphCollection.RemoveDocument(ctx, edgeKey); err != nil {
				if !driver.IsNotFoundGeneral(err) {
					glog.V(6).Infof("Failed to remove prefix edge %s: %v", edgeKey, err)
				}
			} else {
				edgeCount++
			}
		}
		glog.V(7).Infof("Removed %d edges for prefix %s", edgeCount, key)
	}

	// Remove prefix vertex
	if _, err := prefixCollection.RemoveDocument(ctx, key); err != nil {
		if !driver.IsNotFoundGeneral(err) {
			glog.Warningf("Failed to remove BGP prefix vertex %s: %v", key, err)
		}
	}

	glog.V(8).Infof("Removed BGP prefix vertex and edges: %s", key)
	return nil
}

// findBGPPeerNodesForPrefix finds BGP peer nodes that should advertise this prefix
func (uc *UpdateCoordinator) findBGPPeerNodesForPrefix(ctx context.Context, originAS, peerASN uint32, prefixData map[string]interface{}) ([]string, error) {
	var peerNodeIDs []string

	// Get prefix type to determine attachment strategy
	prefix := getStringFromData(prefixData, "prefix")
	prefixLen := getUint32FromInterface(prefixData["prefix_len"])
	prefixType := uc.classifyBGPPrefix(prefix, prefixLen, originAS, peerASN, true)

	glog.V(7).Infof("Finding BGP peers for prefix %s/%d (type: %s, origin AS: %d, peer ASN: %d)", prefix, prefixLen, prefixType, originAS, peerASN)

	switch prefixType {
	case "ebgp_public":
		// Internet prefixes - attach to all public BGP peers
		query := fmt.Sprintf(`
			FOR node IN %s
			FILTER node.node_type == "bgp"
			FILTER node.tier == "tier1" OR node.tier == "tier2" OR node.tier == "tier3"
			RETURN node._id
		`, uc.db.config.BGPNode)

		cursor, err := uc.db.db.Query(ctx, query, nil)
		if err != nil {
			return nil, err
		}
		defer cursor.Close()

		for cursor.HasMore() {
			var nodeID string
			if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
				continue
			}
			peerNodeIDs = append(peerNodeIDs, nodeID)
		}

	case "ebgp_hybrid":
		// Hybrid prefixes - attach to the specific peer that advertised it
		peerIP := getStringFromData(prefixData, "peer_ip")
		query := fmt.Sprintf(`
			FOR node IN %s
			FILTER node.node_type == "bgp"
			FILTER node.local_ip == @peer_ip OR node.router_id == @peer_ip
			RETURN node._id
		`, uc.db.config.BGPNode)

		bindVars := map[string]interface{}{
			"peer_ip": peerIP,
		}

		cursor, err := uc.db.db.Query(ctx, query, bindVars)
		if err != nil {
			return nil, err
		}
		defer cursor.Close()

		for cursor.HasMore() {
			var nodeID string
			if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
				continue
			}
			peerNodeIDs = append(peerNodeIDs, nodeID)
		}

	case "ebgp_private":
		// Private eBGP prefixes - attach to specific BGP node with matching ASN
		query := fmt.Sprintf(`
			FOR node IN %s
			FILTER node.asn == @origin_asn
			FILTER node.node_type == "bgp"
			RETURN node._id
		`, uc.db.config.BGPNode)

		bindVars := map[string]interface{}{
			"origin_asn": originAS,
		}

		cursor, err := uc.db.db.Query(ctx, query, bindVars)
		if err != nil {
			return nil, err
		}
		defer cursor.Close()

		for cursor.HasMore() {
			var nodeID string
			if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
				continue
			}
			peerNodeIDs = append(peerNodeIDs, nodeID)
		}

	case "ibgp":
		// iBGP prefixes - attach to IGP nodes or BGP nodes with matching ASN
		query := fmt.Sprintf(`
			FOR node IN %s
			FILTER node.asn == @asn
			RETURN node._id
		`, uc.db.config.IGPNode)

		bindVars := map[string]interface{}{
			"asn": originAS,
		}

		cursor, err := uc.db.db.Query(ctx, query, bindVars)
		if err != nil {
			return nil, err
		}
		defer cursor.Close()

		for cursor.HasMore() {
			var nodeID string
			if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
				continue
			}
			peerNodeIDs = append(peerNodeIDs, nodeID)
		}

		// If no IGP nodes found, look for BGP nodes
		if len(peerNodeIDs) == 0 {
			query := fmt.Sprintf(`
				FOR node IN %s
				FILTER node.asn == @asn
				FILTER node.node_type == "bgp"
				RETURN node._id
			`, uc.db.config.BGPNode)

			cursor, err := uc.db.db.Query(ctx, query, bindVars)
			if err != nil {
				return nil, err
			}
			defer cursor.Close()

			for cursor.HasMore() {
				var nodeID string
				if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
					continue
				}
				peerNodeIDs = append(peerNodeIDs, nodeID)
			}
		}
	}

	glog.V(7).Infof("Found %d BGP peer nodes for prefix %s/%d (type: %s)", len(peerNodeIDs), prefix, prefixLen, prefixType)
	return peerNodeIDs, nil
}

// getPeerNodeData retrieves the full node document for a given node ID
func (uc *UpdateCoordinator) getPeerNodeData(ctx context.Context, nodeID string) (map[string]interface{}, error) {
	// Extract collection name from node ID (format: "collection/key")
	collectionName := "igp_node" // Default
	if idx := strings.Index(nodeID, "/"); idx != -1 {
		collectionName = nodeID[:idx]
	}

	// Query for the node document
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node._id == @nodeId
		RETURN node
	`, collectionName)

	bindVars := map[string]interface{}{
		"nodeId": nodeID,
	}

	cursor, err := uc.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	if cursor.HasMore() {
		var nodeData map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &nodeData); err != nil {
			return nil, err
		}
		return nodeData, nil
	}

	return nil, fmt.Errorf("node not found: %s", nodeID)
}

// Helper function to safely get string from interface (BGP prefix processor)
func getStringFromData(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}
