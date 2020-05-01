package arangodb

import (
        "github.com/golang/glog"
        "github.com/sbezverk/gobmp/pkg/message"
        "github.com/cisco-ie/jalapeno/processors/topology/pkg/database"
)

func (a *arangoDB) l3vpnHandler(obj *message.L3VPNPrefix) {
        action := obj.Action
        db := a.GetArangoDBInterface()

        l3vpnPrefixDocument := &database.L3VPNPrefix{
                RD:              obj.VPNRD,
                Prefix:          obj.Prefix,
                Length:          obj.PrefixLen,
                RouterID:        obj.Nexthop,
                ControlPlaneID:  obj.PeerIP,
                ASN:             obj.PeerASN,
                VPN_Label:       obj.Labels,
                ExtComm:         obj.BaseAttributes.ExtCommunityList,
                IPv4:            obj.IsIPv4,
        }

        l3vpnNodeDocument := &database.L3VPNNode {
                RD:              []string{obj.VPNRD},
                RouterID:        obj.Nexthop,
                ControlPlaneID:  obj.PeerIP,
                ASN:             obj.PeerASN,
                ExtComm:         obj.BaseAttributes.ExtCommunityList,
        }

        // TODO (no IPv6 handling in GoBMP Collector):
        //handle BMP bug of nexthop being empty in IPv6 messages if nexthop == peer_ip
        //if((obj.IsIPv4 == false) && (obj.Nexthop == "::")) {
            //routerID = controlPlaneID
        //}

       handleL3VPNPrefixDocument(l3vpnPrefixDocument, action, db)
       handleL3VPNNodeDocument(l3vpnNodeDocument, action, db)
}

func handleL3VPNPrefixDocument(l3vpnPrefixDocument *database.L3VPNPrefix, action string, db *database.ArangoConn) {
        if (action == "add") {
                if err := db.Upsert(l3vpnPrefixDocument); err != nil {
                        glog.Errorf("Encountered an error while upserting the l3vpn prefix document: %+v", err)
                        return
                }
                glog.Infof("Successfully added l3vpn prefix document with prefix: %q with RD: %q\n", l3vpnPrefixDocument.Prefix, l3vpnPrefixDocument.RD)

        } else {
                if err := db.Delete(l3vpnPrefixDocument); err != nil {
                        glog.Errorf("Encountered an error while deleting the l3vpn prefix document: %+v", err)
                        return
                } else {
                        glog.Infof("Successfully deleted l3vpn prefix document with prefix: %q with RD: %q\n", l3vpnPrefixDocument.Prefix, l3vpnPrefixDocument.RD)
                }
        }
}

func handleL3VPNNodeDocument(l3vpnNodeDocument *database.L3VPNNode, action string, db *database.ArangoConn) {
        if (action == "add") {
                l3vpnNodeExists := db.CheckExistingL3VPNNode(l3vpnNodeDocument.RouterID)
                if(l3vpnNodeExists) {
                        db.UpdateExistingVPNRDS(l3vpnNodeDocument.RouterID, l3vpnNodeDocument.RD[0])
                } else {
                        if err := db.Upsert(l3vpnNodeDocument); err != nil {
                                glog.Errorf("Encountered an error while upserting the l3vpn node document: %+v", err)
                                return
                        }
                        glog.Infof("Successfully added l3vpn node document for router: %q with RD: %q\n", l3vpnNodeDocument.RouterID, l3vpnNodeDocument.RD)
               }

        } else {
                if err := db.Delete(l3vpnNodeDocument); err != nil {
                        glog.Errorf("Encountered an error while deleting the l3vpn node document: %+v", err)
                        return
                } else {
                        glog.Infof("Successfully deleted l3vpn node document for router: %q with RD: %q\n", l3vpnNodeDocument.RouterID, l3vpnNodeDocument.RD)
                }
        }
}

