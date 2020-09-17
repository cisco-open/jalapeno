package arangodb

import (
	"context"
	"strconv"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

const (
	lsPrefixCollectionName = "LSPrefix_Test"
)

func (a *arangoDB) lsprefixHandler(obj *message.LSPrefix) {
	ctx := context.TODO()
	if obj == nil {
		glog.Warning("LSPrefix object is nil")
		return
	}
	k := obj.Prefix + "_" + strconv.Itoa(int(obj.PrefixLen)) + "_" + obj.IGPRouterID
	// Locking the key "k" to prevent race over the same key value
	a.lckr.Lock(k)
	defer a.lckr.Unlock(k)
	r := &message.LSPrefix{
		Key:                  k,
		ID:                   lsPrefixCollectionName + "/" + k,
		RouterIP:             obj.RouterIP,
		PeerIP:               obj.PeerIP,
		PeerASN:              obj.PeerASN,
		Timestamp:            obj.Timestamp,
		IGPRouterID:          obj.IGPRouterID,
		RouterID:             obj.RouterID,
		LSID:                 obj.LSID,
		ProtocolID:           obj.ProtocolID,
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
		LSPrefixSID:          obj.LSPrefixSID,
		PrefixAttrFlags:      obj.PrefixAttrFlags,
		FlexAlgoPrefixMetric: obj.FlexAlgoPrefixMetric,
		SRv6Locator:          obj.SRv6Locator,
	}

	var prc driver.Collection
	var err error
	if prc, err = a.ensureCollection(lsPrefixCollectionName); err != nil {
		glog.Errorf("failed to ensure for collection %s with error: %+v", lsPrefixCollectionName, err)
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
			glog.V(6).Infof("Update for existing prefix: %s", k)
			if _, err := prc.UpdateDocument(ctx, k, r); err != nil {
				glog.Errorf("failed to update document %s with error: %+v", k, err)
				return
			}
			// All good, the document was updated and processRouteTargets succeeded, returning...
			return
		}
		glog.V(6).Infof("Add new prefix: %s", k)
		if _, err := prc.CreateDocument(ctx, r); err != nil {
			glog.Errorf("failed to create document %s with error: %+v", k, err)
			return
		}
	case "del":
		if ok {
			glog.V(6).Infof("Delete for existing prefix: %s", k)
			// Document by the key exists, hence delete it
			if _, err := prc.RemoveDocument(ctx, k); err != nil {
				glog.Errorf("failed to delete document %s with error: %+v", k, err)
				return
			}
			return
		}
	}
}
