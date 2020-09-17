package arangodb

import (
	"context"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

const (
	peerChangeCollectionName = "Node_Test"
)

func (a *arangoDB) peerChangeHandler(obj *message.PeerStateChange) {
	ctx := context.TODO()
	if obj == nil {
		glog.Warning("Peer Change object is nil")
		return
	}
	k := obj.RouterIP
	// Locking the key "k" to prevent race over the same key value
	a.lckr.Lock(k)
	defer a.lckr.Unlock(k)
	r := &message.PeerStateChange{
		Key:              k,
		ID:               peerChangeCollectionName + "/" + k,
		Sequence:         obj.Sequence,
		Hash:             obj.Hash,
		RouterHash:       obj.RouterHash,
		Name:             obj.Name,
		RemoteBGPID:      obj.RemoteBGPID,
		RouterIP:         obj.RouterIP,
		Timestamp:        obj.Timestamp,
		RemoteASN:        obj.RemoteASN,
		RemoteIP:         obj.RemoteIP,
		PeerRD:           obj.PeerRD,
		RemotePort:       obj.RemotePort,
		LocalASN:         obj.LocalASN,
		LocalIP:          obj.LocalIP,
		LocalPort:        obj.LocalPort,
		LocalBGPID:       obj.LocalBGPID,
		InfoData:         obj.InfoData,
		AdvCapabilities:  obj.AdvCapabilities,
		RcvCapabilities:  obj.RcvCapabilities,
		RemoteHolddown:   obj.RemoteHolddown,
		AdvHolddown:      obj.AdvHolddown,
		BMPReason:        obj.BMPReason,
		BMPErrorCode:     obj.BMPErrorCode,
		BMPErrorSubCode:  obj.BMPErrorSubCode,
		ErrorText:        obj.ErrorText,
		IsL3VPN:          obj.IsL3VPN,
		IsPrepolicy:      obj.IsPrepolicy,
		IsIPv4:           obj.IsIPv4,
		IsLocRIB:         obj.IsLocRIB,
		IsLocRIBFiltered: obj.IsLocRIBFiltered,
		TableName:        obj.TableName,
	}

	var prc driver.Collection
	var err error
	if prc, err = a.ensureCollection(peerChangeCollectionName); err != nil {
		glog.Errorf("failed to ensure for collection %s with error: %+v", peerChangeCollectionName, err)
		return
	}
	ok, err := prc.DocumentExists(ctx, k)
	if err != nil {
		glog.Errorf("failed to check for document %s with error: %+v", k, err)
		return
	}

	switch obj.Action {
	case "up":
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
	case "down":
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
