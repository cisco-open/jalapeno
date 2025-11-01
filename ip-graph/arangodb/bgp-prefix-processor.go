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

	// BGP Best Path Selection: Check if we should update existing prefix
	if shouldUpdate, err := uc.shouldUpdateBGPPrefix(ctx, consistentKey, prefixData, isIPv4); err != nil {
		return fmt.Errorf("failed to check BGP best path: %w", err)
	} else if !shouldUpdate {
		glog.V(7).Infof("BGP prefix %s/%d - existing path is better, skipping update", prefix, prefixLen)
		return nil
	}

	// Classify prefix type and determine processing strategy
	prefixType := uc.classifyBGPPrefix(prefix, prefixLen, originAS, peerASN, isIPv4)

	glog.Infof("Processing %s BGP prefix: %s/%d from AS%d via AS%d (key: %s)",
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
	// BMP peer-centric approach: Always attach to the advertising peer
	// Classification doesn't matter much - we always connect to the specific peer
	return "ebgp_peer_centric"
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
		Key:      bgpNodeKey,
		RouterID: routerID,
		ASN:      originAS,
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
	case "ebgp_peer_centric":
		// All prefixes (both private and public ASN) connect to the BGP peer that advertised them
		// This matches the working initial loading logic
		return uc.findAdvertisingBGPPeer(ctx, prefix, prefixLen, originAS, peerASN, prefixData)
	}

	glog.V(7).Infof("Found %d BGP peer nodes for prefix %s/%d (type: %s)", len(peerNodeIDs), prefix, prefixLen, prefixType)

	// Enhanced debug logging for missing prefixes
	if len(peerNodeIDs) == 0 {
		peerIP := getStringFromData(prefixData, "peer_ip")
		glog.Warningf("No BGP peer nodes found for prefix %s/%d (type: %s, origin AS: %d, peer ASN: %d, peer IP: %s)",
			prefix, prefixLen, prefixType, originAS, peerASN, peerIP)

		// Debug: Check what BGP nodes exist for this ASN
		debugQuery := fmt.Sprintf(`
				FOR node IN %s
				FILTER node.asn == @peer_asn
				RETURN {_id: node._id, router_id: node.router_id, asn: node.asn, tier: node.tier}
			`, uc.db.config.BGPNode)

		debugBindVars := map[string]interface{}{
			"peer_asn": peerASN,
		}

		if debugCursor, err := uc.db.db.Query(ctx, debugQuery, debugBindVars); err == nil {
			defer debugCursor.Close()
			glog.Warningf("Available BGP nodes for ASN %d:", peerASN)
			for debugCursor.HasMore() {
				var nodeInfo map[string]interface{}
				if _, err := debugCursor.ReadDocument(ctx, &nodeInfo); err == nil {
					glog.Warningf("  - Node: %v", nodeInfo)
				}
			}
		}
	}

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

// findPublicBGPPeers finds specific public BGP peers that advertised this internet prefix (based on BMP data)
func (uc *UpdateCoordinator) findPublicBGPPeers(ctx context.Context, prefix string, prefixLen uint32, originAS, peerASN uint32) ([]string, error) {
	var peerNodeIDs []string

	glog.Infof("DEBUG: Internet prefix %s/%d from AS%d - finding all BGP peers that advertised it", prefix, prefixLen, originAS)

	// Find all BMP unicast_prefix_v4 entries for this prefix to get all advertising peers
	unicastQuery := fmt.Sprintf(`
		FOR u IN unicast_prefix_v4
		FILTER u.prefix == @prefix AND u.prefix_len == @prefix_len AND u.origin_as == @origin_as
		FOR p IN peer
		FILTER u.peer_ip == p.remote_ip AND u.peer_asn == p.remote_asn
		FOR bgp IN %s
		FILTER bgp.router_id == p.remote_bgp_id AND bgp.asn == p.remote_asn
		RETURN DISTINCT {_id: bgp._id, router_id: bgp.router_id, asn: bgp.asn, peer_ip: u.peer_ip}
	`, uc.db.config.BGPNode)

	bindVars := map[string]interface{}{
		"prefix":     prefix,
		"prefix_len": prefixLen,
		"origin_as":  originAS,
	}

	cursor, err := uc.db.db.Query(ctx, unicastQuery, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query BGP peers that advertised prefix: %w", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var nodeInfo map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &nodeInfo); err != nil {
			continue
		}
		nodeID := nodeInfo["_id"].(string)
		peerNodeIDs = append(peerNodeIDs, nodeID)
		glog.Infof("DEBUG: Found BGP peer that advertised internet prefix: %v", nodeInfo)
	}

	glog.Infof("DEBUG: Found %d BGP peers that advertised internet prefix %s/%d", len(peerNodeIDs), prefix, prefixLen)
	return peerNodeIDs, nil
}

