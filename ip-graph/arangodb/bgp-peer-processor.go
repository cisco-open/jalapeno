package arangodb

import (
	"context"
	"fmt"
	"strconv"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bgp"
)

// processBGPPeerUpdate processes BGP peer session messages
func (uc *UpdateCoordinator) processBGPPeerUpdate(msg *ProcessingMessage) error {
	glog.Infof("Processing BGP peer update: %s action: %s", msg.Key, msg.Action)

	ctx := context.TODO()

	switch msg.Action {
	case "del":
		return uc.processPeerDeletion(ctx, msg.Key, msg.Data)
	case "add", "update":
		return uc.processPeerAddUpdate(ctx, msg.Key, msg.Data)
	default:
		glog.V(5).Infof("Unknown peer action: %s for key: %s", msg.Action, msg.Key)
		return nil
	}
}

func (uc *UpdateCoordinator) processPeerAddUpdate(ctx context.Context, key string, peerData map[string]interface{}) error {
	// Extract peer information
	localBGPID, _ := peerData["local_bgp_id"].(string)
	remoteBGPID, _ := peerData["remote_bgp_id"].(string)
	localASN := getUint32FromInterface(peerData["local_asn"])
	remoteASN := getUint32FromInterface(peerData["remote_asn"])
	localIP, _ := peerData["local_ip"].(string)
	remoteIP, _ := peerData["remote_ip"].(string)

	if localBGPID == "" || remoteBGPID == "" || localASN == 0 || remoteASN == 0 {
		return fmt.Errorf("invalid peer data: missing required fields")
	}

	// Determine session type
	sessionType := uc.classifyBGPSession(localASN, remoteASN)

	glog.V(8).Infof("Processing %s session: %s (AS%d) ↔ %s (AS%d)",
		sessionType, localBGPID, localASN, remoteBGPID, remoteASN)

	// Process local and remote BGP nodes
	if err := uc.ensureBGPNode(ctx, localBGPID, localASN, localIP, peerData, true); err != nil {
		return fmt.Errorf("failed to ensure local BGP node: %w", err)
	}

	if err := uc.ensureBGPNode(ctx, remoteBGPID, remoteASN, remoteIP, peerData, false); err != nil {
		return fmt.Errorf("failed to ensure remote BGP node: %w", err)
	}

	// Create bidirectional BGP session edges
	if err := uc.createBGPSessionEdges(ctx, key, peerData, sessionType); err != nil {
		return fmt.Errorf("failed to create BGP session edges: %w", err)
	}

	return nil
}

func (uc *UpdateCoordinator) processPeerDeletion(ctx context.Context, key string, peerData map[string]interface{}) error {
	glog.V(7).Infof("Deleting BGP peer session: %s", key)

	// Remove session edges from both IPv4 and IPv6 graphs
	if err := uc.removeBGPSessionEdges(ctx, key); err != nil {
		return fmt.Errorf("failed to remove BGP session edges: %w", err)
	}

	// Note: We don't remove BGP nodes as they might have other sessions
	// Node cleanup can be done separately if needed

	return nil
}

func (uc *UpdateCoordinator) classifyBGPSession(localASN, remoteASN uint32) string {
	if localASN == remoteASN {
		return "ibgp"
	}

	// Check for private ASNs (RFC 1930: 64512-65535, RFC 6996: 4200000000-4294967294)
	isLocalPrivate := (localASN >= 64512 && localASN <= 65535) || (localASN >= 4200000000 && localASN <= 4294967294)
	isRemotePrivate := (remoteASN >= 64512 && remoteASN <= 65535) || (remoteASN >= 4200000000 && remoteASN <= 4294967294)

	if isLocalPrivate && isRemotePrivate {
		return "ebgp_private"
	} else if !isLocalPrivate && !isRemotePrivate {
		return "ebgp_public"
	} else {
		// One private, one public - typically edge/transit connection
		return "ebgp_hybrid"
	}
}

