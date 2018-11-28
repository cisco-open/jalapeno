package database

import (
	"fmt"
	"strings"
	"strconv"
)

func (a *ArangoConn) GetRouterByIP(ip string) *Router {
	r := &Router{}
	q := fmt.Sprintf("FOR r in Routers FILTER r.RouterIP == %q RETURN r", ip)
	results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
		return results[len(results)-1].(*Router)
	}
	return nil
}

func (a *ArangoConn) GetSRBeginningLabel(ip string) int {
        if len(ip) == 0 {
                return 0
        }
        var r string
        q := fmt.Sprintf("FOR r in Routers FILTER r.BGPID == %q AND r.SRGB != null RETURN r.SRGB", ip)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
                srgb := results[len(results)-1].(string)
	        srgb_split := strings.Split(srgb, ", ")
        	sr_beginning_label := srgb_split[0]
	        sr_beginning_label_val, _ := strconv.ParseInt(sr_beginning_label, 10, 0)
		return int(sr_beginning_label_val)
        }
        return 0 
}


func (a *ArangoConn) GetSIDIndex(ip string) string {
        if len(ip) == 0 {
                return ""
        }
        var r string
        q := fmt.Sprintf("FOR r in Routers FILTER r.BGPID == %q AND r.NodeSIDIndex != null RETURN r.NodeSIDIndex", ip)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
                sid_index := results[len(results)-1].(string)
		return sid_index
        }
        return ""
}

func (a *ArangoConn) GetExternalRouterIP(external_router_intf_ip string) string {
        if len(external_router_intf_ip) == 0 {
                return ""
        }
        var r string
        q := fmt.Sprintf("FOR e in ExternalRouterInterfaces FILTER e.RouterInterfaceIP == %q return e.RouterIP", external_router_intf_ip)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
                external_router_ip := results[len(results)-1].(string)
		return external_router_ip
        }
        return ""
}


func (a *ArangoConn) GetRouterKeyFromInterfaceIP(ip string) string {
	if len(ip) == 0 {
		return ""
	}
	var r string
	key := "Routers/" + ip
	col := InternalLinkEdgeName
	q := fmt.Sprintf("FOR e in %s Filter e.ToIP == %q OR e._to == %q  RETURN DISTINCT e._to", col, ip, key)
	results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
		return results[len(results)-1].(string)
	}
	return ""

}

func(a *ArangoConn) GetExternalPrefixEdgeKeysFromInterface(ip string, interface_ip string) []string {
	var edges []string
	if len(ip) == 0 || len(interface_ip) == 0 {
                return edges
        }
	var r string
	q := fmt.Sprintf("For e in ExtPrefixEdges FILTER e._from == %q AND e.InterfaceIP == %q return e._key", ip, interface_ip)
	results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
		for index := range results {
			current_edge := results[index].(string)
			edges = append(edges, current_edge)
		}
		return edges
	}
	return edges
}

func(a *ArangoConn) CheckIfInternal(ip string) bool {
	if len(ip) == 0 {
		return false
	}
        var is_internal bool = false
	var r string
        q := fmt.Sprintf("FOR i in InternalRouters FILTER i.BGPID == %q RETURN { BGPID: i.BGPID }", ip)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
		is_internal = true
        }
	return is_internal
}

func(a *ArangoConn) CheckIfEgress(ip string) bool {
	if len(ip) == 0 {
		return false
	}
	var is_egress bool = false
	var r string
        q := fmt.Sprintf("FOR e in ExternalPeers FILTER e.BGPID == %q RETURN DISTINCT { BGPID: e.BGPID }", ip)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
		is_egress = true
        }
	return is_egress
}


