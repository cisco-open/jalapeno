package arangodb

import (
        "github.com/golang/glog"
        "github.com/sbezverk/gobmp/pkg/message"
        "github.com/cisco-ie/jalapeno/processors/topology/pkg/database"
)

func (a *arangoDB) unicastPrefixHandler(obj *message.UnicastPrefix) {
        db := a.GetArangoDBInterface()
        action := obj.Action

        epePrefixDocument := &database.EPEPrefix {
                PeerIP:        obj.PeerIP,
                PeerASN:       obj.PeerASN,
                Prefix:        obj.Prefix,
                Length:        obj.PrefixLen,
                Nexthop:       obj.Nexthop,
                ASPath:        obj.BaseAttributes.ASPath,
                OriginASN:     obj.OriginAS,
                ASPathCount:   obj.BaseAttributes.ASPathCount,
                MED:           obj.BaseAttributes.MED,
                LocalPref:     obj.BaseAttributes.LocalPref,
                CommunityList: obj.BaseAttributes.CommunityList,
                ExtComm:       obj.BaseAttributes.ExtCommunityList,
                IsIPv4:        obj.IsIPv4,
                IsNexthopIPv4: obj.IsNexthopIPv4,
                Labels:        obj.Labels,
                Timestamp:     obj.Timestamp,
        }

        if (action == "add") {
                if err := db.Upsert(epePrefixDocument); err != nil {
                        glog.Errorf("Encountered an error while upserting the epe prefix document: %+v", err)
                        return
                }
                glog.Infof("Successfully added epe prefix document with peer IP: %q and prefix: %q\n", obj.PeerIP, obj.Prefix)
        } else {
                if err := db.Delete(epePrefixDocument); err != nil {
                        glog.Errorf("Encountered an error while deleting the epe prefix document: %+v", err)
                        return
                } else {
                        glog.Infof("Successfully deleted epe prefix document with peer IP: %q and prefix %q\n", obj.PeerIP, obj.Prefix)
                }
        }
}


