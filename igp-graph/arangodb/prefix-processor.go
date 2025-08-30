package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// Prefix represents a prefix entry for metadata processing
type Prefix struct {
	Key            string      `json:"_key"`
	ID             string      `json:"_id"`
	Prefix         string      `json:"prefix"`
	PrefixLen      int32       `json:"prefix_len"`
	PrefixMetric   uint32      `json:"prefix_metric"`
	PrefixAttrTLVs interface{} `json:"prefix_attr_tlvs"`
	IGPRouterID    string      `json:"igp_router_id"`
	DomainID       interface{} `json:"domain_id"`
	ProtocolID     interface{} `json:"protocol_id"`
	Protocol       string      `json:"protocol"`
	AreaID         string      `json:"area_id"`
	MTID           interface{} `json:"mt_id_tlv"`
}

// loadInitialPrefixes loads all existing ls_prefix entries and processes them
func (a *arangoDB) loadInitialPrefixes(ctx context.Context) error {
	glog.V(6).Info("Loading initial prefixes...")

	query := fmt.Sprintf("FOR p IN %s RETURN p", "ls_prefix")
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query ls_prefix collection: %w", err)
	}
	defer cursor.Close()

	count := 0
	for {
		var prefix map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &prefix)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return fmt.Errorf("error reading prefix document: %w", err)
		}

		if err := a.processInitialPrefix(ctx, prefix); err != nil {
			glog.Warningf("Failed to process initial prefix %v: %v", prefix["_key"], err)
			continue
		}

		count++
		if count%100 == 0 {
			glog.V(3).Infof("Loaded %d prefixes...", count)
		}
	}

	glog.Infof("Loaded %d initial prefixes", count)
	return nil
}

// processInitialPrefix processes a single ls_prefix entry during initial load
func (a *arangoDB) processInitialPrefix(ctx context.Context, prefix map[string]interface{}) error {
	// Filter out BGP protocol (protocol_id = 7)
	if protocolID, ok := prefix["protocol_id"].(float64); ok && protocolID == 7 {
		glog.V(9).Infof("Skipping BGP prefix %s", prefix["_key"])
		return nil
	}

	// Filter out prefixes that match SRv6 locators
	if isMatched, err := a.isPrefixMatchingSRv6Locator(ctx, prefix); err != nil {
		glog.Warningf("Failed to check SRv6 locator match for prefix %s: %v", prefix["_key"], err)
	} else if isMatched {
		glog.V(8).Infof("Skipping prefix %s as it matches an SRv6 locator", prefix["_key"])
		return nil
	}

	// Extract prefix length
	prefixLen, ok := prefix["prefix_len"].(float64)
	if !ok {
		return fmt.Errorf("invalid prefix_len in prefix %s", prefix["_key"])
	}

	// Determine IPv4 vs IPv6 based on MTID
	isIPv6 := false
	if mtidTLV, exists := prefix["mt_id_tlv"]; exists {
		if mtidArray, ok := mtidTLV.([]interface{}); ok {
			for _, mtObj := range mtidArray {
				if mtMap, ok := mtObj.(map[string]interface{}); ok {
					if mtid, ok := mtMap["mt_id"].(float64); ok && mtid == 2 {
						isIPv6 = true
						break
					}
				}
			}
		} else if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			if mtid, ok := mtidObj["mt_id"].(float64); ok && mtid == 2 {
				isIPv6 = true
			}
		}
	}

	// Apply prefix strategy based on prefix length
	if isIPv6 {
		return a.processIPv6Prefix(ctx, prefix, int32(prefixLen))
	} else {
		return a.processIPv4Prefix(ctx, prefix, int32(prefixLen))
	}
}

// processIPv4Prefix processes IPv4 prefixes according to our strategy
func (a *arangoDB) processIPv4Prefix(ctx context.Context, prefix map[string]interface{}, prefixLen int32) error {
	switch {
	case prefixLen == 32:
		// /32 prefixes as node metadata
		return a.addPrefixToNodeMetadata(ctx, prefix)
	case prefixLen == 30 || prefixLen == 31:
		// Skip point-to-point links (original behavior)
		glog.V(9).Infof("Skipping P2P prefix /%d: %s", prefixLen, prefix["_key"])
		return nil
	default:
		// Transit networks as separate vertices
		return a.createPrefixVertex(ctx, prefix, false) // false = IPv4
	}
}

