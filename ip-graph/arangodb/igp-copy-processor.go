package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// IGPCopyProcessor handles copying IGP graph data to full topology graphs
type IGPCopyProcessor struct {
	db *arangoDB
}

// NewIGPCopyProcessor creates a new IGP copy processor
func NewIGPCopyProcessor(db *arangoDB) *IGPCopyProcessor {
	return &IGPCopyProcessor{
		db: db,
	}
}

// CopyIGPGraphsToFullTopology performs initial copy of IGP graphs to full topology
func (icp *IGPCopyProcessor) CopyIGPGraphsToFullTopology(ctx context.Context) error {
	glog.Infof("Starting IGP graph copy to full topology...")

	// Copy IGP nodes first (vertices)
	if err := icp.copyIGPNodes(ctx); err != nil {
		return fmt.Errorf("failed to copy IGP nodes: %w", err)
	}

	// Copy IGP edges (links and prefixes)
	if err := icp.copyIGPv4Edges(ctx); err != nil {
		return fmt.Errorf("failed to copy IGPv4 edges: %w", err)
	}

	if err := icp.copyIGPv6Edges(ctx); err != nil {
		return fmt.Errorf("failed to copy IGPv6 edges: %w", err)
	}

	glog.Infof("IGP graph copy to full topology completed successfully")
	return nil
}

// copyIGPNodes copies all igp_node documents to both IPv4 and IPv6 full topology graphs
func (icp *IGPCopyProcessor) copyIGPNodes(ctx context.Context) error {
	glog.V(6).Infof("Copying IGP nodes to full topology...")

	// Query all IGP nodes
	query := fmt.Sprintf(`
		FOR node IN %s
		RETURN node
	`, icp.db.config.IGPNode)

	cursor, err := icp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query IGP nodes: %w", err)
	}
	defer cursor.Close()

	nodeCount := 0
	for cursor.HasMore() {
		var igpNode map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &igpNode); err != nil {
			return fmt.Errorf("failed to read IGP node: %w", err)
		}

		// Create IP topology node from IGP node
		ipNode := icp.createIPNodeFromIGP(igpNode)

		// Insert into both IPv4 and IPv6 graphs as base topology
		// IGP nodes exist in both graphs since they can carry both IPv4 and IPv6 prefixes
		if err := icp.insertIPNode(ctx, ipNode); err != nil {
			glog.Warningf("Failed to insert IP node %s: %v", ipNode.Key, err)
			continue
		}

		nodeCount++
	}

	glog.V(6).Infof("Copied %d IGP nodes to full topology", nodeCount)
	return nil
}

// createIPNodeFromIGP converts an IGP node to an IP topology node
func (icp *IGPCopyProcessor) createIPNodeFromIGP(igpNode map[string]interface{}) *IPNode {
	// Extract key without collection prefix if present
	key := ""
	if keyVal, ok := igpNode["_key"].(string); ok {
		key = keyVal
	} else if idVal, ok := igpNode["_id"].(string); ok {
		// Extract key from _id (format: "collection/key")
		for i := len(idVal) - 1; i >= 0; i-- {
			if idVal[i] == '/' {
				key = idVal[i+1:]
				break
			}
		}
	}

	ipNode := &IPNode{
		Key:         key,
		Action:      getStringFromMap(igpNode, "action"),
		RouterID:    getStringFromMap(igpNode, "router_id"),
		DomainID:    int64(getUint32FromMap(igpNode, "domain_id")),
		BGPRouterID: getStringFromMap(igpNode, "bgp_router_id"),
		ASN:         getUint32FromMap(igpNode, "asn"),
		// MTID will be handled as interface{} for now
		AreaID:   getStringFromMap(igpNode, "area_id"),
		Protocol: getStringFromMap(igpNode, "protocol"),
		// ProtocolID will be handled as interface{} for now
		Name: getStringFromMap(igpNode, "name"),
		// Complex types will be handled as interface{} for now
		NodeType: "igp", // Mark as IGP-originated node
		Tier:     icp.determineNodeTier(getUint32FromMap(igpNode, "asn")),
	}

	// Copy SRv6 SIDs if present
	if sids, ok := igpNode["sids"].([]interface{}); ok {
		// Convert to SID structs if needed
		ipNode.SIDS = convertToSIDs(sids)
	}

	// Copy prefixes if present
	if prefixes, ok := igpNode["prefixes"].([]interface{}); ok {
		ipNode.Prefixes = prefixes
	}

	return ipNode
}

