package arangodb

import (
	"fmt"

	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// Helper functions to extract data from raw BMP messages using gobmp-arango key formats

func getBMPKey(bmpData map[string]interface{}) string {
	// We need to determine the message type to construct the appropriate key
	// This is a simplified version - the actual key construction depends on message type

	// For debugging, let's first try the common fields
	if routerID, ok := bmpData["igp_router_id"].(string); ok && routerID != "" {
		protocolID, hasProtocol := bmpData["protocol_id"]
		domainID, hasDomain := bmpData["domain_id"]

		if hasProtocol && hasDomain {
			areaID := "0" // Default for non-OSPF
			if aID, ok := bmpData["area_id"].(string); ok {
				areaID = aID
			}

			// This matches lsNodeArangoMessage.MakeKey() format:
			// ProtocolID_DomainID_AreaID_IGPRouterID
			return fmt.Sprintf("%v_%v_%s_%s", protocolID, domainID, areaID, routerID)
		}
		return routerID
	}

	// Fallback: use router_hash if available
	if routerHash, ok := bmpData["router_hash"].(string); ok && routerHash != "" {
		return routerHash
	}

	// Last resort: use message key from Kafka
	if msgKey, ok := bmpData["_message_key"].(string); ok {
		return msgKey
	}

	glog.Warningf("Could not determine key from BMP data, available fields: %v", getAvailableFields(bmpData))
	return ""
}

// getBMPKeyForMessageType constructs keys using the exact same logic as gobmp-arango handlers
func getBMPKeyForMessageType(bmpData map[string]interface{}, msgType dbclient.CollectionType) string {
	switch msgType {
	case bmp.LSNodeMsg:
		return makeLSNodeKey(bmpData)
	case bmp.LSLinkMsg:
		return makeLSLinkKey(bmpData)
	case bmp.LSPrefixMsg:
		return makeLSPrefixKey(bmpData)
	case bmp.LSSRv6SIDMsg:
		return makeLSSRv6SIDKey(bmpData)
	default:
		return getBMPKey(bmpData) // Fallback to generic logic
	}
}

// makeLSNodeKey replicates lsNodeArangoMessage.MakeKey()
func makeLSNodeKey(bmpData map[string]interface{}) string {
	protocolID, ok1 := bmpData["protocol_id"]
	domainID, ok2 := bmpData["domain_id"]
	igpRouterID, ok3 := bmpData["igp_router_id"].(string)

	if !ok1 || !ok2 || !ok3 {
		return ""
	}

	areaID := "0"
	// For OSPF (OSPFv2=3, OSPFv3=6), use actual area_id, otherwise default to "0"
	if proto, ok := protocolID.(float64); ok && (proto == 3 || proto == 6) {
		if area, ok := bmpData["area_id"].(string); ok {
			areaID = area
		}
	}

	return fmt.Sprintf("%v_%v_%s_%s", protocolID, domainID, areaID, igpRouterID)
}

// makeLSLinkKey replicates lsLinkArangoMessage.MakeKey()
func makeLSLinkKey(bmpData map[string]interface{}) string {
	protocolID, ok1 := bmpData["protocol_id"]
	domainID, ok2 := bmpData["domain_id"]
	igpRouterID, ok3 := bmpData["igp_router_id"].(string)
	remoteIGPRouterID, ok4 := bmpData["remote_igp_router_id"].(string)
	areaID, ok5 := bmpData["area_id"].(string)

	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return ""
	}

	// Handle MTID
	mtid := 0
	if mtData, ok := bmpData["mt_id_tlv"]; ok {
		if mtArray, ok := mtData.([]interface{}); ok && len(mtArray) > 0 {
			if mtItem, ok := mtArray[0].(map[string]interface{}); ok {
				if mt, ok := mtItem["mt_id"].(float64); ok {
					mtid = int(mt)
				}
			}
		}
	}

	// Determine local and remote IDs
	var localID, remoteID string
	if localIP, ok := bmpData["local_link_ip"].(string); ok && localIP != "" {
		if remoteIP, ok := bmpData["remote_link_ip"].(string); ok && remoteIP != "" {
			localID = localIP
			remoteID = remoteIP
		}
	} else {
		// Use Link IDs in dotted notation for unnumbered links
		if localLinkID, ok := bmpData["local_link_id"].(float64); ok {
			id := uint32(localLinkID)
			localID = fmt.Sprintf("%d.%d.%d.%d",
				(id>>24)&0xff, (id>>16)&0xff, (id>>8)&0xff, id&0xff)
		}
		if remoteLinkID, ok := bmpData["remote_link_id"].(float64); ok {
			id := uint32(remoteLinkID)
			remoteID = fmt.Sprintf("%d.%d.%d.%d",
				(id>>24)&0xff, (id>>16)&0xff, (id>>8)&0xff, id&0xff)
		}
	}

	// Handle BGP case (protocol 7)
	routerID := igpRouterID
	remoteRouterID := remoteIGPRouterID
	if proto, ok := protocolID.(float64); ok && proto == 7 {
		if bgpID, ok := bmpData["bgp_router_id"].(string); ok {
			routerID = bgpID
		}
		if remoteBGPID, ok := bmpData["bgp_remote_router_id"].(string); ok {
			remoteRouterID = remoteBGPID
		}
	}

	return fmt.Sprintf("%v_%v_%d_%s_%s_%s_%s_%s",
		protocolID, domainID, mtid, areaID, routerID, localID, remoteRouterID, remoteID)
}