func (uc *UpdateCoordinator) ensureBGPNode(ctx context.Context, bgpID string, asn uint32, ip string, peerData map[string]interface{}, isLocal bool) error {
	// Skip local nodes - we only create BGP nodes for remote peers (matching original logic)
	if isLocal {
		return nil
	}

	// Check if this ASN is already in IGP domain (matching original filter logic)
	igpASNExists, err := uc.checkIGPASNExists(ctx, asn)
	if err != nil {
		return fmt.Errorf("failed to check IGP ASN existence: %w", err)
	}

	if igpASNExists {
		glog.V(8).Infof("ASN %d exists in IGP domain, skipping BGP node creation", asn)
		return nil
	}

	// Create BGP node using original format: router_id + asn as key
	bgpNodeKey := fmt.Sprintf("%s_%d", bgpID, asn)

	// Simple BGP node structure (matching original)
	bgpNode := &BGPNode{
		Key:      bgpNodeKey,
		RouterID: bgpID, // Use BGP Router ID from peer message
		ASN:      asn,
	}

	// Create or update BGP node
	if _, err := uc.db.bgpNode.CreateDocument(ctx, bgpNode); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create BGP node: %w", err)
		}
		// Node already exists, which is fine (ignoreErrors: true in original)
	}

	glog.V(8).Infof("Ensured BGP node: %s (AS%d)", bgpID, asn)
	return nil
}

