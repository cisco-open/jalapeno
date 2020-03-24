package handler

import (
        "strings"
        "fmt"
        "github.com/cisco-ie/jalapeno/processors/topology/database"
        "github.com/cisco-ie/jalapeno/processors/topology/openbmp"
)

func l3vpn(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        action             :=  m.GetStr("action")
        vpn_rd             :=  m.GetStr("vpn_rd")
        //vpnRD
        prefix             :=  m.GetStr("prefix")
        peer_ip            :=  m.GetStr("peer_ip")
        peer_asn           :=  m.GetStr("peer_asn")
        nexthop            :=  m.GetStr("nexthop")
        labels, ok         :=  m.GetInt("labels")
        ext_community_list :=  m.GetStr("ext_community_list")
        //extCommunityList
        is_ipv4            :=  m.GetStr("is_ipv4")
        prefix_len, ok     :=  m.GetInt("prefix_len")
        if !ok {
            prefix_len = 0
        }

        // handle BMP bug of nexthop being empty in IPv6 messages if nexthop == peer_ip
        if((is_ipv4 == "0") && (nexthop == "::")) {
            nexthop = peer_ip
        }

        real_vpn_rd := parse_vpn_rd(vpn_rd) // handles BMP bug of vpn_rds being incorrectly formated

        //l3vpn_prefix_key := vpn_rd + "_" + prefix + "_" + strconv.Itoa(prefix_len)
        //l3vpn_node_key := nexthop
        l3vpn_prefix_object := parse_l3vpn_prefix(a, real_vpn_rd, prefix, prefix_len, nexthop, peer_ip, peer_asn, labels, ext_community_list, is_ipv4)
        l3vpn_node_object := parse_l3vpn_node(a, real_vpn_rd, nexthop, peer_ip, peer_asn, ext_community_list)
        if (action == "add") {
            create_l3vpn_prefix(a, l3vpn_prefix_object)
            l3vpn_node_exists := a.db.CheckExistingL3VPNNode(nexthop)
            if (l3vpn_node_exists) {
                a.db.UpdateExistingVPNRDS(nexthop, real_vpn_rd)
            } else {
                create_l3vpn_node(a, l3vpn_node_object)
            }
        } else {
            delete_l3vpn_prefix(a, l3vpn_prefix_object)
            delete_l3vpn_node(a, l3vpn_node_object)
        }
}

// Parses VPN_RD from OpenBMP's L3VPN message's vpn_rd field. Currently the field has a bug. There is a repeat of 
// part of the vpn_rd at the beginning of the field, i.e. 99101:99. The 99 at the beginning should be removed.
func parse_vpn_rd(vpn_rd string) string {
    vpn_rd_split := strings.Split(vpn_rd, ":")
    vpn_rd_bug := vpn_rd_split[len(vpn_rd_split)-1]
    real_vpn_rd := strings.TrimPrefix(vpn_rd, vpn_rd_bug)
    return real_vpn_rd
}

// Deletes L3VPNPrefix document from L3VPNPrefix Collection
func delete_l3vpn_prefix(a *ArangoHandler, l3vpn_prefix_object *database.L3VPNPrefix) {
    if err := a.db.Delete(l3vpn_prefix_object); err != nil {
        fmt.Println("While deleting the current message's l3vpn_prefix document, encountered an error:", err)
    } else {
        fmt.Printf("Successfully deleted current message's l3vpn_prefix document: For Prefix: %q with RD: %q\n", l3vpn_prefix_object.Prefix, l3vpn_prefix_object.RD)
    }
}

// Deletes L3VPNNode document from L3VPNNode Collection
func delete_l3vpn_node(a *ArangoHandler, l3vpn_node_object *database.L3VPNNode) {
    if err := a.db.Delete(l3vpn_node_object); err != nil {
        fmt.Println("While deleting the current message's l3vpn_node document, encountered an error:", err)
    } else {
        fmt.Printf("Successfully deleted current message's l3vpn_node document: For Router: %q with RD: %q\n", l3vpn_node_object.RouterID, l3vpn_node_object.RD)
    }
}

// Creates L3VPNPrefix document in L3VPNPrefix Collection
func create_l3vpn_prefix(a *ArangoHandler, l3vpn_prefix_object *database.L3VPNPrefix) {
    if err := a.db.Upsert(l3vpn_prefix_object); err != nil {
        fmt.Println("While upserting the current message's l3vpn_prefix document, encountered an error:", err)
    } else {
        fmt.Printf("Successfully added current message's l3vpn_prefix document: For Prefix: %q with RD: %q\n", l3vpn_prefix_object.Prefix, l3vpn_prefix_object.RD)
    }
}

// Creates L3VPNNode document in L3VPNNode Collection
func create_l3vpn_node(a *ArangoHandler, l3vpn_node_object *database.L3VPNNode) {
    if err := a.db.Upsert(l3vpn_node_object); err != nil {
        fmt.Println("While upserting the current message's l3vpn_node document, encountered an error:", err)
    } else {
        fmt.Printf("Successfully added current message's l3vpn_node document: For Router: %q with RD: %q\n", l3vpn_node_object.RouterID, l3vpn_node_object.RD)
    }
}


// Parses L3VPNPrefix object from current OpenBMP-L3VPN message 
func parse_l3vpn_prefix(a *ArangoHandler, vpn_rd string, prefix string, prefix_len int, next_hop string, 
                        peer_ip string, peer_asn string, labels int, ext_community_list string, is_ipv4 string) *database.L3VPNPrefix {
        fmt.Printf("Parsing current message for l3vpn_prefix object: For Prefix: %q with RD: %q\n", prefix, vpn_rd)
        l3vpn_prefix_object := &database.L3VPNPrefix{
                RD:              vpn_rd,
                Prefix:          prefix,
                Length:          prefix_len,
                RouterID:        next_hop,
                ControlPlaneID:  peer_ip,
                ASN:             peer_asn,
                VPN_Label:       labels,
                ExtComm:         ext_community_list,
                IPv4:            is_ipv4,
        }
        return l3vpn_prefix_object
}

// Parses L3VPNNode object from current OpenBMP-L3VPN message 
func parse_l3vpn_node(a *ArangoHandler, vpn_rd string, nexthop string, peer_ip string, peer_asn string, ext_community_list string) *database.L3VPNNode{
        fmt.Printf("Parsing current message's l3vpn_node object: For Router: %q with RD: %q\n", nexthop, vpn_rd)
        var vpn_rd_list [] string
        vpn_rd_list = append(vpn_rd_list, vpn_rd)
        l3vpn_node_object := &database.L3VPNNode {
                RD:               vpn_rd_list,
                RouterID:         nexthop,
                ControlPlaneID:   peer_ip,
                ASN:              peer_asn,
                ExtComm:          ext_community_list,
        }
        if (nexthop != peer_ip) {
                fmt.Println("Note: OpenBMP-L3VPN message nexthop does not match peer_ip")
        }
        return l3vpn_node_object
}