// makeLSPrefixKey replicates lsPrefixArangoMessage.MakeKey()
func makeLSPrefixKey(bmpData map[string]interface{}) string {
	protocolID, ok1 := bmpData["protocol_id"]
	domainID, ok2 := bmpData["domain_id"]
	areaID, ok3 := bmpData["area_id"].(string)
	prefix, ok4 := bmpData["prefix"].(string)
	prefixLen, ok5 := bmpData["prefix_len"]
	igpRouterID, ok6 := bmpData["igp_router_id"].(string)

	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 {
		return ""
	}

	mtid := 0
	if mtData, ok := bmpData["mt_id_tlv"]; ok {
		if mtArray, ok := mtData.([]interface{}); ok && len(mtArray) > 0 {
			if mtItem, ok := mtArray[0].(map[string]interface{}); ok {
				if mt, ok := mtItem["mt_id"].(float64); ok {
					mtid = int(mt)
				}
			}
		}
	}

	ospfRouteType := 0
	if rt, ok := bmpData["ospf_route_type"]; ok {
		if rtVal, ok := rt.(float64); ok {
			ospfRouteType = int(rtVal)
		}
	}

	return fmt.Sprintf("%v_%v_%d_%s_%d_%s_%v_%s",
		protocolID, domainID, mtid, areaID, ospfRouteType, prefix, prefixLen, igpRouterID)
}

// makeLSSRv6SIDKey replicates lsSRv6SIDArangoMessage.MakeKey()
func makeLSSRv6SIDKey(bmpData map[string]interface{}) string {
	domainID, ok1 := bmpData["domain_id"]
	igpRouterID, ok2 := bmpData["igp_router_id"].(string)
	srv6SID, ok3 := bmpData["srv6_sid"].(string)

	if !ok1 || !ok2 || !ok3 {
		return ""
	}

	return fmt.Sprintf("%v_%s_%s", domainID, igpRouterID, srv6SID)
}

func getBMPAction(bmpData map[string]interface{}) string {
	// The action field should be directly available in BMP data
	if action, ok := bmpData["action"].(string); ok && action != "" {
		// Normalize action values
		switch action {
		case "delete":
			return "del"
		case "add", "update", "del":
			return action
		default:
			glog.V(6).Infof("Unknown action '%s', defaulting to 'add'", action)
			return "add"
		}
	}

	// Fallback: check for withdrawal indicators (should rarely be needed with action field)
	if isWithdrawal, ok := bmpData["is_withdrawn"].(bool); ok && isWithdrawal {
		return "del"
	}

	if withdrawn, ok := bmpData["withdrawn"].(bool); ok && withdrawn {
		return "del"
	}

	// Default action if no explicit action found
	glog.V(6).Infof("No action field found in BMP data, defaulting to 'add'")
	return "add"
}

func getBMPID(bmpData map[string]interface{}, msgType dbclient.CollectionType) string {
	key := getBMPKey(bmpData)
	if key == "" {
		return ""
	}

	// Construct ID based on collection type
	var collectionName string
	switch msgType {
	case bmp.LSNodeMsg:
		collectionName = "ls_node"
	case bmp.LSLinkMsg:
		collectionName = "ls_link"
	case bmp.LSPrefixMsg:
		collectionName = "ls_prefix"
	case bmp.LSSRv6SIDMsg:
		collectionName = "ls_srv6_sid"
	default:
		collectionName = "unknown"
	}

	return fmt.Sprintf("%s/%s", collectionName, key)
}

// Helper function for debugging - lists available fields in BMP data
func getAvailableFields(bmpData map[string]interface{}) []string {
	fields := make([]string, 0, len(bmpData))
	for key := range bmpData {
		fields = append(fields, key)
	}
	return fields
}
