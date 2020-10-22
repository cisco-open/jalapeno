package arangodb

import (
	"github.com/golang/glog"
	"github.com/jalapeno-sdn/jalapeno/pkg/topology/database"
	"github.com/sbezverk/gobmp/pkg/message"
)

func (a *arangoDB) unicastPrefixHandler(obj *message.UnicastPrefix) {
	db := a.GetArangoDBInterface()
	action := obj.Action

	originAS := obj.OriginAS
	is_internal_asn := checkASNLocation(int32(originAS))
	if int(originAS) == db.ASN || int(obj.PeerASN) == db.ASN || is_internal_asn {
		glog.Infof("%q Not an External Origin or Peer ASN, not parsing unicast-prefix message as EPE-Prefix", obj.OriginAS)
		return
	}

	UnicastPrefixDocument := &database.UnicastPrefix{
		PeerIP:         obj.PeerIP,
		PeerASN:        obj.PeerASN,
		Prefix:         obj.Prefix,
		Length:         obj.PrefixLen,
		Nexthop:        obj.Nexthop,
		BaseAttributes: obj.BaseAttributes,
		OriginASN:      obj.OriginAS,
		IsIPv4:         obj.IsIPv4,
		IsNexthopIPv4:  obj.IsNexthopIPv4,
		Labels:         obj.Labels,
		PrefixSID:      obj.PrefixSID,
		Timestamp:      obj.Timestamp,
	}

	if action == "add" {
		if err := db.Upsert(UnicastPrefixDocument); err != nil {
			glog.Errorf("Encountered an error while upserting the unicast prefix document: %+v", err)
			return
		}
		glog.Infof("Successfully added unicast prefix document with peer IP: %q and prefix: %q\n", obj.PeerIP, obj.Prefix)
	} else {
		if err := db.Delete(UnicastPrefixDocument); err != nil {
			glog.Errorf("Encountered an error while deleting the epe prefix document: %+v", err)
			return
		} else {
			glog.Infof("Successfully deleted unicast prefix document with peer IP: %q and prefix %q\n", obj.PeerIP, obj.Prefix)
		}
	}

	/* 	epePeerDocument := &database.EPEPeer{
	   		PeerIP:        obj.PeerIP,
	   		PeerASN:       obj.PeerASN,
	   		Nexthop:       obj.Nexthop,
	   		RouterIP:      obj.RouterIP,
	   		IsNexthopIPv4: obj.IsNexthopIPv4,
	   		Labels:        obj.Labels,
	   		Timestamp:     obj.Timestamp,
	   	}

	   	if action == "add" {
	   		if err := db.Upsert(epePeerDocument); err != nil {
	   			glog.Errorf("Encountered an error while upserting the epe peer document: %+v", err)
	   			return
	   		}
	   		glog.Infof("Successfully added epe peer document with peer IP: %q and egress IP: %q\n", obj.PeerIP, obj.RouterIP)
	   	} else {
	   		if err := db.Delete(UnicastPrefixDocument); err != nil {
	   			glog.Errorf("Encountered an error while deleting the epe prefix document: %+v", err)
	   			return
	   		} else {
	   			glog.Infof("Successfully deleted epe peer document with peer IP: %q and egress IP %q\n", obj.PeerIP, obj.RouterIP)
	   		}
	   	} */
}