// processIPv6Prefix processes IPv6 prefixes according to our strategy
func (a *arangoDB) processIPv6Prefix(ctx context.Context, prefix map[string]interface{}, prefixLen int32) error {
	switch {
	case prefixLen == 128:
		// /128 prefixes as node metadata
		return a.addPrefixToNodeMetadata(ctx, prefix)
	case prefixLen == 126 || prefixLen == 127:
		// Skip point-to-point links (original behavior)
		glog.V(9).Infof("Skipping P2P prefix /%d: %s", prefixLen, prefix["_key"])
		return nil
	default:
		// Transit networks as separate vertices
		return a.createPrefixVertex(ctx, prefix, true) // true = IPv6
	}
}

// addPrefixToNodeMetadata adds /32 or /128 prefix as metadata to the corresponding IGP node
func (a *arangoDB) addPrefixToNodeMetadata(ctx context.Context, prefix map[string]interface{}) error {
	routerID, ok := prefix["igp_router_id"].(string)
	if !ok {
		return fmt.Errorf("missing igp_router_id in prefix %s", prefix["_key"])
	}

	domainID := prefix["domain_id"]
	protocolID := prefix["protocol_id"]
	areaID, _ := prefix["area_id"].(string)

	// Find the corresponding IGP node
	node, err := a.findNodeForPrefix(ctx, routerID, domainID, protocolID, areaID)
	if err != nil {
		return fmt.Errorf("failed to find IGP node for prefix %s: %w", prefix["_key"], err)
	}

	// Create prefix metadata object
	prefixMeta := map[string]interface{}{
		"prefix":           prefix["prefix"],
		"prefix_len":       prefix["prefix_len"],
		"prefix_metric":    prefix["prefix_metric"],
		"prefix_attr_tlvs": prefix["prefix_attr_tlvs"],
		"_key":             prefix["_key"],
	}

	// Add to node's prefixes array
	return a.addPrefixMetadataToNode(ctx, node, prefixMeta)
}

// findNodeForPrefix finds the IGP node that corresponds to a prefix
func (a *arangoDB) findNodeForPrefix(ctx context.Context, routerID string, domainID, protocolID interface{}, areaID string) (map[string]interface{}, error) {
	query := fmt.Sprintf("FOR d IN %s", a.config.IGPNode)
	query += " FILTER d.igp_router_id == @routerId"
	query += " FILTER d.domain_id == @domainId"

	bindVars := map[string]interface{}{
		"routerId": routerID,
		"domainId": domainID,
	}

	// For OSPF (protocol 3=OSPFv2, 6=OSPFv3), include area_id in query
	if proto, ok := protocolID.(float64); ok && (proto == 3 || proto == 6) {
		query += " FILTER d.area_id == @areaId"
		bindVars["areaId"] = areaID
	}

	query += " RETURN d"

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
	}

	if count == 0 {
		return nil, fmt.Errorf("no IGP node found for router %s", routerID)
	}
	if count > 1 {
		return nil, fmt.Errorf("multiple IGP nodes found for router %s", routerID)
	}

	return node, nil
}

// addPrefixMetadataToNode adds prefix metadata to a node's prefixes array
func (a *arangoDB) addPrefixMetadataToNode(ctx context.Context, node map[string]interface{}, prefixMeta map[string]interface{}) error {
	nodeKey, ok := node["_key"].(string)
	if !ok {
		return fmt.Errorf("invalid node key")
	}

	// Get existing prefixes array or create new one
	var prefixes []interface{}
	if existingPrefixes, exists := node["prefixes"]; exists {
		if prefixArray, ok := existingPrefixes.([]interface{}); ok {
			prefixes = prefixArray
		}
	}

	// Check if prefix already exists (avoid duplicates)
	prefixKey := prefixMeta["_key"].(string)
	for _, existing := range prefixes {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			if existingKey, ok := existingMap["_key"].(string); ok && existingKey == prefixKey {
				glog.V(9).Infof("Prefix %s already exists in node %s metadata", prefixKey, nodeKey)
				return nil
			}
		}
	}

	// Add new prefix to array
	prefixes = append(prefixes, prefixMeta)

	// Update node document
	update := map[string]interface{}{
		"prefixes": prefixes,
	}

	collection, err := a.db.Collection(ctx, a.config.IGPNode)
	if err != nil {
		return fmt.Errorf("failed to get IGP node collection: %w", err)
	}

	_, err = collection.UpdateDocument(ctx, nodeKey, update)
	if err != nil {
		return fmt.Errorf("failed to update node %s with prefix metadata: %w", nodeKey, err)
	}

	glog.V(8).Infof("Added prefix %s as metadata to node %s", prefixKey, nodeKey)
	return nil
}

