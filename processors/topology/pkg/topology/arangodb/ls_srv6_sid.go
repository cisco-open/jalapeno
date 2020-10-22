package arangodb

import (
	"github.com/golang/glog"
	"github.com/jalapeno-sdn/jalapeno/pkg/topology/database"
	"github.com/sbezverk/gobmp/pkg/message"
)

func (a *arangoDB) lsSRv6SIDHandler(obj *message.LSSRv6SID) {
	db := a.GetArangoDBInterface()
	action := obj.Action

	lsSRv6SIDDocument := &database.LSSRv6SID{
		RouterIP:             obj.RouterIP,
		PeerIP:               obj.PeerIP,
		PeerASN:              obj.PeerASN,
		Timestamp:            obj.Timestamp,
		IGPRouterID:          obj.IGPRouterID,
		LocalNodeASN:         obj.LocalNodeASN,
		RouterID:             obj.RouterID,
		OSPFAreaID:           obj.OSPFAreaID,
		ISISAreaID:           obj.ISISAreaID,
		Protocol:             obj.Protocol,
		Nexthop:              obj.Nexthop,
		IGPFlags:             obj.IGPFlags,
		MTID:                 obj.MTID,
		OSPFFwdAddr:          obj.OSPFFwdAddr,
		IGPRouteTag:          obj.IGPRouteTag,
		IGPExtRouteTag:       obj.IGPExtRouteTag,
		IGPMetric:            obj.IGPMetric,
		Prefix:               obj.Prefix,
		PrefixLen:            obj.PrefixLen,
		SRv6SID:              obj.SRv6SID,
		SRv6EndpointBehavior: obj.SRv6EndpointBehavior,
		SRv6BGPPeerNodeSID:   obj.SRv6BGPPeerNodeSID,
		SRv6SIDStructure:     obj.SRv6SIDStructure,
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
