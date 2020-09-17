package arangodb

import (
	"context"
	"strconv"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

const (
	lsLinkCollectionName = "LSLink_Test"
)

func (a *arangoDB) lslinkHandler(obj *message.LSLink) {
	ctx := context.TODO()
	if obj == nil {
		glog.Warning("LSLink object is nil")
		return
	}
	var localIP, remoteIP, localID, remoteID string
	switch obj.MTID {
	case 0:
		localIP = "0.0.0.0"
		remoteIP = "0.0.0.0"
	case 2:
		localIP = "::"
		remoteIP = "::"
	default:
		localIP = "unknown-mt-id"
		remoteIP = "unknown-mt-id"
	}
	if len(obj.LocalLinkIP) != 0 {
		localIP = obj.LocalLinkIP[0]
	}
	if len(obj.RemoteLinkIP) != 0 {
		remoteIP = obj.RemoteLinkIP[0]
	}
	localID = strconv.Itoa(int(obj.LocalLinkID))
	remoteID = strconv.Itoa(int(obj.RemoteLinkID))
	k := obj.IGPRouterID + "_" + localIP + "_" + localID + "_" + obj.RemoteIGPRouterID + "_" + remoteIP + "_" + remoteID
	// Locking the key "k" to prevent race over the same key value
	a.lckr.Lock(k)
	defer a.lckr.Unlock(k)
	r := &message.LSLink{
		Key:                   k,
		ID:                    lsLinkCollectionName + "/" + k,
		RouterIP:              obj.RouterIP,
		PeerHash:              obj.PeerHash,
		PeerIP:                obj.PeerIP,
		PeerASN:               obj.PeerASN,
		Timestamp:             obj.Timestamp,
		IGPRouterID:           obj.IGPRouterID,
		RouterID:              obj.RouterID,
		LSID:                  obj.LSID,
		Protocol:              obj.Protocol,
		Nexthop:               obj.Nexthop,
		MTID:                  obj.MTID,
		LocalLinkID:           obj.LocalLinkID,
		RemoteLinkID:          obj.RemoteLinkID,
		LocalLinkIP:           obj.LocalLinkIP,
		RemoteLinkIP:          obj.RemoteLinkIP,
		IGPMetric:             obj.IGPMetric,
		AdminGroup:            obj.AdminGroup,
		MaxLinkBW:             obj.MaxLinkBW,
		MaxResvBW:             obj.MaxResvBW,
		UnResvBW:              obj.UnResvBW,
		TEDefaultMetric:       obj.TEDefaultMetric,
		LinkProtection:        obj.LinkProtection,
		MPLSProtoMask:         obj.MPLSProtoMask,
		SRLG:                  obj.SRLG,
		LinkName:              obj.LinkName,
		RemoteNodeHash:        obj.RemoteNodeHash,
		LocalNodeHash:         obj.LocalNodeHash,
		RemoteIGPRouterID:     obj.RemoteIGPRouterID,
		RemoteRouterID:        obj.RemoteRouterID,
		LocalNodeASN:          obj.LocalNodeASN,
		RemoteNodeASN:         obj.RemoteNodeASN,
		SRv6BGPPeerNodeSID:    obj.SRv6BGPPeerNodeSID,
		SRv6ENDXSID:           obj.SRv6ENDXSID,
		LSAdjacencySID:        obj.LSAdjacencySID,
		LinkMSD:               obj.LinkMSD,
		AppSpecLinkAttr:       obj.AppSpecLinkAttr,
		UnidirLinkDelay:       obj.UnidirLinkDelay,
		UnidirLinkDelayMinMax: obj.UnidirLinkDelayMinMax,
		UnidirDelayVariation:  obj.UnidirDelayVariation,
		UnidirPacketLoss:      obj.UnidirPacketLoss,
		UnidirResidualBW:      obj.UnidirResidualBW,
		UnidirAvailableBW:     obj.UnidirAvailableBW,
		UnidirBWUtilization:   obj.UnidirBWUtilization,
	}

	var prc driver.Collection
	var err error
	if prc, err = a.ensureCollection(lsLinkCollectionName); err != nil {
		glog.Errorf("failed to ensure for collection %s with error: %+v", lsLinkCollectionName, err)
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
			glog.V(6).Infof("Update for existing link: %s", k)
			if _, err := prc.UpdateDocument(ctx, k, r); err != nil {
				glog.Errorf("failed to update document %s with error: %+v", k, err)
				return
			}
			// All good, the document was updated and processRouteTargets succeeded, returning...
			return
		}
		glog.V(6).Infof("Add new link: %s", k)
		if _, err := prc.CreateDocument(ctx, r); err != nil {
			glog.Errorf("failed to create document %s with error: %+v", k, err)
			return
		}
	case "del":
		if ok {
			glog.V(6).Infof("Delete for existing link: %s", k)
			// Document by the key exists, hence delete it
			if _, err := prc.RemoveDocument(ctx, k); err != nil {
				glog.Errorf("failed to delete document %s with error: %+v", k, err)
				return
			}
			return
		}
	}
}