// createPrefixVertex creates a separate vertex for transit network prefixes
func (a *arangoDB) createPrefixVertex(ctx context.Context, prefix map[string]interface{}, isIPv6 bool) error {
	// Find the node that advertises this prefix
	routerID, ok := prefix["igp_router_id"].(string)
	if !ok {
		return fmt.Errorf("missing igp_router_id in prefix %s", prefix["_key"])
	}

	domainID := prefix["domain_id"]
	protocolID := prefix["protocol_id"]
	areaID, _ := prefix["area_id"].(string)

	node, err := a.findNodeForPrefix(ctx, routerID, domainID, protocolID, areaID)
	if err != nil {
		return fmt.Errorf("failed to find IGP node for prefix %s: %w", prefix["_key"], err)
	}

	// Create prefix vertex edges in appropriate graph
	if isIPv6 {
		return a.createIPv6PrefixEdges(ctx, prefix, node)
	} else {
		return a.createIPv4PrefixEdges(ctx, prefix, node)
	}
}

// createIPv4PrefixEdges creates bidirectional edges between node and prefix in IPv4 graph
func (a *arangoDB) createIPv4PrefixEdges(ctx context.Context, prefix map[string]interface{}, node map[string]interface{}) error {
	return a.createPrefixEdges(ctx, prefix, node, a.config.IGPv4Graph, false)
}

// createIPv6PrefixEdges creates bidirectional edges between node and prefix in IPv6 graph
func (a *arangoDB) createIPv6PrefixEdges(ctx context.Context, prefix map[string]interface{}, node map[string]interface{}) error {
	return a.createPrefixEdges(ctx, prefix, node, a.config.IGPv6Graph, true)
}

// createPrefixEdges creates bidirectional edges between node and prefix (mirrors original logic)
func (a *arangoDB) createPrefixEdges(ctx context.Context, prefix map[string]interface{}, node map[string]interface{}, graphCollection string, isIPv6 bool) error {
	prefixKey, _ := prefix["_key"].(string)
	prefixID, _ := prefix["_id"].(string)
	nodeKey, _ := node["_key"].(string)
	nodeID, _ := node["_id"].(string)

	// Extract MTID
	var mtid uint16 = 0
	if mtidTLV, exists := prefix["mt_id_tlv"]; exists {
		if mtidArray, ok := mtidTLV.([]interface{}); ok {
			for _, mtObj := range mtidArray {
				if mtMap, ok := mtObj.(map[string]interface{}); ok {
					if mt, ok := mtMap["mt_id"].(float64); ok {
						mtid = uint16(mt)
						break
					}
				}
			}
		} else if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			if mt, ok := mtidObj["mt_id"].(float64); ok {
				mtid = uint16(mt)
			}
		}
	}

	// Helper function for safe conversions
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

	getString := func(m map[string]interface{}, key string) string {
		if val, ok := m[key].(string); ok {
			return val
		}
		return ""
	}

	// Node to Prefix direction
	nodeToPrefix := IGPGraphObject{
		Key:            nodeKey + "_to_" + prefixKey,
		From:           nodeID,
		To:             prefixID,
		Link:           prefixKey,
		ProtocolID:     prefix["protocol_id"],
		DomainID:       prefix["domain_id"],
		MTID:           mtid,
		AreaID:         getString(prefix, "area_id"),
		Protocol:       getString(prefix, "protocol"),
		LocalNodeASN:   getUint32(node["asn"]),
		Prefix:         getString(prefix, "prefix"),
		PrefixLen:      int32(getUint32(prefix["prefix_len"])),
		PrefixMetric:   getUint32(prefix["prefix_metric"]),
		PrefixAttrTLVs: prefix["prefix_attr_tlvs"],
	}

	// Prefix to Node direction
	prefixToNode := IGPGraphObject{
		Key:            prefixKey + "_to_" + nodeKey,
		From:           prefixID,
		To:             nodeID,
		Link:           prefixKey,
		ProtocolID:     prefix["protocol_id"],
		DomainID:       prefix["domain_id"],
		MTID:           mtid,
		AreaID:         getString(prefix, "area_id"),
		Protocol:       getString(prefix, "protocol"),
		LocalNodeASN:   getUint32(node["asn"]),
		Prefix:         getString(prefix, "prefix"),
		PrefixLen:      int32(getUint32(prefix["prefix_len"])),
		PrefixMetric:   getUint32(prefix["prefix_metric"]),
		PrefixAttrTLVs: prefix["prefix_attr_tlvs"],
	}

	// Create/Update both directions
	collection, err := a.db.Collection(ctx, graphCollection)
	if err != nil {
		return fmt.Errorf("failed to get graph collection %s: %w", graphCollection, err)
	}
	for _, edge := range []*IGPGraphObject{&nodeToPrefix, &prefixToNode} {
		if _, err := collection.CreateDocument(ctx, edge); err != nil {
			if !driver.IsConflict(err) {
				return fmt.Errorf("failed to create prefix edge %s: %w", edge.Key, err)
			}
			// Document exists, update it
			if _, err := collection.UpdateDocument(ctx, edge.Key, edge); err != nil {
				return fmt.Errorf("failed to update prefix edge %s: %w", edge.Key, err)
			}
		}
	}

	glog.V(8).Infof("Created prefix edges for %s in %s graph", prefixKey, graphCollection)
	return nil
}