// findSpecificBGPPeer finds a specific BGP peer for private ASN prefixes
func (uc *UpdateCoordinator) findSpecificBGPPeer(ctx context.Context, prefix string, prefixLen uint32, originAS, peerASN uint32, prefixData map[string]interface{}) ([]string, error) {
	var peerNodeIDs []string
	peerIP := getStringFromData(prefixData, "peer_ip")

	glog.Infof("DEBUG: Private ASN prefix %s/%d - finding specific peer (peer_ip: %s, peer_asn: %d, origin_as: %d)", prefix, prefixLen, peerIP, peerASN, originAS)

	// First, find the peer session to get the remote_bgp_id
	peerQuery := `
		FOR p IN peer
		FILTER p.remote_ip == @peer_ip AND p.remote_asn == @peer_asn
		RETURN {remote_bgp_id: p.remote_bgp_id, remote_asn: p.remote_asn}
	`

	peerBindVars := map[string]interface{}{
		"peer_ip":  peerIP,
		"peer_asn": peerASN,
	}

	peerCursor, err := uc.db.db.Query(ctx, peerQuery, peerBindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query peer table: %w", err)
	}
	defer peerCursor.Close()

	var remoteBGPID string
	if peerCursor.HasMore() {
		var peerInfo map[string]interface{}
		if _, err := peerCursor.ReadDocument(ctx, &peerInfo); err == nil {
			if bgpID, ok := peerInfo["remote_bgp_id"].(string); ok {
				remoteBGPID = bgpID
			}
		}
	}

	if remoteBGPID == "" {
		glog.Warningf("No peer session found for prefix %s/%d (peer_ip: %s, peer_asn: %d)", prefix, prefixLen, peerIP, peerASN)
		return peerNodeIDs, nil
	}

	// Now find the BGP node using the correct router_id and peer ASN
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.router_id == @router_id AND node.asn == @peer_asn
		RETURN {_id: node._id, router_id: node.router_id, asn: node.asn}
	`, uc.db.config.BGPNode)

	bindVars := map[string]interface{}{
		"router_id": remoteBGPID,
		"peer_asn":  peerASN,
	}

	cursor, err := uc.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query specific BGP peer: %w", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var nodeInfo map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &nodeInfo); err != nil {
			continue
		}
		nodeID := nodeInfo["_id"].(string)
		peerNodeIDs = append(peerNodeIDs, nodeID)
		glog.Infof("DEBUG: Found specific BGP peer: %v", nodeInfo)
	}

	glog.Infof("DEBUG: Found %d specific BGP peers for private prefix %s/%d", len(peerNodeIDs), prefix, prefixLen)
	return peerNodeIDs, nil
}

// findAdvertisingBGPPeer finds the BGP peer that advertised this prefix (replicating working initial loading logic)
func (uc *UpdateCoordinator) findAdvertisingBGPPeer(ctx context.Context, prefix string, prefixLen uint32, originAS, peerASN uint32, prefixData map[string]interface{}) ([]string, error) {
	var peerNodeIDs []string
	peerIP := getStringFromData(prefixData, "peer_ip")

	glog.Infof("DEBUG: Finding BGP peer that advertised prefix %s/%d (peer_ip: %s, peer_asn: %d, origin_as: %d)",
		prefix, prefixLen, peerIP, peerASN, originAS)

	// Replicate the working initial loading logic:
	// 1. Find the peer session to get remote_bgp_id
	// 2. Find BGP node with router_id == remote_bgp_id AND asn == peer_asn
	// This matches your working example: bgp_node/10.109.9.1_100009

	peerQuery := `
		FOR p IN peer
		FILTER p.remote_ip == @peer_ip AND p.remote_asn == @peer_asn
		RETURN {remote_bgp_id: p.remote_bgp_id, remote_asn: p.remote_asn}
	`

	peerBindVars := map[string]interface{}{
		"peer_ip":  peerIP,
		"peer_asn": peerASN,
	}

	peerCursor, err := uc.db.db.Query(ctx, peerQuery, peerBindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query peer table: %w", err)
	}
	defer peerCursor.Close()

	var remoteBGPID string
	if peerCursor.HasMore() {
		var peerInfo map[string]interface{}
		if _, err := peerCursor.ReadDocument(ctx, &peerInfo); err == nil {
			if bgpID, ok := peerInfo["remote_bgp_id"].(string); ok {
				remoteBGPID = bgpID
			}
		}
	}

	if remoteBGPID == "" {
		glog.Warningf("No peer session found for prefix %s/%d (peer_ip: %s, peer_asn: %d)", prefix, prefixLen, peerIP, peerASN)
		return peerNodeIDs, nil
	}

	glog.Infof("DEBUG: Found peer session - looking for BGP node with router_id == %s AND asn == %d", remoteBGPID, peerASN)

	// Find the BGP node that matches this peer session
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.router_id == @router_id AND node.asn == @peer_asn
		RETURN {_id: node._id, router_id: node.router_id, asn: node.asn}
	`, uc.db.config.BGPNode)

	bindVars := map[string]interface{}{
		"router_id": remoteBGPID,
		"peer_asn":  peerASN,
	}

	cursor, err := uc.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query advertising BGP peer: %w", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var nodeInfo map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &nodeInfo); err != nil {
			continue
		}
		nodeID := nodeInfo["_id"].(string)
		peerNodeIDs = append(peerNodeIDs, nodeID)
		glog.Infof("DEBUG: Found advertising BGP peer: %v", nodeInfo)
	}

	glog.Infof("DEBUG: Found %d BGP peers for prefix %s/%d", len(peerNodeIDs), prefix, prefixLen)
	return peerNodeIDs, nil
}