// insertIPNode inserts an IP node into the topology
func (icp *IGPCopyProcessor) insertIPNode(ctx context.Context, ipNode *IPNode) error {
	// For now, we don't create separate node collections for IP topology
	// The nodes exist in igp_node and bgp_node collections
	// The graphs (ipv4_graph, ipv6_graph) reference these nodes via their _id

	// This is a placeholder - in practice, IGP nodes already exist and we just
	// need to ensure the graph edge collections can reference them
	glog.V(9).Infof("IP node ready for graph referencing: %s", ipNode.Key)
	return nil
}

// copyIGPv4Edges copies all IGPv4 graph edges to the full IPv4 topology
func (icp *IGPCopyProcessor) copyIGPv4Edges(ctx context.Context) error {
	glog.V(6).Infof("Copying IGPv4 edges to full topology...")

	// Query all IGPv4 graph edges
	query := fmt.Sprintf(`
		FOR edge IN %s
		RETURN edge
	`, icp.db.config.IGPv4Graph)

	cursor, err := icp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query IGPv4 edges: %w", err)
	}
	defer cursor.Close()

	edgeCount := 0
	for cursor.HasMore() {
		var igpEdge map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &igpEdge); err != nil {
			return fmt.Errorf("failed to read IGPv4 edge: %w", err)
		}

		// Create IP graph edge from IGP edge
		ipEdge := icp.createIPEdgeFromIGP(igpEdge, true) // true for IPv4

		// Insert into IPv4 full topology graph
		if _, err := icp.db.ipv4Graph.CreateDocument(ctx, ipEdge); err != nil {
			if !driver.IsConflict(err) {
				glog.Warningf("Failed to insert IPv4 edge %s: %v", ipEdge.Key, err)
				continue
			}
			// Update existing edge
			if _, err := icp.db.ipv4Graph.UpdateDocument(ctx, ipEdge.Key, ipEdge); err != nil {
				glog.Warningf("Failed to update IPv4 edge %s: %v", ipEdge.Key, err)
				continue
			}
		}

		edgeCount++
	}

	glog.V(6).Infof("Copied %d IGPv4 edges to full topology", edgeCount)
	return nil
}

// copyIGPv6Edges copies all IGPv6 graph edges to the full IPv6 topology
func (icp *IGPCopyProcessor) copyIGPv6Edges(ctx context.Context) error {
	glog.V(6).Infof("Copying IGPv6 edges to full topology...")

	// Query all IGPv6 graph edges
	query := fmt.Sprintf(`
		FOR edge IN %s
		RETURN edge
	`, icp.db.config.IGPv6Graph)

	cursor, err := icp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query IGPv6 edges: %w", err)
	}
	defer cursor.Close()

	edgeCount := 0
	for cursor.HasMore() {
		var igpEdge map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &igpEdge); err != nil {
			return fmt.Errorf("failed to read IGPv6 edge: %w", err)
		}

		// Create IP graph edge from IGP edge
		ipEdge := icp.createIPEdgeFromIGP(igpEdge, false) // false for IPv6

		// Insert into IPv6 full topology graph
		if _, err := icp.db.ipv6Graph.CreateDocument(ctx, ipEdge); err != nil {
			if !driver.IsConflict(err) {
				glog.Warningf("Failed to insert IPv6 edge %s: %v", ipEdge.Key, err)
				continue
			}
			// Update existing edge
			if _, err := icp.db.ipv6Graph.UpdateDocument(ctx, ipEdge.Key, ipEdge); err != nil {
				glog.Warningf("Failed to update IPv6 edge %s: %v", ipEdge.Key, err)
				continue
			}
		}

		edgeCount++
	}

	glog.V(6).Infof("Copied %d IGPv6 edges to full topology", edgeCount)
	return nil
}

// createIPEdgeFromIGP converts an IGP graph edge to an IP topology edge
func (icp *IGPCopyProcessor) createIPEdgeFromIGP(igpEdge map[string]interface{}, isIPv4 bool) *IPGraphObject {
	// Extract key, ensuring uniqueness in the IP graph
	originalKey := getStringFromMap(igpEdge, "_key")
	ipEdgeKey := fmt.Sprintf("igp_%s", originalKey)

	ipVersion := "IPv6"
	if isIPv4 {
		ipVersion = "IPv4"
	}

	ipEdge := &IPGraphObject{
		Key:                   ipEdgeKey,
		From:                  getStringFromMap(igpEdge, "_from"),
		To:                    getStringFromMap(igpEdge, "_to"),
		Link:                  getStringFromMap(igpEdge, "link"),
		Protocol:              fmt.Sprintf("IGP_%s", ipVersion),
		DomainID:              getUint32FromMap(igpEdge, "domain_id"),
		MTID:                  getUint16FromMap(igpEdge, "mt_id"),
		AreaID:                getStringFromMap(igpEdge, "area_id"),
		LocalLinkID:           getUint32FromMap(igpEdge, "local_link_id"),
		RemoteLinkID:          getUint32FromMap(igpEdge, "remote_link_id"),
		LocalLinkIP:           getStringFromMap(igpEdge, "local_link_ip"),
		RemoteLinkIP:          getStringFromMap(igpEdge, "remote_link_ip"),
		LocalNodeASN:          getUint32FromMap(igpEdge, "local_node_asn"),
		RemoteNodeASN:         getUint32FromMap(igpEdge, "remote_node_asn"),
		IGPMetric:             getUint32FromMap(igpEdge, "igp_metric"),
		MaxLinkBWKbps:         getUint64FromMap(igpEdge, "max_link_bw_kbps"),
		SRv6EndXSID:           getInterfaceFromMap(igpEdge, "srv6_endx_sid"),
		LSAdjacencySID:        getInterfaceFromMap(igpEdge, "ls_adjacency_sid"),
		UnidirLinkDelayMinMax: getInterfaceFromMap(igpEdge, "unidir_link_delay_min_max"),
		AppSpecLinkAttr:       getInterfaceFromMap(igpEdge, "app_spec_link_attr"),
	}

	return ipEdge
}

