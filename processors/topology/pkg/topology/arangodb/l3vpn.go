package arangodb

import (
	"github.com/golang/glog"
	"github.com/jalapeno-sdn/jalapeno/pkg/topology/database"
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/srv6"
        "strings"
)

func (a *arangoDB) l3vpnHandler(obj *message.L3VPNPrefix) {
	action := obj.Action
	db := a.GetArangoDBInterface()

	var extCommunityList []string
	if obj.BaseAttributes != nil {
		extCommunityList = obj.BaseAttributes.ExtCommunityList
	}
	var l3Service *srv6.L3Service
	if obj.PrefixSID != nil {
		if obj.PrefixSID.SRv6L3Service != nil {
			l3Service = obj.PrefixSID.SRv6L3Service
		}
	}

	l3vpnPrefixDocument := &database.L3VPNPrefix{
		RD:             obj.VPNRD,
		Prefix:         obj.Prefix,
		Length:         obj.PrefixLen,
		RouterID:       obj.Nexthop,
		ControlPlaneID: obj.PeerIP,
		ASN:            obj.PeerASN,
		VPN_Label:      obj.Labels,
		SRv6_SID:       l3Service,
		ExtComm:        extCommunityList,
		IPv4:           obj.IsIPv4,
		Origin_AS:      obj.OriginAS,
	}

	l3vpnNodeDocument := &database.L3VPNNode{
		RD:             []string{obj.VPNRD},
		RouterID:       obj.Nexthop,
		ControlPlaneID: obj.PeerIP,
		ASN:            obj.PeerASN,
		ExtComm:        extCommunityList,
	}

        // Processing for L3VPN RT collection
	extComm := obj.BaseAttributes.ExtCommunityList
        prefix := obj.Prefix

	for _, ext := range extComm {
               if !strings.HasPrefix(ext, "rt=") {
                       continue
                }
                rt := strings.TrimPrefix(ext, "rt=")
                // glog.Infof("for prefix key: %s found route target: %s", key, rt)

		rtExists := db.CheckExistingL3VPNRT(rt)
		if rtExists {
			db.UpdateExistingL3VPNRT(rt, prefix)
		} else {

			l3vpnRtDocument := &database.L3VPNRT{
				RT:       rt,
				Prefix: []string{obj.Prefix},
				//Length: length,
			}

			handleL3VPNRTDocument(l3vpnRtDocument, action, db)
		}
	}

	handleL3VPNPrefixDocument(l3vpnPrefixDocument, action, db)
        handleL3VPNNodeDocument(l3vpnNodeDocument, action, db)
}

func handleL3VPNPrefixDocument(l3vpnPrefixDocument *database.L3VPNPrefix, action string, db *database.ArangoConn) {
	if action == "add" {
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
	if action == "add" {
		l3vpnNodeExists := db.CheckExistingL3VPNNode(l3vpnNodeDocument.RouterID)
		if l3vpnNodeExists {
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

func handleL3VPNRTDocument(l3vpnRtDocument *database.L3VPNRT, action string, db *database.ArangoConn) {
        if action == "add" {
                l3vpnRtExists := db.CheckExistingL3VPNRT(l3vpnRtDocument.RT)
                if l3vpnRtExists {
                        db.UpdateExistingL3VPNRT(l3vpnRtDocument.RT, l3vpnRtDocument.Prefix[0])
                } else {
                        if err := db.Upsert(l3vpnRtDocument); err != nil {
                                glog.Errorf("Encountered an error while upserting the l3vpn rt document: %+v", err)
                                return
                        }
                        glog.Infof("Successfully added l3vpn rt document for RT: %q with Prefix: %q\n", l3vpnRtDocument.RT, l3vpnRtDocument.Prefix)
                }

        } else {
                if err := db.Delete(l3vpnRtDocument); err != nil {
                        glog.Errorf("Encountered an error while deleting the l3vpn rt document: %+v", err)
                        return
                } else {
                        glog.Infof("Successfully deleted l3vpn rt document for RT: %q with Prefix: %q\n", l3vpnRtDocument.RT, l3vpnRtDocument.Prefix)
                }
        }
}