// shouldUpdateBGPPrefix implements BGP best path selection
func (uc *UpdateCoordinator) shouldUpdateBGPPrefix(ctx context.Context, prefixKey string, newPrefixData map[string]interface{}, isIPv4 bool) (bool, error) {
	// Determine target collection
	var collection driver.Collection
	if isIPv4 {
		collection = uc.db.bgpPrefixV4
	} else {
		collection = uc.db.bgpPrefixV6
	}

	// Check if prefix already exists
	var existingPrefix BGPPrefix
	_, err := collection.ReadDocument(ctx, prefixKey, &existingPrefix)
	if driver.IsNotFound(err) {
		// No existing prefix - accept new one
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to read existing prefix: %w", err)
	}

	// Extract BGP attributes from new prefix data
	newASPath := getBGPASPath(newPrefixData)
	newLocalPref := getBGPLocalPref(newPrefixData)
	newASPathCount := len(newASPath)

	// Get existing prefix attributes (we need to query the original BMP data)
	existingASPathCount := uc.getStoredASPathCount(existingPrefix)
	existingIsIBGP := uc.isIBGPPrefix(existingPrefix)
	newIsIBGP := (newLocalPref > 0) // iBGP has local_pref

	glog.V(8).Infof("BGP best path comparison for %s: existing(ibgp=%v, path_len=%d) vs new(ibgp=%v, path_len=%d)",
		prefixKey, existingIsIBGP, existingASPathCount, newIsIBGP, newASPathCount)

	// BGP Best Path Selection Rules:
	// 1. eBGP > iBGP
	if existingIsIBGP && !newIsIBGP {
		glog.V(7).Infof("BGP best path: eBGP beats iBGP for %s", prefixKey)
		return true, nil
	}
	if !existingIsIBGP && newIsIBGP {
		glog.V(7).Infof("BGP best path: existing eBGP beats new iBGP for %s", prefixKey)
		return false, nil
	}

	// 2. Shorter AS path wins (if same eBGP/iBGP type)
	if newASPathCount < existingASPathCount {
		glog.V(7).Infof("BGP best path: shorter AS path (%d < %d) for %s", newASPathCount, existingASPathCount, prefixKey)
		return true, nil
	}
	if newASPathCount > existingASPathCount {
		glog.V(7).Infof("BGP best path: existing shorter AS path (%d < %d) for %s", existingASPathCount, newASPathCount, prefixKey)
		return false, nil
	}

	// 3. If tie, keep existing (first wins)
	glog.V(7).Infof("BGP best path: tie - keeping existing path for %s", prefixKey)
	return false, nil
}

// Helper functions for BGP best path selection
func getBGPASPath(prefixData map[string]interface{}) []interface{} {
	if baseAttrs, ok := prefixData["base_attrs"].(map[string]interface{}); ok {
		if asPath, ok := baseAttrs["as_path"].([]interface{}); ok {
			return asPath
		}
	}
	return []interface{}{}
}

func getBGPLocalPref(prefixData map[string]interface{}) uint32 {
	if baseAttrs, ok := prefixData["base_attrs"].(map[string]interface{}); ok {
		if localPref, ok := baseAttrs["local_pref"].(float64); ok {
			return uint32(localPref)
		}
	}
	return 0
}

func (uc *UpdateCoordinator) getStoredASPathCount(prefix BGPPrefix) int {
	// For now, estimate based on prefix_type
	// TODO: Store AS path count in BGPPrefix struct for accurate comparison
	switch prefix.PrefixType {
	case "ibgp":
		return 1 // Typically internal
	case "ebgp_private", "ebgp_private_4byte":
		return 2 // Typically 1-2 hops
	case "ebgp_public":
		return 3 // Typically longer paths
	default:
		return 2 // Default estimate
	}
}

func (uc *UpdateCoordinator) isIBGPPrefix(prefix BGPPrefix) bool {
	return prefix.PrefixType == "ibgp"
}

// Helper function to safely get string from interface (BGP prefix processor)
func getStringFromData(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}
