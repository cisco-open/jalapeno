package arangodb

import (
	"github.com/golang/glog"
	"github.com/jalapeno-sdn/jalapeno/pkg/topology/database"
	"github.com/sbezverk/gobmp/pkg/message"
	//        "github.com/sbezverk/gobmp/pkg/topology/database"
	//"github.com/cisco-ie/jalapeno/processors/topology/pkg/database"
)

func (a *arangoDB) lsSRv6SIDHandler(obj *message.LSSRv6SID) {
	db := a.GetArangoDBInterface()
	action := obj.Action

	//endpointBehavior := obj.SRv6EndpointBehavior
	//var srv6EndpointBehavior uint16
	//var srv6Flag uint8
	//var srv6Algorithm uint8
	//if endpointBehavior != nil {
	//	srv6EndpointBehavior = endpointBehavior.EndpointBehavior
	//	srv6Flag = endpointBehavior.Flag
	//	srv6Algorithm = endpointBehavior.Algorithm
	//}

	//srv6BGPPeerNodeSID := obj.SRv6BGPPeerNodeSID
	//var srv6BGPPeerNodeSIDFlag uint8
	//var srv6BGPPeerNodeSIDWeight uint8
	//var srv6BGPPeerNodeSIDPeerASN uint32
	//var srv6BGPPeerNodeSIDID []byte
	//if srv6BGPPeerNodeSID != nil {
	//	srv6BGPPeerNodeSIDFlag = srv6BGPPeerNodeSID.Flag
	//	srv6BGPPeerNodeSIDWeight = srv6BGPPeerNodeSID.Weight
	//	srv6BGPPeerNodeSIDPeerASN = srv6BGPPeerNodeSID.PeerASN
	//	srv6BGPPeerNodeSIDID = srv6BGPPeerNodeSID.PeerID
	//}

	//srv6SIDStructure := obj.SRv6SIDStructure
	//var srv6SIDStructureLBLength uint8
	//var srv6SIDStructureLNLength uint8
	//var srv6SIDStructureFunLength uint8
	//var srv6SIDStructureArgLength uint8
	//if srv6BGPPeerNodeSID != nil {
	//	srv6SIDStructureLBLength = srv6SIDStructure.LBLength
	//	srv6SIDStructureLNLength = srv6SIDStructure.LNLength
	//	srv6SIDStructureFunLength = srv6SIDStructure.FunLength
	//	srv6SIDStructureArgLength = srv6SIDStructure.ArgLength
	//}

	lsSRv6SIDDocument := &database.LSSRv6SID{
		RouterIP:                  obj.RouterIP,
		PeerIP:                    obj.PeerIP,
		PeerASN:                   obj.PeerASN,
		Timestamp:                 obj.Timestamp,
		IGPRouterID:               obj.IGPRouterID,
		LocalNodeASN:              obj.LocalNodeASN,
		RouterID:                  obj.RouterID,
		OSPFAreaID:                obj.OSPFAreaID,
		ISISAreaID:                obj.ISISAreaID,
		Protocol:                  obj.Protocol,
		Nexthop:                   obj.Nexthop,
		IGPFlags:                  obj.IGPFlags,
		MTID:                      obj.MTID,
		OSPFRouteType:             obj.OSPFRouteType,
		IGPRouteTag:               obj.IGPRouteTag,
		IGPExtRouteTag:            obj.IGPExtRouteTag,
		OSPFFwdAddr:               obj.OSPFFwdAddr,
		IGPMetric:                 obj.IGPMetric,
		Prefix:                    obj.Prefix,
		PrefixLen:                 obj.PrefixLen,
		SRv6SID:                   obj.SRv6SID,
		SRv6EndpointBehavior:      obj.SRv6EndpointBehavior,
		SRv6BGPPeerNodeSID:        obj.SRv6BGPPeerNodeSID,
		SRv6SIDStructure:          obj.SRv6SIDStructure,

		//SRv6EndpointBehaviorRaw:   obj.SRv6EndpointBehavior,
		//SRv6BGPPeerNodeSIDRaw:     obj.SRv6BGPPeerNodeSID,
		//SRv6SIDStructureRaw:       obj.SRv6SIDStructure,
		//SRv6EndpointBehavior:      srv6EndpointBehavior,
		//SRv6Flag:                  srv6Flag,
		//SRv6Algorithm:             srv6Algorithm,
		//SRv6BGPPeerNodeSIDFlag:    srv6BGPPeerNodeSIDFlag,
		//SRv6BGPPeerNodeSIDWeight:  srv6BGPPeerNodeSIDWeight,
		//SRv6BGPPeerNodeSIDPeerASN: srv6BGPPeerNodeSIDPeerASN,
		//SRv6BGPPeerNodeSIDID:      srv6BGPPeerNodeSIDID,
		//SRv6SIDStructureLBLength:  srv6SIDStructureLBLength,
		//SRv6SIDStructureLNLength:  srv6SIDStructureLNLength,
		//SRv6SIDStructureFunLength: srv6SIDStructureFunLength,
		//SRv6SIDStructureArgLength: srv6SIDStructureArgLength,
	}
	if action == "add" {
		if err := db.Upsert(lsSRv6SIDDocument); err != nil {
			glog.Errorf("Encountered an error while upserting the LS_SRv6_SID Document: %+v", err)
			return
		}
		glog.Infof("Successfully added LS_SRv6_SID Document with SID: %q and IGPRouterID: %q\n", lsSRv6SIDDocument.SRv6SID, lsSRv6SIDDocument.IGPRouterID)
	} else {
		if err := db.Delete(lsSRv6SIDDocument); err != nil {
			glog.Errorf("Encountered an error while deleting the LS_SRv6_SID Document: %+v", err)
			return
		} else {
			glog.Infof("Successfully deleted LS_SRv6_SID Document with IGPRouterID: %q\n", lsSRv6SIDDocument.IGPRouterID)
		}
	}
}