func (a *ArangoConn) GetExternalPrefixEdgeData() []map[string]interface{} {
        var r string

        q := fmt.Sprintf("FOR e in InternalRouter FOR l in ExternalPeer FILTER e.SRNodeSID != NULL AND e.BGPID == l.BGPID RETURN { BGPID: e.BGPID, SRNodeSID: e.SRNodeSID, ExtPeer: l.ExtPeer }")
        result, _ := a.Query(q, nil, r)
//        fmt.Println(results)
//        fmt.Println(len(results))

        egress_prefix_edge_slice := make([]map[string]interface{}, 0)
        for index := range result {        
	        // fmt.Println(result[index])
        	current_result := result[index].(map[string]interface{})
                egress_prefix_edge_slice = append(egress_prefix_edge_slice, current_result)
	        // fmt.Println(current_result["ExtPeer"])
	}
        //fmt.Println("We got egress_prefix_edge_slice")
        //fmt.Println(egress_prefix_edge_slice)
        return egress_prefix_edge_slice
//        if len(results) > 0 {
//            return results
//        }
 //       return "" 
}


func (a *ArangoConn) CreateInternalLinkEdge(edge_from string, edge_to string, edge_src_intf string, edge_dst_intf string, edge_protocol string, edge_label string) {
        var r string
	edge_key := edge_from + "_" + edge_src_intf + "_" + edge_dst_intf + "_" + edge_to
        edge_from = "Routers/" + edge_from
        edge_to = "Routers/" + edge_to
	
	q := fmt.Sprintf("INSERT { _key: %q, _from: %q, _to: %q, SrcInterfaceIP: %q, DstInterfaceIP: %q, Protocol: %q, Label: %q } into InternalLinkEdges RETURN { after: NEW }", edge_key, edge_from, edge_to, edge_src_intf, edge_dst_intf, edge_protocol, edge_label)
        results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
                fmt.Println("Created InternalLinkEdge:", edge_key, "with Label:", edge_label)
        } else {
                fmt.Println("InternalLinkEdge was not created.")
        }
}

func (a *ArangoConn) CreateExternalLinkEdge(edge_from string, edge_to string, edge_src_intf string, edge_dst_intf string, edge_protocol string, edge_label string) {
        var r string
	edge_key := edge_from + "_" + edge_src_intf + "_" + edge_dst_intf + "_" + edge_to
        edge_from = "Routers/" + edge_from
        edge_to = "Routers/" + edge_to
	
	q := fmt.Sprintf("INSERT { _key: %q, _from: %q, _to: %q, Source: %q, SrcInterfaceIP: %q, Destination: %q, DstInterfaceIP: %q, Protocol: %q, Label: %q } into ExternalLinkEdges RETURN { after: NEW }", edge_key, edge_from, edge_to, edge_from, edge_src_intf, edge_to, edge_dst_intf, edge_protocol, edge_label)
        results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
                fmt.Println("Created ExternalLinkEdge:", edge_key, "with Label:", edge_label)
        } else {
                fmt.Println("ExternalLinkEdge was not created.")
        }
}

func (a *ArangoConn) UpdateExternalLinkEdge(edge_from string, edge_to string, edge_src_intf string, edge_dst_intf string, edge_protocol string, edge_label string) {
        var r string
	edge_key := edge_from + "_" + edge_src_intf + "_" + edge_dst_intf + "_" + edge_to
        edge_from = "Routers/" + edge_from
        edge_to = "Routers/" + edge_to
	
	q := fmt.Sprintf("UPDATE { _key: %q, _from: %q, _to: %q, Source: %q, SrcInterfaceIP: %q, Destination: %q, DstInterfaceIP: %q, Protocol: %q, Label: %q } into ExternalLinkEdges RETURN { after: NEW }", edge_key, edge_from, edge_to, edge_from, edge_src_intf, edge_to, edge_dst_intf, edge_protocol, edge_label)
        results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
                fmt.Println("Updated ExternalLinkEdge:", edge_key, "with Label:", edge_label)
        } else {
                fmt.Println("ExternalLinkEdge was not updated.")
        }
}

