package arangodb

import (
	"fmt"

	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// Helper functions to extract data from raw BMP messages using gobmp-arango key formats

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
	case bmp.PeerStateChangeMsg:
		return makePeerKey(bmpData)
	case bmp.UnicastPrefixV4Msg, bmp.UnicastPrefixV6Msg:
		return makeUnicastPrefixKey(bmpData)
	}
	return ""
}

// makeLSNodeKey replicates lsNodeArangoMessage.MakeKey()
func makeLSNodeKey(bmpData map[string]interface{}) string {
	protocolID, ok1 := bmpData["protocol_id"]
	domainID, ok2 := bmpData["domain_id"]
	igpRouterID, ok3 := bmpData["igp_router_id"].(string)

	if !ok1 || !ok2 || !ok3 {
		glog.V(8).Infof("Missing required fields for ls_node key: protocol_id=%v, domain_id=%v, igp_router_id=%s", ok1, ok2, ok3)
		return ""
	}

	// Default area_id to "0" unless protocol is OSPF
	areaID := "0"
	if proto, ok := protocolID.(float64); ok && (proto == 3 || proto == 6) { // OSPFv2 or OSPFv3
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

	if !ok1 || !ok2 || !ok3 {
		return ""
	}

	// Handle MTID - can be array or single object
	mtid := "0"
	if mtidTLV, exists := bmpData["mt_id_tlv"]; exists {
		if mtidArray, ok := mtidTLV.([]interface{}); ok && len(mtidArray) > 0 {
			if mtObj, ok := mtidArray[0].(map[string]interface{}); ok {
				if mt, ok := mtObj["mt_id"].(float64); ok {
					mtid = fmt.Sprintf("%.0f", mt)
				}
			}
		} else if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			if mt, ok := mtidObj["mt_id"].(float64); ok {
				mtid = fmt.Sprintf("%.0f", mt)
			}
		}
	}

	// Default area_id to "0" unless protocol is OSPF
	areaID := "0"
	if proto, ok := protocolID.(float64); ok && (proto == 3 || proto == 6) { // OSPFv2 or OSPFv3
		if area, ok := bmpData["area_id"].(string); ok {
			areaID = area
		}
	}

	// Get local and remote identifiers
	localLinkIP, _ := bmpData["local_link_ip"].(string)
	remoteLinkIP, _ := bmpData["remote_link_ip"].(string)
	localLinkID := "0"
	remoteLinkID := "0"

	if localID, ok := bmpData["local_link_id"].(float64); ok {
		localLinkID = fmt.Sprintf("%.0f", localID)
	}
	if remoteID, ok := bmpData["remote_link_id"].(float64); ok {
		remoteLinkID = fmt.Sprintf("%.0f", remoteID)
	}

	remoteIGPRouterID, _ := bmpData["remote_igp_router_id"].(string)

	return fmt.Sprintf("%v_%v_%s_%s_%s_%s_%s_%s_%s_%s",
		protocolID, domainID, areaID, igpRouterID, localLinkIP, localLinkID, remoteIGPRouterID, remoteLinkIP, remoteLinkID, mtid)
}

// makeLSPrefixKey replicates lsPrefixArangoMessage.MakeKey()
func makeLSPrefixKey(bmpData map[string]interface{}) string {
	protocolID, ok1 := bmpData["protocol_id"]
	domainID, ok2 := bmpData["domain_id"]
	igpRouterID, ok3 := bmpData["igp_router_id"].(string)
	prefix, ok4 := bmpData["prefix"].(string)
	prefixLen, ok5 := bmpData["prefix_len"].(float64)

	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return ""
	}

	// Handle MTID - can be array or single object
	mtid := "0"
	if mtidTLV, exists := bmpData["mt_id_tlv"]; exists {
		if mtidArray, ok := mtidTLV.([]interface{}); ok && len(mtidArray) > 0 {
			if mtObj, ok := mtidArray[0].(map[string]interface{}); ok {
				if mt, ok := mtObj["mt_id"].(float64); ok {
					mtid = fmt.Sprintf("%.0f", mt)
				}
			}
		} else if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			if mt, ok := mtidObj["mt_id"].(float64); ok {
				mtid = fmt.Sprintf("%.0f", mt)
			}
		}
	}

	// Default area_id to "0" unless protocol is OSPF
	areaID := "0"
	if proto, ok := protocolID.(float64); ok && (proto == 3 || proto == 6) { // OSPFv2 or OSPFv3
		if area, ok := bmpData["area_id"].(string); ok {
			areaID = area
		}
	}

	// Handle OSPF route type
	ospfRouteType := "0"
	if rt, ok := bmpData["ospf_route_type"].(float64); ok {
		ospfRouteType = fmt.Sprintf("%.0f", rt)
	}

	return fmt.Sprintf("%v_%v_%s_%s_%s_%s_%.0f_%s",
		protocolID, domainID, areaID, igpRouterID, mtid, prefix, prefixLen, ospfRouteType)
}

