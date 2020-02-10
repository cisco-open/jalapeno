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
        peer_ip            :=  m.GetStr("peer_ip")
        peer_asn           :=  m.GetStr("peer_asn")
        nexthop            :=  m.GetStr("nexthop")
        labels             :=  m.GetStr("labels")
        ext_community_list :=  m.GetStr("ext_community_list")

        if !ok {
            prefix_len = 0
        }

        real_vpn_rd := parse_vpn_rd(vpn_rd)        
        parse_l3vpn_prefix(a, action, real_vpn_rd, prefix, prefix_len, peer_ip, peer_asn, nexthop, labels, ext_community_list)
        parse_l3vpn_router(a, action, real_vpn_rd, peer_ip, peer_asn, nexthop, ext_community_list)
        //parse_l3vpn_route()
}

// Parses VPN_RD from OpenBMP's L3VPN message's vpn_rd field. Currently the field has a bug. There is a repeat of 
// part of the vpn_rd at the beginning of the field, i.e. 99101:99. The 99 at the beginning should be removed.
func parse_vpn_rd(vpn_rd string) string {
    vpn_rd_split := strings.Split(vpn_rd, ":")
    vpn_rd_bug := vpn_rd_split[len(vpn_rd_split)-1]
    real_vpn_rd := strings.TrimPrefix(vpn_rd, vpn_rd_bug)
    return real_vpn_rd
}

// Parses a L3VPN Prefix from the current L3VPN OpenBMP Message
// Upserts the created L3VPN Prefix document into the "L3VPN_Prefixes" collection
func parse_l3vpn_prefix(a *ArangoHandler, action string, vpn_rd string, prefix string, prefix_len int, peer_ip string,
                        peer_asn string, nexthop string, labels string, ext_community_list string) {
        fmt.Println("Parsing L3VPN - document: l3vpn_prefix_document")
        fmt.Printf("Parsing current L3VPN message's l3vpn_prefix document: For Prefix: %q with RD: %q\n", prefix, vpn_rd)
        l3vpn_prefix_document := &database.L3VPN_Prefix{
                RD:              vpn_rd,
                Prefix:          prefix,
                Length:          prefix_len,
                RouterIP:        nexthop,
                ASN:             peer_asn,
                AdvertisingPeer: peer_ip,
                VPN_Label:       labels,
                ExtComm:         ext_community_list,
        }

    if (action == "del") {
        if err := a.db.Delete(l3vpn_prefix_document); err != nil {
                    fmt.Println("While deleting the current L3VPN message's l3vpn_prefix document, encountered an error:", err)
            } else {
                    fmt.Printf("Successfully deleted current L3VPN message's l3vpn_prefix document: For Prefix: %q with RD: %q\n", prefix, vpn_rd)
            }       
    } else {
    	if err := a.db.Upsert(l3vpn_prefix_document); err != nil {
                    fmt.Println("While upserting the current L3VPN message's l3vpn_prefix document, encountered an error:", err)
            } else {
                    fmt.Printf("Successfully added current L3VPN message's l3vpn_prefix document: For Prefix: %q with RD: %q\n", prefix, vpn_rd)
            }
    }
}

// Parses a L3VPN Router from the current L3VPN OpenBMP Message
// Upserts the created L3VPN Router document into the "L3VPN_Router" collection
func parse_l3vpn_router(a *ArangoHandler, action string, vpn_rd string, peer_ip string, 
                        peer_asn string, nexthop string, ext_community_list string) {
        fmt.Println("Parsing L3VPN - document: l3vpn_router_document")
        fmt.Printf("Parsing current L3VPN message's l3vpn_router document: For Router: %q with RD: %q\n", peer_ip, vpn_rd)

        var vpn_rd_list [] string
        vpn_rd_list = append(vpn_rd_list, vpn_rd)
        l3vpn_router_exists := a.db.CheckExistingL3VPNRouter(nexthop)
        if (l3vpn_router_exists) {
            //tempVariable := a.db.GetExistingVPNRDS(nexthop)
            //print(tempVariable)
            a.db.UpdateExistingVPNRDS(nexthop, vpn_rd)
        } else {
            l3vpn_router_document := &database.L3VPN_Router{
                    RD:               vpn_rd_list,
                    RouterIP:         nexthop,
                    ASN:              peer_asn,
                    AdvertisingPeer:  peer_ip,
                    ExtComm:          ext_community_list,
            }

            if (nexthop != peer_ip) {
                    fmt.Println("Note: L3VPN message nexthop does not match L3VPN message peer_ip")
            }

            if (action == "del") {
                if err := a.db.Delete(l3vpn_router_document); err != nil {
                    fmt.Println("While deleting the current L3VPN message's l3vpn_router document, encountered an error:", err)
                } else {
                    fmt.Printf("Successfully deleted current L3VPN message's l3vpn_router document: For Router: %q with RD: %q\n", nexthop, vpn_rd)
                }
            } else {
                if err := a.db.Upsert(l3vpn_router_document); err != nil {
                    fmt.Println("While upserting the current L3VPN message's l3vpn_router document, encountered an error:", err)
                } else {
                    fmt.Printf("Successfully added current L3VPN message's l3vpn_router document: For Router: %q with RD: %q\n", nexthop, vpn_rd)
                }
            }
        }
}
