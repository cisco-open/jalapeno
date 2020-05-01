package arangodb

import (
        "encoding/binary"
        "github.com/golang/glog"
        "github.com/sbezverk/gobmp/pkg/message"
        "github.com/cisco-ie/jalapeno/processors/topology/pkg/database"
)

func (a *arangoDB) lsNodeHandler(obj *message.LSNode) {
        db := a.GetArangoDBInterface()
        action := obj.Action

        //var tmp *sr.Capability
        //tmp = obj.SRCapabilities
        srCapabilities := obj.SRCapabilities.TLV[0]
        srgbRange := srCapabilities.Range
        srgbStart := parseSRStart(srCapabilities.SID.Value)
        srPrefixSID := getSRPrefixSID(srgbStart, obj.RouterID, db)

        lsNodeDocument := &database.LSNode{
                Name:      obj.Name,
                RouterID:  obj.RouterID,
                ASN:       obj.PeerASN,
                PrefixSID: srPrefixSID,
                SRGBStart: srgbStart,
                SRGBRange: srgbRange,
                SRGB:      obj.SRCapabilities,
                IGPID:     obj.IGPRouterID,
                SRv6Capabilities: obj.SRv6CapabilitiesTLV,
                SRLocalBlock: obj.SRLocalBlock,
                SRAlgorithm: obj.SRAlgorithm,
                NodeMaxSIDDepth: obj.NodeMSD,
                AreaID:    obj.ISISAreaID,
                Protocol:  obj.Protocol,
        }
        if (action == "add") {
                if err := db.Upsert(lsNodeDocument); err != nil {
                        glog.Errorf("Encountered an error while upserting the ls node document: %+v", err)
                        return
                }
                glog.Infof("Successfully added ls node document with router ID: %q, PrefixSID: %q, and name: %q\n", lsNodeDocument.RouterID, lsNodeDocument.PrefixSID, lsNodeDocument.Name)
        } else {
                if err := db.Delete(lsNodeDocument); err != nil {
                        glog.Errorf("Encountered an error while deleting the ls node document: %+v", err)
                        return
                } else {
                        glog.Infof("Successfully deleted ls node document with router ID: %q, PrefixSID: %q, and name: %q\n", lsNodeDocument.RouterID, lsNodeDocument.PrefixSID, lsNodeDocument.Name)
                }
        }
}


func parseSRStart(SID []byte) int {
        var data []byte
        if(len(SID) != 4) {
                data = make([]byte,4)
                copy(data[4-len(SID):], SID)
        } else {
                data = SID
        }
        srStart := binary.BigEndian.Uint32(data)
        return int(srStart)
}

func getSRPrefixSID(srgbStart int, routerID string, db *database.ArangoConn) int {
        var prefixSID int
        sidIndex := db.GetSIDIndex(routerID)
        if(sidIndex != "") {
                prefixSID = calculateSID(srgbStart, sidIndex)
        }
        return prefixSID
}

