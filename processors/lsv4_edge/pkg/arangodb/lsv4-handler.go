package arangodb

import (
	"context"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	notifier "github.com/jalapeno-sdn/topology/pkg/kafkanotifier"
	"github.com/sbezverk/gobmp/pkg/bgp"
	"github.com/sbezverk/gobmp/pkg/message"
)

// LSV4 defines lsv4 edge record
type LSV4 struct {
	ID   string `json:"_id,omitempty"`
	Key  string `json:"_key,omitempty"`
	Rev  string `json:"_rev,omitempty"`
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
	//RT       string            `json:"RT,omitempty"`
	//Prefixes map[string]string `json:"Prefixes,omitempty"`
}

func (a *arangoDB) lsv4Handler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	// Check if Collection encoded in ID exists
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.lsv4.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.lsv4.Name(), c)
	}
	glog.V(5).Infof("Processing action: %s for ls_link: %s", obj.Action, obj.Key)
	var o message.LSLink
	_, err := a.lsv4.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
	}
	switch obj.Action {
	case "add":
		if err := a.processAddLSv4Edge(ctx, obj.Key, obj.ID, o.IGPRouterID, o.RemoteIGPRouterID); err != nil {
			return fmt.Errorf("failed to update the edge collection %s with reference to %s with error: %+v", a.lsv4.Name(), obj.Key, err)
		}
	case "update":
		if err := a.processUpdateLSv4Edge(ctx, obj.Key, obj.ID, o.IGPRouterID, o.RemoteIGPRouterID); err != nil {
			return fmt.Errorf("failed to update the edge collection %s with reference to %s with error: %+v", a.lsv4.Name(), obj.Key, err)
		}
	case "del":
		if err := a.processDeleteLSv4Edge(ctx, obj.Key, obj.ID); err != nil {
			return fmt.Errorf("failed to clean up the edge collection %s from references to %s with error: %+v", a.lsv4.Name(), obj.Key, err)
		}
	}
	return nil
}

//func (a *arangoDB) processAddRouteTargets(ctx context.Context, key, id string, extCommunities []string) error {
func (a *arangoDB) processAddLSv4Edge(ctx context.Context, key, id string, from string, to string) error {
	v := strings.TrimPrefix(ext, bgp.ECPRouteTarget)
	if err := a.processRTAdd(ctx, id, key, v); err != nil {
		return err
	}

	return nil
}

//func (a *arangoDB) processDeleteRouteTargets(ctx context.Context, key, id string) error {
//	rts, ok := a.store[key]
//	if !ok {
//		return nil
//	}
//	for _, rt := range rts {
//		if err := a.processRTDel(ctx, id, key, rt); err != nil {
//			return err
//		}
//	}
//	delete(a.store, key)

//	return nil
//}

//func (a *arangoDB) processUpdateRouteTargets(ctx context.Context, key, id string, extCommunities []string) error {
//	erts, ok := a.store[key]
//	if !ok {
//		return fmt.Errorf("attempting to update non existing in the store prefix: %s", key)
//	}
//	nrts := make([]string, len(extCommunities))
//	for i, rt := range extCommunities {
//		if !strings.HasPrefix(rt, bgp.ECPRouteTarget) {
//			continue
//		}
//		v := strings.TrimPrefix(rt, bgp.ECPRouteTarget)
//		nrts[i] = v
//	}
//	toAdd, toDel := ExtCommGetDiff("", erts, nrts)

//	for _, rt := range toAdd {
//		if err := a.processRTAdd(ctx, id, key, rt); err != nil {
//			return err
//		}
//	}
//	for _, rt := range toDel {
//		if err := a.processRTDel(ctx, id, key, rt); err != nil {
//			return err
//		}
//	}
//	rts, ok := a.store[key]
//	if !ok {
//		return fmt.Errorf("update corrupted existing entry in the store prefix: %s", key)
//	}
//	if len(rts) == 0 {
//		delete(a.store, key)
//	}

//	return nil
//}
func (a *arangoDB) processLSV4Add(ctx context.Context, id, key, v string) error {
	found, err := a.lsv4.DocumentExists(ctx, v)
	if err != nil {
		return err
	}
	rtr := &LSv4Edge{}
	nctx := driver.WithWaitForSync(ctx)
	if found {
		_, err := a.lsv4.ReadDocument(nctx, v, rtr)
		if err != nil {
			return err
		}
		if _, ok := rtr.Prefixes[id]; ok {
			return nil
		}
		rtr.Prefixes[id] = key
		_, err = a.lsv4.UpdateDocument(nctx, v, rtr)
		if err != nil {
			return err
		}
	} else {
		rtr.ID = a.lsv4.Name() + "/" + v
		rtr.Key = v
		rtr.From = v
		rtr.Prefixes = map[string]string{
			id: key,
		}
		if _, err := a.rt.CreateDocument(nctx, rtr); err != nil {
			return err
		}
	}
	// Updating store with new  prefix - rt entry
	rts, ok := a.store[key]
	if !ok {
		rts = make([]string, 1)
	}
	rts = append(rts, v)
	a.store[key] = rts

	return nil
}

func (a *arangoDB) processRTDel(ctx context.Context, id, key, v string) error {
	found, err := a.rt.DocumentExists(ctx, v)
	if err != nil {
		return err
	}
	rtr := &L3VPNRT{}
	nctx := driver.WithWaitForSync(ctx)
	if !found {
		return nil
	}
	if _, err := a.rt.ReadDocument(nctx, v, rtr); err != nil {
		return err
	}
	if _, ok := rtr.Prefixes[id]; !ok {
		return nil
	}
	delete(rtr.Prefixes, id)
	// Check If route target document has any references to other prefixes, if no, then deleting
	// Route Target document, otherwise updating it
	if len(rtr.Prefixes) == 0 {
		glog.Infof("RT with key %s has no more entries, deleting it...", v)
		_, err := a.rt.RemoveDocument(ctx, v)
		if err != nil {
			return fmt.Errorf("failed to delete empty route target %s with error: %+v", v, err)
		}
		return nil
	}
	_, err = a.rt.UpdateDocument(nctx, v, rtr)
	if err != nil {
		return err
	}
	// Updating store with new lsv4_edge entry
	rts, ok := a.store[key]
	if ok {
		for i, rt := range rts {
			if strings.Compare(rt, v) == 0 {
				rts = append(rts[:i], rts[i+1:]...)
				break
			}
		}
	}
	a.store[key] = rts

	return nil
}
