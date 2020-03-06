package handler

import (
    "strings"
    "fmt"
    "wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/database"
    "wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/openbmp"
)

func l3vpn(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        action             :=  m.GetStr("action")
        vpn_rd             :=  m.GetStr("vpn_rd")
        prefix             :=  m.GetStr("prefix")
        prefix_len, ok     :=  m.GetInt("prefix_len")
        if !ok {
            prefix_len = 0
        }
        peer_ip            :=  m.GetStr("peer_ip")
        peer_asn           :=  m.GetStr("peer_asn")
        nexthop            :=  m.GetStr("nexthop")
        labels, ok         :=  m.GetInt("labels")
        ext_community_list :=  m.GetStr("ext_community_list")

        real_vpn_rd := parse_vpn_rd(vpn_rd)        
        parse_l3vpn_prefix(a, action, real_vpn_rd, prefix, prefix_len, peer_ip, peer_asn, labels, ext_community_list)
        parse_l3vpn_node(a, action, real_vpn_rd, peer_ip, peer_asn, nexthop, ext_community_list)
}

// Parses VPN_RD from OpenBMP's L3VPN message's vpn_rd field. Currently the field has a bug. There is a repeat of 
// part of the vpn_rd at the beginning of the field, i.e. 99101:99. The 99 at the beginning should be removed.
func parse_vpn_rd(vpn_rd string) string {
    vpn_rd_split := strings.Split(vpn_rd, ":")
    vpn_rd_bug := vpn_rd_split[len(vpn_rd_split)-1]
    real_vpn_rd := strings.TrimPrefix(vpn_rd, vpn_rd_bug)
    return real_vpn_rd
}

// Parses a L3VPNPrefix from the current L3VPN OpenBMP Message
// Upserts the created L3VPNPrefix document into the "L3VPNPrefix" collection
func parse_l3vpn_prefix(a *ArangoHandler, action string, vpn_rd string, prefix string, prefix_len int, peer_ip string,
                        peer_asn string, labels int, ext_community_list string) {
        fmt.Println("Parsing L3VPNPrefix - document: l3vpn_prefix_document")
        fmt.Printf("Parsing current OpenBMP-L3VPN message for l3vpn_prefix document: For Prefix: %q with RD: %q\n", prefix, vpn_rd)
        l3vpn_prefix_document := &database.L3VPNPrefix{
                RD:              vpn_rd,
                Prefix:          prefix,
                Length:          prefix_len,
                RouterID:        peer_ip,
                ASN:             peer_asn,
                VPN_Label:       labels,
                ExtComm:         ext_community_list,
        }

    if (action == "del") {
        if err := a.db.Delete(l3vpn_prefix_document); err != nil {
                    fmt.Println("While deleting the current OpenBMP-L3VPN message's l3vpn_prefix document, encountered an error:", err)
            } else {
                    fmt.Printf("Successfully deleted current OpenBMP-L3VPN message's l3vpn_prefix document: For Prefix: %q with RD: %q\n", prefix, vpn_rd)
            }       
    } else {
    	if err := a.db.Upsert(l3vpn_prefix_document); err != nil {
                    fmt.Println("While upserting the current OpenBMP-L3VPN message's l3vpn_prefix document, encountered an error:", err)
            } else {
                    fmt.Printf("Successfully added current OpenBMP-L3VPN message's l3vpn_prefix document: For Prefix: %q with RD: %q\n", prefix, vpn_rd)
            }
    }
}

// Parses a L3VPN-Node from the current L3VPN OpenBMP Message
// Upserts the created L3VPN-Node document into the "L3VPNNode" collection
func parse_l3vpn_node(a *ArangoHandler, action string, vpn_rd string, peer_ip string, peer_asn string, nexthop string, ext_community_list string) {
        fmt.Println("Parsing OpenBMP-L3VPN - document: l3vpn_node_document")
        fmt.Printf("Parsing current OpenBMP-L3VPN message's l3vpn_node document: For Router: %q with RD: %q\n", peer_ip, vpn_rd)

        var vpn_rd_list [] string
        vpn_rd_list = append(vpn_rd_list, vpn_rd)
        l3vpn_node_exists := a.db.CheckExistingL3VPNNode(peer_ip)
        if (l3vpn_node_exists) {
            a.db.UpdateExistingVPNRDS(peer_ip, vpn_rd)
        } else {
            l3vpn_node_document := &database.L3VPNNode {
                    RD:               vpn_rd_list,
                    RouterID:         peer_ip,
                    ASN:              peer_asn,
                    ExtComm:          ext_community_list,
            }

            if (nexthop != peer_ip) {
                    fmt.Println("Note: OpenBMP-L3VPN message nexthop does not match OpenBMP-L3VPN message peer_ip")
            }

            if (action == "del") {
                if err := a.db.Delete(l3vpn_node_document); err != nil {
                    fmt.Println("While deleting the current OpenBMP-L3VPN message's l3vpn_node document, encountered an error:", err)
                } else {
                    fmt.Printf("Successfully deleted current OpenBMP-L3VPN message's l3vpn_node document: For Router: %q with RD: %q\n", peer_ip, vpn_rd)
                }
            } else {
                if err := a.db.Upsert(l3vpn_node_document); err != nil {
                    fmt.Println("While upserting the current OpenBMP-L3VPN message's l3vpn_node document, encountered an error:", err)
                } else {
                    fmt.Printf("Successfully added current OpenBMP-L3VPN message's l3vpn_node document: For Router: %q with RD: %q\n", peer_ip, vpn_rd)
                }
            }
        }
}