func(a *ArangoConn) CheckExternalLinkEdgeExists(key string) bool {
	if len(key) == 0 {
		return false
	}
        var external_link_edge_exists bool = false
	var r string
        q := fmt.Sprintf("FOR e in ExternalLinkEdges FILTER e._key == %q RETURN { key: e._key }", key)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
		external_link_edge_exists = true
        }
	return external_link_edge_exists
}


func (a *ArangoConn) CreateExternalPrefixEdge(peer_ip string, peer_asn string, peer_intf_ip string, prefix_ip string, prefix_asn string, prefix_length int) {
	edge_from := ""
	if(peer_ip != "") {
		edge_from = "Routers/" + peer_ip
	} else {
		edge_from = "Routers/" + peer_intf_ip
	}
	edge_to := "Prefixes/" + prefix_ip
	edge_key := peer_asn + peer_intf_ip + strings.Replace(edge_to, "/", "_", -1)

        var r string
	q := fmt.Sprintf("INSERT { _key: %q, _from: %q, _to: %q, SrcRouterIP: %q, SrcRouterASN: %q, SrcIntfIP: %q, DstPrefix: %q, DstPrefixASN: %q } into ExternalPrefixEdges RETURN { after: NEW }", edge_key, edge_from, edge_to, peer_ip, peer_asn, peer_intf_ip, prefix_ip, prefix_asn)
        fmt.Println(q)
        results, _ := a.Query(q, nil, r)
	if len(results) > 0 {
                fmt.Println("Created ExternalPrefixEdge:", edge_key)
        } else {
                fmt.Println("ExternalPrefixEdge was not created.")
        }
}

func (a *ArangoConn) GetExternalPrefixEdgeKeysFromInterfaceAndASN(interface_ip string, asn string) []string {
        var edge_keys []string
        if interface_ip == "" || asn == "" {
                return edge_keys
        }
	var r string
        q := fmt.Sprintf("FOR e in ExternalPrefixEdges FILTER e.SrcIntfIP == %q AND e.SrcRouterASN == %q RETURN DISTINCT e._key", interface_ip, asn)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
                for index := range results {
                        current_key := results[index].(string)
                        edge_keys = append(edge_keys, current_key)
                }
                return edge_keys
        }
	return edge_keys
}

func (a *ArangoConn) GetExternalPrefixEdgeSrcIP(key string) string {
	fmt.Printf("Collecting ExternalPrefixEdge SrcRouterIP for ExternalPrefixEdge key: %q", key)
	src_ip := ""
	var r string
        q := fmt.Sprintf("FOR e in ExternalPrefixEdges FILTER e._key == %q RETURN e.SrcRouterIP", key)
        results, _ := a.Query(q, nil, r)
	fmt.Printf("ExternalPrefixEdge SrcRouterIP is: %q", results)
	if len(results) > 0 {
		src_ip = results[len(results)-1].(string)
	}
	return src_ip
}

func (a *ArangoConn) UpdateExternalPrefixEdgeSrcIP(key string, src_router_ip string) {
        var r string
	edge_from := "Routers/" + src_router_ip

	q := fmt.Sprintf("FOR e in ExternalPrefixEdges FILTER e._key == %q UPDATE { _key: e._key, _from: %q, SrcRouterIP: %q } IN ExternalPrefixEdges RETURN { before: OLD, after: NEW }", key, edge_from, src_router_ip)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
                fmt.Println("Updated ExternalPrefixEdge: %q with from field: %q and SrcRouterIP: %q", key, edge_from, src_router_ip)
        } else {
                fmt.Println("ExternalPrefixEdge was not updated with SrcRouterIP -- something may have gone wrong")
        }
}

