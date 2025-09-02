package arangodb

import (
	"context"
	"fmt"

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

	// Classify prefix type and determine processing strategy
	prefixType := uc.classifyBGPPrefix(prefix, prefixLen, originAS, peerASN, isIPv4)

	glog.V(8).Infof("Processing %s BGP prefix: %s/%d from AS%d via AS%d",
		prefixType, prefix, prefixLen, originAS, peerASN)

	// Determine if this should be node metadata or separate vertex
	if uc.shouldAttachAsNodeMetadata(prefixLen, isIPv4) {
		return uc.attachPrefixToOriginNode(ctx, prefix, prefixLen, originAS, prefixData, isIPv4)
	} else {
		return uc.createBGPPrefixVertex(ctx, key, prefixData, prefixType, isIPv4)
	}
}

func (uc *UpdateCoordinator) processPrefixWithdrawal(ctx context.Context, key string, prefixData map[string]interface{}) error {
	prefix, _ := prefixData["prefix"].(string)
	prefixLen := getUint32FromInterface(prefixData["prefix_len"])
	originAS := getUint32FromInterface(prefixData["origin_as"])
	isIPv4, _ := prefixData["is_ipv4"].(bool)

	glog.V(7).Infof("Withdrawing BGP prefix: %s/%d from AS%d", prefix, prefixLen, originAS)

	// Determine if this was node metadata or separate vertex
	if uc.shouldAttachAsNodeMetadata(prefixLen, isIPv4) {
		return uc.removePrefixFromOriginNode(ctx, prefix, prefixLen, originAS, isIPv4)
	} else {
		return uc.removeBGPPrefixVertex(ctx, key, prefixData, isIPv4)
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

	// Strategy 2: Look for existing BGP nodes with matching ASN
	// For now, we'll create a representative BGP node using origin_as
	// In a more sophisticated implementation, we might track all BGP speakers per AS
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

	// Find origin node
	originNodeID, err := uc.findOriginNode(ctx, originAS, prefixData)
	if err != nil {
		return fmt.Errorf("failed to find origin node: %w", err)
	}

	if originNodeID == "" {
		// Create origin node if it doesn't exist
		originNodeID, err = uc.createOriginBGPNode(ctx, originAS, prefixData)
		if err != nil {
			return fmt.Errorf("failed to create origin node: %w", err)
		}
	}

	// Determine target graph collection
	var targetCollection driver.Collection
	var prefixCollection string
	if isIPv4 {
		targetCollection = uc.db.ipv4Graph
		prefixCollection = uc.db.config.BGPPrefixV4
	} else {
		targetCollection = uc.db.ipv6Graph
		prefixCollection = uc.db.config.BGPPrefixV6
	}

	// Create bidirectional edges between origin node and prefix
	prefixVertexID := fmt.Sprintf("%s/%s", prefixCollection, prefixKey)

	// Edge from origin to prefix
	edgeKey1 := fmt.Sprintf("%s_to_prefix", prefixKey)
	edge1 := &IPGraphObject{
		Key:      edgeKey1,
		From:     originNodeID,
		To:       prefixVertexID,
		Protocol: "BGP_PREFIX",
		Link:     prefixKey,
	}

	// Edge from prefix to origin
	edgeKey2 := fmt.Sprintf("prefix_to_%s", prefixKey)
	edge2 := &IPGraphObject{
		Key:      edgeKey2,
		From:     prefixVertexID,
		To:       originNodeID,
		Protocol: "BGP_PREFIX",
		Link:     prefixKey,
	}

	// Create both edges
	for _, edge := range []*IPGraphObject{edge1, edge2} {
		if _, err := targetCollection.CreateDocument(ctx, edge); err != nil {
			if !driver.IsConflict(err) {
				return fmt.Errorf("failed to create prefix edge %s: %w", edge.Key, err)
			}
			// Update existing edge
			if _, err := targetCollection.UpdateDocument(ctx, edge.Key, edge); err != nil {
				return fmt.Errorf("failed to update prefix edge %s: %w", edge.Key, err)
			}
		}
	}

	glog.V(8).Infof("Created prefix-to-origin edges for: %s", prefixKey)
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
	// Remove prefix vertex and associated edges
	var targetCollections []driver.Collection
	if isIPv4 {
		targetCollections = []driver.Collection{uc.db.bgpPrefixV4, uc.db.ipv4Graph}
	} else {
		targetCollections = []driver.Collection{uc.db.bgpPrefixV6, uc.db.ipv6Graph}
	}

	// Remove prefix vertex
	if _, err := targetCollections[0].RemoveDocument(ctx, key); err != nil {
		if !driver.IsNotFoundGeneral(err) {
			glog.Warningf("Failed to remove BGP prefix vertex %s: %v", key, err)
		}
	}

	// Remove associated edges
	edgeKeys := []string{
		fmt.Sprintf("%s_to_prefix", key),
		fmt.Sprintf("prefix_to_%s", key),
	}

	for _, edgeKey := range edgeKeys {
		if _, err := targetCollections[1].RemoveDocument(ctx, edgeKey); err != nil {
			if !driver.IsNotFoundGeneral(err) {
				glog.V(6).Infof("Failed to remove prefix edge %s: %v", edgeKey, err)
			}
		}
	}

	glog.V(8).Infof("Removed BGP prefix vertex and edges: %s", key)
	return nil
}

// Helper function to safely get string from interface (BGP prefix processor)
func getStringFromData(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}
