package arangodb

import (
	"context"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

const (
	lsSRv6SIDLSCollectionName = "LSSRv6SID_Test"
)

func (a *arangoDB) lsSRV6SIDHandler(obj *message.LSSRv6SID) {
	ctx := context.TODO()
	if obj == nil {
		glog.Warning("srv6LS object is nil")
		return
	}
	k := obj.IGPRouterID
	// Locking the key "k" to prevent race over the same key value
	a.lckr.Lock(k)
	defer a.lckr.Unlock(k)
	r := &message.LSSRv6SID{
		Key:                  k,
		ID:                   lsSRv6SIDLSCollectionName + "/" + k,
		Rev:                  obj.Rev,
		Action:               obj.Action,
		Sequence:             obj.Sequence,
		Hash:                 obj.Hash,
		RouterHash:           obj.RouterHash,
		RouterIP:             obj.RouterIP,
		PeerHash:             obj.PeerHash,
		PeerIP:               obj.PeerIP,
		PeerASN:              obj.PeerASN,
		Timestamp:            obj.Timestamp,
		IGPRouterID:          obj.IGPRouterID,
		LocalNodeASN:         obj.LocalNodeASN,
		RouterID:             obj.RouterID,
		LSID:                 obj.LSID,
		OSPFAreaID:           obj.OSPFAreaID,
		ISISAreaID:           obj.ISISAreaID,
		Protocol:             obj.Protocol,
		Nexthop:              obj.Nexthop,
		LocalNodeHash:        obj.LocalNodeHash,
		MTID:                 obj.MTID,
		OSPFRouteType:        obj.OSPFRouteType,
		IGPFlags:             obj.IGPFlags,
		IGPRouteTag:          obj.IGPRouteTag,
		IGPExtRouteTag:       obj.IGPExtRouteTag,
		OSPFFwdAddr:          obj.OSPFFwdAddr,
		IGPMetric:            obj.IGPMetric,
		Prefix:               obj.Prefix,
		PrefixLen:            obj.PrefixLen,
		IsPrepolicy:          obj.IsPrepolicy,
		IsAdjRIBIn:           obj.IsAdjRIBIn,
		SRv6SID:              obj.SRv6SID,
		SRv6EndpointBehavior: obj.SRv6EndpointBehavior,
		SRv6BGPPeerNodeSID:   obj.SRv6BGPPeerNodeSID,
		SRv6SIDStructure:     obj.SRv6SIDStructure,
	}

	var prc driver.Collection
	var err error
	if prc, err = a.ensureCollection(lsSRv6SIDLSCollectionName); err != nil {
		glog.Errorf("failed to ensure for collection %s with error: %+v", lsSRv6SIDLSCollectionName, err)
		return
	}
	ok, err := prc.DocumentExists(ctx, k)
	if err != nil {
		glog.Errorf("failed to check for document %s with error: %+v", k, err)
		return
	}

	switch obj.Action {
	case "add":
		if ok {
			glog.V(6).Infof("Update for existing srv6 sid: %s", k)
			if _, err := prc.UpdateDocument(ctx, k, r); err != nil {
				glog.Errorf("failed to update document %s with error: %+v", k, err)
				return
			}
			// All good, the document was updated and processRouteTargets succeeded, returning...
			return
		}
		glog.V(6).Infof("Add new srv6 sid: %s", k)
		if _, err := prc.CreateDocument(ctx, r); err != nil {
			glog.Errorf("failed to create document %s with error: %+v", k, err)
			return
		}
	case "del":
		if ok {
			glog.V(6).Infof("Delete for existing srv6 sid: %s", k)
			// Document by the key exists, hence delete it
			if _, err := prc.RemoveDocument(ctx, k); err != nil {
				glog.Errorf("failed to delete document %s with error: %+v", k, err)
				return
			}
			return
		}
	}
}