// removePrefixFromNodeMetadata removes a prefix from node metadata
func (a *arangoDB) removePrefixFromNodeMetadata(ctx context.Context, prefix map[string]interface{}) error {
	routerID, ok := prefix["igp_router_id"].(string)
	if !ok {
		return fmt.Errorf("missing igp_router_id in prefix %s", prefix["_key"])
	}

	domainID := prefix["domain_id"]
	protocolID := prefix["protocol_id"]
	areaID, _ := prefix["area_id"].(string)

	// Find the corresponding IGP node
	node, err := a.findNodeForPrefix(ctx, routerID, domainID, protocolID, areaID)
	if err != nil {
		glog.V(6).Infof("IGP node not found for prefix removal %s: %v", prefix["_key"], err)
		return nil // Node might have been deleted already
	}

	nodeKey, ok := node["_key"].(string)
	if !ok {
		return fmt.Errorf("invalid node key")
	}

	// Get existing prefixes array
	var prefixes []interface{}
	if existingPrefixes, exists := node["prefixes"]; exists {
		if prefixArray, ok := existingPrefixes.([]interface{}); ok {
			prefixes = prefixArray
		}
	}

	// Remove the prefix from array
	prefixKey := prefix["_key"].(string)
	var updatedPrefixes []interface{}
	for _, existing := range prefixes {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			if existingKey, ok := existingMap["_key"].(string); ok && existingKey != prefixKey {
				updatedPrefixes = append(updatedPrefixes, existing)
			}
		}
	}

	// Update node document
	update := map[string]interface{}{
		"prefixes": updatedPrefixes,
	}

	collection, err := a.db.Collection(ctx, a.config.IGPNode)
	if err != nil {
		return fmt.Errorf("failed to get IGP node collection: %w", err)
	}

	_, err = collection.UpdateDocument(ctx, nodeKey, update)
	if err != nil {
		return fmt.Errorf("failed to remove prefix %s from node %s metadata: %w", prefixKey, nodeKey, err)
	}

	glog.V(8).Infof("Removed prefix %s from node %s metadata", prefixKey, nodeKey)
	return nil
}