// syncIGPNodeUpdate handles real-time IGP node updates
func (uc *UpdateCoordinator) syncIGPNodeUpdate(ctx context.Context, nodeKey string, action string) error {
	glog.V(7).Infof("Syncing IGP node update: %s action: %s", nodeKey, action)

	switch action {
	case "del":
		return uc.removeIGPNodeFromFullTopology(ctx, nodeKey)
	case "add", "update":
		return uc.updateIGPNodeInFullTopology(ctx, nodeKey)
	default:
		glog.V(6).Infof("Unknown IGP node action: %s for key: %s", action, nodeKey)
		return nil
	}
}

// updateIGPNodeInFullTopology updates an IGP node in the full topology
func (uc *UpdateCoordinator) updateIGPNodeInFullTopology(ctx context.Context, nodeKey string) error {
	// Read the updated IGP node
	var igpNode map[string]interface{}
	if _, err := uc.db.igpNode.ReadDocument(ctx, nodeKey, &igpNode); err != nil {
		return fmt.Errorf("failed to read IGP node %s: %w", nodeKey, err)
	}

	// The IGP node already exists in the igp_node collection
	// The full topology graphs reference it by _id, so no additional action needed
	// unless we're maintaining separate IP node collections (which we're not currently)

	glog.V(8).Infof("IGP node %s updated in full topology", nodeKey)
	return nil
}

// removeIGPNodeFromFullTopology removes an IGP node from the full topology
func (uc *UpdateCoordinator) removeIGPNodeFromFullTopology(ctx context.Context, nodeKey string) error {
	// When an IGP node is deleted, we need to clean up any edges in the full topology
	// that reference this node
	nodeID := fmt.Sprintf("%s/%s", uc.db.config.IGPNode, nodeKey)

	// Remove edges from IPv4 graph
	if err := uc.removeEdgesReferencingNode(ctx, uc.db.ipv4Graph, nodeID); err != nil {
		glog.Warningf("Failed to clean up IPv4 edges for node %s: %v", nodeKey, err)
	}

	// Remove edges from IPv6 graph
	if err := uc.removeEdgesReferencingNode(ctx, uc.db.ipv6Graph, nodeID); err != nil {
		glog.Warningf("Failed to clean up IPv6 edges for node %s: %v", nodeKey, err)
	}

	glog.V(8).Infof("IGP node %s removed from full topology", nodeKey)
	return nil
}

// removeEdgesReferencingNode removes all edges that reference a specific node
func (uc *UpdateCoordinator) removeEdgesReferencingNode(ctx context.Context, collection driver.Collection, nodeID string) error {
	// Query edges that reference this node as _from or _to
	query := fmt.Sprintf(`
		FOR edge IN %s
		FILTER edge._from == @nodeId OR edge._to == @nodeId
		REMOVE edge IN %s
	`, collection.Name(), collection.Name())

	bindVars := map[string]interface{}{
		"nodeId": nodeID,
	}

	cursor, err := uc.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to remove edges for node %s: %w", nodeID, err)
	}
	defer cursor.Close()

	return nil
}

