package arangodb

import (
	"context"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

const (
	lsNodeCollectionName = "LSNode_Test"
)

func (a *arangoDB) lsnodeHandler(obj *message.LSNode) {
	ctx := context.TODO()
	if obj == nil {
		glog.Warning("LSNode object is nil")
		return
	}
	k := obj.RouterIP + "_" + obj.PeerIP
	// Locking the key "k" to prevent race over the same key value
	a.lckr.Lock(k)
	defer a.lckr.Unlock(k)
	r := &message.LSNode{
		Key:                 k,
		ID:                  lsNodeCollectionName + "/" + k,
		Sequence:            obj.Sequence,
		Hash:                obj.Hash,
		RouterHash:          obj.RouterHash,
		RouterIP:            obj.RouterIP,
		PeerHash:            obj.PeerHash,
		PeerIP:              obj.PeerIP,
		PeerASN:             obj.PeerASN,
		Timestamp:           obj.Timestamp,
		IGPRouterID:         obj.IGPRouterID,
		RouterID:            obj.RouterID,
		ASN:                 obj.ASN,
		LSID:                obj.LSID,
		MTID:                obj.MTID,
		OSPFAreaID:          obj.OSPFAreaID,
		ISISAreaID:          obj.ISISAreaID,
		Protocol:            obj.Protocol,
		ProtocolID:          obj.ProtocolID,
		NodeFlags:           obj.NodeFlags,
		Name:                obj.Name,
		SRCapabilities:      obj.SRCapabilities,
		SRAlgorithm:         obj.SRAlgorithm,
		SRLocalBlock:        obj.SRLocalBlock,
		SRv6CapabilitiesTLV: obj.SRv6CapabilitiesTLV,
		NodeMSD:             obj.NodeMSD,
		IsPrepolicy:         obj.IsPrepolicy,
		IsAdjRIBIn:          obj.IsAdjRIBIn,
		FlexAlgoDefinition:  obj.FlexAlgoDefinition,
	}

	var prc driver.Collection
	var err error
	if prc, err = a.ensureCollection(lsNodeCollectionName); err != nil {
		glog.Errorf("failed to ensure for collection %s with error: %+v", lsNodeCollectionName, err)
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
			glog.V(6).Infof("Update for existing node: %s", k)
			if _, err := prc.UpdateDocument(ctx, k, r); err != nil {
				glog.Errorf("failed to update document %s with error: %+v", k, err)
				return
			}
			// All good, the document was updated and processRouteTargets succeeded, returning...
			return
		}
		glog.V(6).Infof("Add new node: %s", k)
		if _, err := prc.CreateDocument(ctx, r); err != nil {
			glog.Errorf("failed to create document %s with error: %+v", k, err)
			return
		}
	case "del":
		if ok {
			glog.V(6).Infof("Delete for existing node: %s", k)
			// Document by the key exists, hence delete it
			if _, err := prc.RemoveDocument(ctx, k); err != nil {
				glog.Errorf("failed to delete document %s with error: %+v", k, err)
				return
			}
			return
		}
	}
}