func (uc *UpdateCoordinator) checkIGPASNExists(ctx context.Context, asn uint32) (bool, error) {
	// Check if this ASN exists in IGP domain (matching original filter logic)
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.asn == @asn
		LIMIT 1
		RETURN node._key
	`, uc.db.config.IGPNode)

	bindVars := map[string]interface{}{
		"asn": asn,
	}

	cursor, err := uc.db.db.Query(ctx, query, bindVars)
	if err != nil {
		return false, fmt.Errorf("failed to query IGP ASNs: %w", err)
	}
	defer cursor.Close()

	return cursor.HasMore(), nil
}

func (uc *UpdateCoordinator) extractCapabilities(peerData map[string]interface{}, capField string) *bgp.Capability {
	// TODO: Implement BGP capability extraction
	// This would parse the capability data from the peer message
	// For now, return nil - capabilities are optional for graph topology
	return nil
}

func (uc *UpdateCoordinator) createBGPSessionEdges(ctx context.Context, sessionKey string, peerData map[string]interface{}, sessionType string) error {
	localBGPID, _ := peerData["local_bgp_id"].(string)
	remoteBGPID, _ := peerData["remote_bgp_id"].(string)
	localASN := getUint32FromInterface(peerData["local_asn"])
	remoteASN := getUint32FromInterface(peerData["remote_asn"])
	localIP, _ := peerData["local_ip"].(string)
	remoteIP, _ := peerData["remote_ip"].(string)

	// Determine source and target node IDs
	localNodeID, err := uc.getBGPNodeID(ctx, localBGPID, localASN, localIP)
	if err != nil {
		return fmt.Errorf("failed to get local node ID: %w", err)
	}

	remoteNodeID, err := uc.getBGPNodeID(ctx, remoteBGPID, remoteASN, remoteIP)
	if err != nil {
		return fmt.Errorf("failed to get remote node ID: %w", err)
	}

	// Create bidirectional session edges
	if err := uc.createSessionEdge(ctx, sessionKey, localNodeID, remoteNodeID, peerData, sessionType, true); err != nil {
		return fmt.Errorf("failed to create local->remote session edge: %w", err)
	}

	if err := uc.createSessionEdge(ctx, sessionKey, remoteNodeID, localNodeID, peerData, sessionType, false); err != nil {
		return fmt.Errorf("failed to create remote->local session edge: %w", err)
	}

	return nil
}

func (uc *UpdateCoordinator) createSessionEdge(ctx context.Context, sessionKey, fromNodeID, toNodeID string, peerData map[string]interface{}, sessionType string, isForward bool) error {
	// Create edge key
	direction := "fwd"
	if !isForward {
		direction = "rev"
	}
	edgeKey := fmt.Sprintf("%s_%s", sessionKey, direction)

	// Determine which graph collections to use based on IP version
	isIPv4, _ := peerData["is_ipv4"].(bool)

	var targetCollection driver.Collection
	if isIPv4 {
		targetCollection = uc.db.ipv4Graph
	} else {
		targetCollection = uc.db.ipv6Graph
	}

	// Extract session data
	localIP, _ := peerData["local_ip"].(string)
	remoteIP, _ := peerData["remote_ip"].(string)
	localASN := getUint32FromInterface(peerData["local_asn"])
	remoteASN := getUint32FromInterface(peerData["remote_asn"])

	// Create session edge object
	sessionEdge := &IPGraphObject{
		Key:           edgeKey,
		From:          fromNodeID,
		To:            toNodeID,
		LocalIP:       localIP,
		RemoteIP:      remoteIP,
		LocalNodeASN:  localASN,
		RemoteNodeASN: remoteASN,
		Protocol:      fmt.Sprintf("BGP_%s", sessionType),
	}

	// Create edge
	if _, err := targetCollection.CreateDocument(ctx, sessionEdge); err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create session edge: %w", err)
		}
		// Update existing edge
		if _, err := targetCollection.UpdateDocument(ctx, edgeKey, sessionEdge); err != nil {
			return fmt.Errorf("failed to update session edge: %w", err)
		}
	}

	glog.V(8).Infof("Created BGP session edge: %s (%s)", edgeKey, sessionType)
	return nil
}

func (uc *UpdateCoordinator) getBGPNodeID(ctx context.Context, bgpID string, asn uint32, ip string) (string, error) {
	// Check if this ASN exists in IGP domain
	igpASNExists, err := uc.checkIGPASNExists(ctx, asn)
	if err != nil {
		return "", err
	}

	if igpASNExists {
		// Return IGP node ID - look for IGP node with matching router_id or bgp_router_id
		query := fmt.Sprintf(`
			FOR node IN %s
			FILTER (node.router_id == @routerId OR node.bgp_router_id == @routerId) AND node.asn == @asn
			LIMIT 1
			RETURN node._id
		`, uc.db.config.IGPNode)

		bindVars := map[string]interface{}{
			"routerId": bgpID,
			"asn":      asn,
		}

		cursor, err := uc.db.db.Query(ctx, query, bindVars)
		if err != nil {
			return "", fmt.Errorf("failed to query IGP node ID: %w", err)
		}
		defer cursor.Close()

		if cursor.HasMore() {
			var nodeID string
			if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
				return "", fmt.Errorf("failed to read IGP node ID: %w", err)
			}
			return nodeID, nil
		}
	}

	// Return BGP node ID using original key format
	bgpNodeKey := fmt.Sprintf("%s_%d", bgpID, asn)
	return fmt.Sprintf("%s/%s", uc.db.config.BGPNode, bgpNodeKey), nil
}

func (uc *UpdateCoordinator) removeBGPSessionEdges(ctx context.Context, sessionKey string) error {
	// Remove both directions of the session edge
	edgeKeys := []string{
		fmt.Sprintf("%s_fwd", sessionKey),
		fmt.Sprintf("%s_rev", sessionKey),
	}

	collections := []driver.Collection{uc.db.ipv4Graph, uc.db.ipv6Graph}

	for _, collection := range collections {
		for _, edgeKey := range edgeKeys {
			if _, err := collection.RemoveDocument(ctx, edgeKey); err != nil {
				if !driver.IsNotFoundGeneral(err) {
					glog.Warningf("Failed to remove session edge %s from %s: %v", edgeKey, collection.Name(), err)
				}
			}
		}
	}

	glog.V(8).Infof("Removed BGP session edges for: %s", sessionKey)
	return nil
}

// Helper function to safely convert interface{} to uint32
func getUint32FromInterface(v interface{}) uint32 {
	switch val := v.(type) {
	case float64:
		return uint32(val)
	case uint32:
		return val
	case int:
		return uint32(val)
	case int64:
		return uint32(val)
	case string:
		if parsed, err := strconv.ParseUint(val, 10, 32); err == nil {
			return uint32(parsed)
		}
	}
	return 0
}