// removePrefixVertex removes prefix vertex edges from graphs
func (a *arangoDB) removePrefixVertex(ctx context.Context, prefix map[string]interface{}, isIPv6 bool) error {
	prefixKey, _ := prefix["_key"].(string)

	// Find the node that advertised this prefix
	routerID, ok := prefix["igp_router_id"].(string)
	if !ok {
		return fmt.Errorf("missing igp_router_id in prefix %s", prefixKey)
	}

	domainID := prefix["domain_id"]
	protocolID := prefix["protocol_id"]
	areaID, _ := prefix["area_id"].(string)

	node, err := a.findNodeForPrefix(ctx, routerID, domainID, protocolID, areaID)
	if err != nil {
		glog.V(6).Infof("IGP node not found for prefix vertex removal %s: %v", prefixKey, err)
		return nil // Node might have been deleted already
	}

	nodeKey, _ := node["_key"].(string)

	// Remove edges from appropriate graph
	var graphCollection string
	if isIPv6 {
		graphCollection = a.config.IGPv6Graph
	} else {
		graphCollection = a.config.IGPv4Graph
	}

	collection, err := a.db.Collection(ctx, graphCollection)
	if err != nil {
		return fmt.Errorf("failed to get graph collection %s: %w", graphCollection, err)
	}

	// Remove both directions
	edgeKeys := []string{
		nodeKey + "_to_" + prefixKey,
		prefixKey + "_to_" + nodeKey,
	}

	for _, edgeKey := range edgeKeys {
		if _, err := collection.RemoveDocument(ctx, edgeKey); err != nil {
			if !driver.IsNotFoundGeneral(err) {
				glog.Warningf("Failed to remove prefix edge %s: %v", edgeKey, err)
			}
		}
	}

	glog.V(8).Infof("Removed prefix vertex edges for %s from %s graph", prefixKey, graphCollection)
	return nil
}

// isPrefixMatchingSRv6Locator checks if a prefix matches any existing SRv6 locator
// by comparing the prefix length with the calculated locator length from SRv6 SID structure
func (a *arangoDB) isPrefixMatchingSRv6Locator(ctx context.Context, prefix map[string]interface{}) (bool, error) {
	prefixStr, ok := prefix["prefix"].(string)
	if !ok {
		return false, fmt.Errorf("invalid prefix string in prefix %s", prefix["_key"])
	}

	prefixLen, ok := prefix["prefix_len"].(float64)
	if !ok {
		return false, fmt.Errorf("invalid prefix_len in prefix %s", prefix["_key"])
	}

	routerID, ok := prefix["igp_router_id"].(string)
	if !ok {
		return false, fmt.Errorf("missing igp_router_id in prefix %s", prefix["_key"])
	}

	domainID := prefix["domain_id"]
	prefixLenInt := int32(prefixLen)

	// Query ls_srv6_sid collection for SRv6 SIDs from the same router that match this prefix
	// Calculate the actual locator length from the SRv6 SID structure
	query := fmt.Sprintf(`
		FOR sid IN %s
		FILTER sid.igp_router_id == @routerId
		FILTER sid.domain_id == @domainId
		FILTER STARTS_WITH(sid.srv6_sid, @prefix)
		LET locatorLength = (
			HAS(sid, "srv6_sid_structure") && 
			HAS(sid.srv6_sid_structure, "locator_block_length") && 
			HAS(sid.srv6_sid_structure, "locator_node_length")
		) ? (
			sid.srv6_sid_structure.locator_block_length + sid.srv6_sid_structure.locator_node_length
		) : null
		FILTER locatorLength != null
		FILTER locatorLength == @prefixLen
		RETURN {
			srv6_sid: sid.srv6_sid,
			locator_length: locatorLength,
			sid_structure: sid.srv6_sid_structure
		}
	`, "ls_srv6_sid")

	bindVars := map[string]interface{}{
		"routerId":  routerID,
		"domainId":  domainID,
		"prefix":    prefixStr,
		"prefixLen": prefixLenInt,
	}

	glog.V(9).Infof("Checking SRv6 locator match for prefix %s/%d from router %s", prefixStr, prefixLenInt, routerID)

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return false, fmt.Errorf("failed to query SRv6 SIDs: %w", err)
	}
	defer cursor.Close()

	// If we find any matching SRv6 SID with exact locator length match, filter out this prefix
	if cursor.HasMore() {
		var result map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			return false, fmt.Errorf("error reading SRv6 SID document: %w", err)
		}

		sidStr, _ := result["srv6_sid"].(string)
		locatorLen, _ := result["locator_length"].(float64)

		glog.V(8).Infof("Prefix %s/%d matches SRv6 locator %s with calculated length /%d",
			prefixStr, prefixLenInt, sidStr, int32(locatorLen))
		return true, nil
	}

	return false, nil
}
