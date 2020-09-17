package arangodb

import (
	"context"
	"strconv"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

const (
	unicastCollectionName = "UnicastPrefix_Test"
)

func (a *arangoDB) unicastPrefixHandler(obj *message.UnicastPrefix) {
	ctx := context.TODO()
	if obj == nil {
		glog.Warning("UnicastPrefix object is nil")
		return
	}
	k := obj.Prefix + "_" + strconv.Itoa(int(obj.PrefixLen)) + "_" + obj.PeerIP
	// Locking the key "k" to prevent race over the same key value
	a.lckr.Lock(k)
	defer a.lckr.Unlock(k)
	r := &message.UnicastPrefix{
		Key:            k,
		ID:             unicastCollectionName + "/" + k,
		Sequence:       obj.Sequence,
		Hash:           obj.Hash,
		RouterHash:     obj.RouterHash,
		RouterIP:       obj.RouterIP,
		BaseAttributes: obj.BaseAttributes,
		PeerHash:       obj.PeerHash,
		PeerIP:         obj.PeerIP,
		PeerASN:        obj.PeerASN,
		Timestamp:      obj.Timestamp,
		Prefix:         obj.Prefix,
		PrefixLen:      obj.PrefixLen,
		IsIPv4:         obj.IsIPv4,
		OriginAS:       obj.OriginAS,
		Nexthop:        obj.Nexthop,
		IsNexthopIPv4:  obj.IsNexthopIPv4,
		PathID:         obj.PathID,
		Labels:         obj.Labels,
		IsPrepolicy:    obj.IsPrepolicy,
		IsAdjRIBIn:     obj.IsAdjRIBIn,
		PrefixSID:      obj.PrefixSID,
	}

	var prc driver.Collection
	var err error
	if prc, err = a.ensureCollection(unicastCollectionName); err != nil {
		glog.Errorf("failed to ensure for collection %s with error: %+v", unicastCollectionName, err)
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