// makeLSSRv6SIDKey replicates lsSRv6SIDArangoMessage.MakeKey()
func makeLSSRv6SIDKey(bmpData map[string]interface{}) string {
	protocolID, ok1 := bmpData["protocol_id"]
	domainID, ok2 := bmpData["domain_id"]
	igpRouterID, ok3 := bmpData["igp_router_id"].(string)
	srv6SID, ok4 := bmpData["srv6_sid"].(string)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return ""
	}

	// Handle MTID - can be array or single object
	mtid := "0"
	if mtidTLV, exists := bmpData["mt_id_tlv"]; exists {
		if mtidArray, ok := mtidTLV.([]interface{}); ok && len(mtidArray) > 0 {
			if mtObj, ok := mtidArray[0].(map[string]interface{}); ok {
				if mt, ok := mtObj["mt_id"].(float64); ok {
					mtid = fmt.Sprintf("%.0f", mt)
				}
			}
		} else if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			if mt, ok := mtidObj["mt_id"].(float64); ok {
				mtid = fmt.Sprintf("%.0f", mt)
			}
		}
	}

	return fmt.Sprintf("%v_%v_%s_%s_%s", protocolID, domainID, igpRouterID, mtid, srv6SID)
}

// makePeerKey creates a key for BGP peer messages
// Replicates peerStateChangeArangoMessage.MakeKey(): RemoteBGPID + "_" + RemoteIP
func makePeerKey(bmpData map[string]interface{}) string {
	remoteBGPID, ok1 := bmpData["remote_bgp_id"].(string)
	remoteIP, ok2 := bmpData["remote_ip"].(string)

	if !ok1 || !ok2 {
		glog.V(8).Infof("Missing required fields for peer key: remote_bgp_id=%v, remote_ip=%v", ok1, ok2)
		return ""
	}

	return fmt.Sprintf("%s_%s", remoteBGPID, remoteIP)
}

// makeUnicastPrefixKey creates a key for unicast prefix messages
// Replicates unicastPrefixArangoMessage.MakeKey(): Prefix + "_" + PrefixLen + "_" + PeerIP
func makeUnicastPrefixKey(bmpData map[string]interface{}) string {
	prefix, ok1 := bmpData["prefix"].(string)
	prefixLen, ok2 := bmpData["prefix_len"].(float64)
	peerIP, ok3 := bmpData["peer_ip"].(string)

	if !ok1 || !ok2 || !ok3 {
		glog.V(8).Infof("Missing required fields for unicast prefix key: prefix=%v, prefix_len=%v, peer_ip=%v", ok1, ok2, ok3)
		return ""
	}

	return fmt.Sprintf("%s_%.0f_%s", prefix, prefixLen, peerIP)
}

// getBMPAction extracts the action from BMP data
func getBMPAction(bmpData map[string]interface{}) string {
	if action, ok := bmpData["action"].(string); ok && action != "" {
		return action
	}
	return "add" // Default action
}

// getBMPID constructs the document ID for ArangoDB
func getBMPID(bmpData map[string]interface{}, msgType dbclient.CollectionType) string {
	key := getBMPKeyForMessageType(bmpData, msgType)
	if key == "" {
		return ""
	}

	var collection string
	switch msgType {
	case bmp.LSNodeMsg:
		collection = "ls_node"
	case bmp.LSLinkMsg:
		collection = "ls_link"
	case bmp.LSPrefixMsg:
		collection = "ls_prefix"
	case bmp.LSSRv6SIDMsg:
		collection = "ls_srv6_sid"
	case bmp.PeerStateChangeMsg:
		collection = "peer"
	case bmp.UnicastPrefixV4Msg:
		collection = "unicast_prefix_v4"
	case bmp.UnicastPrefixV6Msg:
		collection = "unicast_prefix_v6"
	default:
		return ""
	}

	return fmt.Sprintf("%s/%s", collection, key)
}

// Helper function for debugging - lists available fields in BMP data
func getAvailableFields(bmpData map[string]interface{}) []string {
	fields := make([]string, 0, len(bmpData))
	for key := range bmpData {
		fields = append(fields, key)
	}
	return fields
}