func (a *ArangoConn) CreateExternalPrefixEdgeSource(router_ip string, router_asn string, router_intf_ip string) {
	// we need to get all ExternalPrefixEdge records where asn = router_asn and intf_ip = router_intf_ip
	// for each that exists, if SrcRouterIP is empty: update "_from" to be "router_ip" and update "SrcRouterIP" to be "router_ip"
	fmt.Printf("Collecting ExternalPrefixEdge keys for SrcIntfIP: %q and SrcASN: %q\n", router_intf_ip, router_asn)

        if router_intf_ip == "" || router_asn == "" {
                return
        }
	var r string
        q := fmt.Sprintf("FOR e in ExternalPrefixEdges FILTER e.SrcIntfIP == %q AND e.SrcRouterASN == %q RETURN DISTINCT e._key", router_intf_ip, router_asn)
        results, _ := a.Query(q, nil, r)
        if len(results) > 0 {
                for index := range results {
                        current_key := results[index].(string)
			fmt.Println("Current ExternalPrefixEdge key:", current_key)
		        fmt.Printf("Collecting ExternalPrefixEdge SrcRouterIP for ExternalPrefixEdge key: %q\n", current_key)
		        src_ip := ""
		        q2 := fmt.Sprintf("FOR e in ExternalPrefixEdges FILTER e._key == %q RETURN e.SrcRouterIP", current_key)
		        results2, _ := a.Query(q2, nil, r)
		       	fmt.Printf("ExternalPrefixEdge SrcRouterIP is: %q\n", results2)
		        if len(results2) > 0 {
		               	src_ip = results2[len(results2)-1].(string)
		        }
			external_prefix_edge_src_ip := src_ip
			fmt.Println("Current ExternalPrefixEdge SrcRouterIP:", external_prefix_edge_src_ip)
			if(external_prefix_edge_src_ip == "") {
			        edge_from := "Routers/" + router_ip
				q3 := fmt.Sprintf("FOR e in ExternalPrefixEdges FILTER e._key == %q UPDATE { _key: e._key, _from: %q, SrcRouterIP: %q } IN ExternalPrefixEdges RETURN { before: OLD, after: NEW }", current_key, edge_from, router_ip)			
			        results3, _ := a.Query(q3, nil, r)
			        if len(results3) > 0 {
			                fmt.Printf("Updated ExternalPrefixEdge: %q with from field: %q and SrcRouterIP: %q\n", current_key, edge_from, router_ip)
			        } else {
			                fmt.Println("ExternalPrefixEdge was not updated with SrcRouterIP -- something may have gone wrong")
			        }
			} else {
				fmt.Println("ExternalPrefixEdge SrcRouterIP already filled -- moving on")
			}
                }
        }
}

/*
func (a *ArangoConn) CreateExternalPrefixEdgeSource(router_ip string, router_asn string, router_intf_ip string) {
	// we need to get all ExternalPrefixEdge records where asn = router_asn and intf_ip = router_intf_ip
	// for each that exists, if SrcRouterIP is empty: update "_from" to be "router_ip" and update "SrcRouterIP" to be "router_ip"
	fmt.Printf("Collecting ExternalPrefixEdge keys for SrcIntfIP: %q and SrcASN: %q", router_intf_ip, router_asn)
	external_prefix_edge_keys := GetExternalPrefixEdgeKeysFromInterfaceAndASN(router_intf_ip, router_asn)
	fmt.Println("Collected the following keys:", external_prefix_edge_keys)
	if len(external_prefix_edge_keys) > 0 {
		for current_key := range external_prefix_edge_keys {
			fmt.Println("Current ExternalPrefixEdge key:", current_key)
			external_prefix_edge_src_ip := GetExternalPrefixEdgeSrcIP(current_key)
			fmt.Println("Current ExternalPrefixEdge SrcRouterIP:", external_prefix_edge_src_ip)
			if(external_prefix_edge_src_ip == "") {
				UpdateExternalPrefixEdgeSrcIP(current_key, router_ip)
			} else {
				fmt.Println("ExternalPrefixEdge SrcRouterIP already filled -- moving on")
			}
		}
	}
}
*/


