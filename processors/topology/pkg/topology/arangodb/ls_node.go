package arangodb

import (
	"encoding/binary"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/sr"
	//        "github.com/sbezverk/gobmp/pkg/topology/database"
	"github.com/jalapeno-sdn/jalapeno/pkg/topology/database"
)

func (a *arangoDB) lsNodeHandler(obj *message.LSNode) {
	db := a.GetArangoDBInterface()
	action := obj.Action

	srCapabilities := obj.SRCapabilities
	var srCapabilityFlags uint8
	var srgbStart int
	var srgbRange uint32

	if srCapabilities != nil {
		srCapabilityFlags = srCapabilities.Flags
		srCapabilitiesTLVSlice := srCapabilities.TLV
		if srCapabilitiesTLVSlice != nil && len(srCapabilitiesTLVSlice) > 0 {
			srCapabilitiesTLV := srCapabilitiesTLVSlice[0]
			if (sr.CapabilityTLV{} != srCapabilitiesTLV) {
				srgbRange = srCapabilitiesTLV.Range
				if srCapabilitiesTLV.SID != nil {
					srPrefixSIDValue := srCapabilitiesTLV.SID.Value
					srgbStart = parseSRStart(srPrefixSIDValue)
				}
			}
		}
	}

	lsNodeDocument := &database.LSNode{
		Name:              obj.Name,
		IGPID:             obj.IGPRouterID,
		RouterID:          obj.RouterID,
		ASN:               obj.PeerASN,
		SRGBStart:         srgbStart,
		SRGBRange:         srgbRange,
		SRCapabilityFlags: srCapabilityFlags,
		SRv6Capabilities:  obj.SRv6CapabilitiesTLV,
		SRLocalBlock:      obj.SRLocalBlock,
		SRAlgorithm:       obj.SRAlgorithm,
		NodeMaxSIDDepth:   obj.NodeMSD,
		AreaID:            obj.ISISAreaID,
		Protocol:          obj.Protocol,
	}

	if action == "add" {
		if err := db.Upsert(lsNodeDocument); err != nil {
			glog.Errorf("Encountered an error while upserting the ls node document: %+v", err)
			return
		}
		glog.Infof("Successfully added ls node document with IGPRouterID: %q, SRGBStart: %d, and name: %q\n", lsNodeDocument.IGPID, lsNodeDocument.SRGBStart, lsNodeDocument.Name)
	} else {
		if err := db.Delete(lsNodeDocument); err != nil {
			glog.Errorf("Encountered an error while deleting the ls node document: %+v", err)
			return
		} else {
			glog.Infof("Successfully deleted ls node document with IGPRouterID: %q, SRGBStart: %d, and name: %q\n", lsNodeDocument.IGPID, lsNodeDocument.SRGBStart, lsNodeDocument.Name)
		}
	}
}

func parseSRStart(SID []byte) int {
	var data []byte
	if len(SID) != 4 {
		data = make([]byte, 4)
		copy(data[4-len(SID):], SID)
	} else {
		data = SID
	}
	srStart := binary.BigEndian.Uint32(data)
	return int(srStart)
}