// syncIGPLinkUpdate handles real-time IGP link updates
func (uc *UpdateCoordinator) syncIGPLinkUpdate(ctx context.Context, linkKey string, action string, isIPv4 bool) error {
	ipVersion := "IPv4"
	if !isIPv4 {
		ipVersion = "IPv6"
	}

	glog.V(7).Infof("Syncing IGP %s link update: %s action: %s", ipVersion, linkKey, action)

	var targetCollection driver.Collection
	var sourceCollection string

	if isIPv4 {
		targetCollection = uc.db.ipv4Graph
		sourceCollection = uc.db.config.IGPv4Graph
	} else {
		targetCollection = uc.db.ipv6Graph
		sourceCollection = uc.db.config.IGPv6Graph
	}

	switch action {
	case "del":
		return uc.removeIGPLinkFromFullTopology(ctx, targetCollection, linkKey)
	case "add", "update":
		return uc.updateIGPLinkInFullTopology(ctx, targetCollection, sourceCollection, linkKey, isIPv4)
	default:
		glog.V(6).Infof("Unknown IGP link action: %s for key: %s", action, linkKey)
		return nil
	}
}

// updateIGPLinkInFullTopology updates an IGP link in the full topology
func (uc *UpdateCoordinator) updateIGPLinkInFullTopology(ctx context.Context, targetCollection driver.Collection, sourceCollection, linkKey string, isIPv4 bool) error {
	// Read the IGP edge from source collection
	query := fmt.Sprintf(`
		FOR edge IN %s
		FILTER edge._key == @linkKey
		RETURN edge
	`, sourceCollection)

	bindVars := map[string]interface{}{
		"linkKey": linkKey,
	}

	cursor, err := uc.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to query IGP edge %s: %w", linkKey, err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		glog.V(6).Infof("IGP edge %s not found in source collection", linkKey)
		return nil
	}

	var igpEdge map[string]interface{}
	if _, err := cursor.ReadDocument(ctx, &igpEdge); err != nil {
		return fmt.Errorf("failed to read IGP edge %s: %w", linkKey, err)
	}

	// Create IP edge from IGP edge
	icp := NewIGPCopyProcessor(uc.db)
	ipEdge := icp.createIPEdgeFromIGP(igpEdge, isIPv4)

	// Insert or update in target collection
	if _, err := targetCollection.CreateDocument(ctx, ipEdge); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create IP edge %s: %w", ipEdge.Key, err)
		}
		// Update existing edge
		if _, err := targetCollection.UpdateDocument(ctx, ipEdge.Key, ipEdge); err != nil {
			return fmt.Errorf("failed to update IP edge %s: %w", ipEdge.Key, err)
		}
	}

	glog.V(8).Infof("IGP link %s updated in full topology", linkKey)
	return nil
}

// removeIGPLinkFromFullTopology removes an IGP link from the full topology
func (uc *UpdateCoordinator) removeIGPLinkFromFullTopology(ctx context.Context, targetCollection driver.Collection, linkKey string) error {
	// Remove the corresponding IP edge
	ipEdgeKey := fmt.Sprintf("igp_%s", linkKey)

	if _, err := targetCollection.RemoveDocument(ctx, ipEdgeKey); err != nil {
		if !driver.IsNotFoundGeneral(err) {
			return fmt.Errorf("failed to remove IP edge %s: %w", ipEdgeKey, err)
		}
		// Edge doesn't exist, which is fine
	}

	glog.V(8).Infof("IGP link %s removed from full topology", linkKey)
	return nil
}

// Helper functions for safe type conversion
func getStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getUint32FromMap(data map[string]interface{}, key string) uint32 {
	return getUint32FromInterface(data[key])
}

func getUint16FromMap(data map[string]interface{}, key string) uint16 {
	if val := getUint32FromInterface(data[key]); val <= 65535 {
		return uint16(val)
	}
	return 0
}

func getUint64FromMap(data map[string]interface{}, key string) uint64 {
	switch val := data[key].(type) {
	case float64:
		return uint64(val)
	case uint64:
		return val
	case uint32:
		return uint64(val)
	case int64:
		if val >= 0 {
			return uint64(val)
		}
	}
	return 0
}

func getInterfaceFromMap(data map[string]interface{}, key string) interface{} {
	return data[key]
}

func (icp *IGPCopyProcessor) determineNodeTier(asn uint32) string {
	if asn >= 64512 && asn <= 65535 {
		return "private"
	} else if asn >= 4200000000 && asn <= 4294967294 {
		return "private_4byte"
	} else if asn >= 1 && asn <= 64511 {
		if asn <= 100 {
			return "tier1"
		} else if asn <= 10000 {
			return "tier2"
		} else {
			return "tier3"
		}
	}
	return "unknown"
}

// convertToSIDs converts interface{} slice to SID structs
func convertToSIDs(sids []interface{}) []SID {
	result := make([]SID, 0, len(sids))
	for _, sidData := range sids {
		if sidMap, ok := sidData.(map[string]interface{}); ok {
			sid := SID{
				SRv6SID: getStringFromMap(sidMap, "srv6_sid"),
				// Other fields can be added as needed
			}
			result = append(result, sid)
		}
	}
	return result
}
