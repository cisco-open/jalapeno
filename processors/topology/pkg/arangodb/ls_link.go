package arangodb

import (
        "github.com/golang/glog"
        "github.com/sbezverk/gobmp/pkg/message"
        "github.com/cisco-ie/jalapeno/processors/topology/pkg/database"
)

func (a *arangoDB) lsLinkHandler(obj *message.LSLink) {
        db := a.GetArangoDBInterface()
        action := obj.Action

        localRouterKey := "LSNode/" + obj.RouterID
        remoteRouterKey := "LSNode/" + obj.RemoteRouterID

        lsLinkDocument := &database.LSLink{
                LocalRouterKey:    localRouterKey,
                RemoteRouterKey:   remoteRouterKey,
                LocalRouterID:     obj.RouterID,
                RemoteRouterID:    obj.RemoteRouterID,
                ASN:               obj.LocalNodeASN,
                LocalInterfaceIP:  obj.InterfaceIP,
                RemoteInterfaceIP: obj.NeighborIP,
                Protocol:          obj.Protocol,
                IGPID:             obj.IGPRouterID,
                IGPMetric:         obj.IGPMetric,
                TEMetric:          obj.TEDefaultMetric,
                AdminGroup:        obj.AdminGroup,
                MaxLinkBW:         obj.MaxLinkBW,
                MaxResvBW:         obj.MaxResvBW,
                UnResvBW:          obj.UnResvBW,
                LinkProtection:    obj.LinkProtection,
                SRLG:              obj.SRLG,
                LinkName:          obj.LinkName,
                AdjacencySID:      obj.LSAdjacencySID,
                Timestamp:         obj.Timestamp,
                LinkMaxSIDDepth:   obj.LinkMSD,
        }

        //if (lsLinkDocument.AdjacencySID.String() == "") {
        //        glog.Infof("No NodeSID data available, not parsing ls link document")
        //        return
        //}

        if (action == "add") {
                if err := db.Upsert(lsLinkDocument); err != nil {
                        glog.Errorf("Encountered an error while upserting the ls link document: %+v", err)
                        return
                }
                glog.Infof("Successfully added ls link document from Router: %q through Interface: %q " +
                            "to Router: %q through Interface: %q\n", lsLinkDocument.LocalRouterKey, lsLinkDocument.LocalInterfaceIP, lsLinkDocument.RemoteRouterKey, lsLinkDocument.RemoteInterfaceIP)
                /*
                aSids := strings.Split(lsLinkDocument.AdjacencySID, ", ")
                key := lsLinkDocument.LocalRouterID + "_" + lsLinkDocument.LocalInterfaceIP + "_" + lsLinkDocument.RemoteInterfaceIP + "_" + lsLinkDocument.RemoteRouterID
                for _, aSid :=  range aSids {
                        ,"ls_adjacency_sid": {"sr_adj_sid_flags":112,"sr_adj_sid":"AF3B"}
                        46: (ls_adjacency_sid): BVL 0 24002, VL 0 24003
                        "ls_adjacency_sid":[{"flags":112,"weight":0,"sid":"AF3C"},{"flags":48,"weight":0,"sid":"AF3D"}]
                        s := strings.Split(aSid, " ")
                        adj_sid := s[2]
                        flags := s[0]
                        weight := s[1]
                        a.db.CreateAdjacencyList(key, adj_sid, flags, weight)
                }
                */
        } else {
                if err := db.Delete(lsLinkDocument); err != nil {
                        glog.Errorf("Encountered an error while deleting the ls link document: %+v", err)
                        return
                } else {
                        glog.Infof("Successfully deleted ls link document from Router: %q through Interface: %q " +
                            "to Router: %q through Interface: %q\n", lsLinkDocument.LocalRouterKey, lsLinkDocument.LocalInterfaceIP, lsLinkDocument.RemoteRouterKey, lsLinkDocument.RemoteInterfaceIP)
                }
        }
}

