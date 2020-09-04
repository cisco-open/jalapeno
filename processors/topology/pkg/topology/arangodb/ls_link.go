package arangodb

import (
	"encoding/binary"
	"github.com/golang/glog"
	"github.com/jalapeno-sdn/jalapeno/pkg/topology/database"
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

func (a *arangoDB) lsLinkHandler(obj *message.LSLink) {
	db := a.GetArangoDBInterface()
	action := obj.Action

	var SRv6EndXSID *srv6.EndXSIDTLV
	if obj.SRv6ENDXSID != nil {
		SRv6EndXSID = obj.SRv6ENDXSID
	}

	localRouterKey := "LSNode/" + obj.IGPRouterID
	remoteRouterKey := "LSNode/" + obj.RemoteIGPRouterID
	adjacencySIDS := parseAdjacencySIDS(obj.LSAdjacencySID)

	lsLinkDocument := &database.LSLink{
		LocalRouterKey:    		localRouterKey,
		RemoteRouterKey:   		remoteRouterKey,
		Timestamp:         		obj.Timestamp,
		LocalRouterID:     		obj.RouterID,
		RemoteRouterID:    		obj.RemoteRouterID,
		LocalInterfaceIP:  		obj.InterfaceIP,
		RemoteInterfaceIP: 		obj.NeighborIP,
		LocalLinkID:			obj.LocalLinkID,
		RemoteLinkID:			obj.RemoteLinkID,
		LocalIGPID:        		obj.IGPRouterID,
		RemoteIGPID:       		obj.RemoteIGPRouterID,
		Protocol:          		obj.Protocol,
		ProtocolID:        		obj.ProtocolID,
		MTID:              		obj.MTID,
		IGPMetric:         		obj.IGPMetric,
		TEMetric:          		obj.TEDefaultMetric,
		AdminGroup:        		obj.AdminGroup,
		MaxLinkBW:         		obj.MaxLinkBW,
		MaxResvBW:         		obj.MaxResvBW,
		UnResvBW:          		obj.UnResvBW,
		LinkProtection:    		obj.LinkProtection,
		MPLSProtoMask:     		obj.MPLSProtoMask,
		SRLG:              		obj.SRLG,
		LinkName:          		obj.LinkName,
		LocalNodeASN:      		obj.LocalNodeASN,
        RemoteNodeASN:     		obj.RemoteNodeASN,
		AdjacencySID:      		adjacencySIDS,
		SRv6EndXSID:       		SRv6EndXSID,
		LinkMSD:           		obj.LinkMSD,
		UnidirLinkDelay:    	obj.UnidirLinkDelay,
		UnidirLinkDelayMinMax:	obj.UnidirLinkDelayMinMax,
		UnidirDelayVariation:	obj.UnidirDelayVariation,
		UnidirPacketLoss:		obj.UnidirPacketLoss,
		UnidirResidualBW:		obj.UnidirResidualBW,
		UnidirAvailableBW:		obj.UnidirAvailableBW,
		UnidirBWUtilization:	obj.UnidirBWUtilization,
	}

	if action == "add" {
		if err := db.Upsert(lsLinkDocument); err != nil {
			glog.Errorf("Encountered an error while upserting the ls link document with local IP: %q %+v", lsLinkDocument.LocalInterfaceIP, err)
			return
		}
		glog.Infof("Successfully added ls link document from Router: %q through Interface: %q "+
			"to Router: %q through Interface: %q\n", lsLinkDocument.LocalIGPID, lsLinkDocument.LocalInterfaceIP, lsLinkDocument.RemoteIGPID, lsLinkDocument.RemoteInterfaceIP)
	} else {
		if err := db.Delete(lsLinkDocument); err != nil {
			glog.Errorf("Encountered an error while deleting the ls link document: %+v", err)
			return
		} else {
			glog.Infof("Successfully deleted ls link document from Router: %q through Interface: %q "+
				"to Router: %q through Interface: %q\n", lsLinkDocument.LocalIGPID, lsLinkDocument.LocalInterfaceIP, lsLinkDocument.RemoteIGPID, lsLinkDocument.RemoteInterfaceIP)
		}
	}
}

func parseAdjacencySIDS(adjacencySIDList []*sr.AdjacencySIDTLV) []map[string]int {
	var adjacencySIDS []map[string]int
	for _, value := range adjacencySIDList {
		var data []byte
		if len(value.SID) != 4 {
			data = make([]byte, 4)
			copy(data[4-len(value.SID):], value.SID)
		} else {
			data = value.SID
		}
		adjacencySID := binary.BigEndian.Uint32(data)
		adj_dict := map[string]int{"flags": int(value.Flags), "weight": int(value.Weight), "sid": int(adjacencySID)}
		adjacencySIDS = append(adjacencySIDS, adj_dict)
	}
	return adjacencySIDS
}
