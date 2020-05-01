package arangodb

import (
        "github.com/golang/glog"
        "github.com/sbezverk/gobmp/pkg/message"
        "github.com/cisco-ie/jalapeno/processors/topology/pkg/database"
)

func (a *arangoDB) lsPrefixHandler(obj *message.LSPrefix) {
        db := a.GetArangoDBInterface()
        action := obj.Action
        nodePrefixSID := obj.LSPrefixSID
        nodePrefixSIDIndex := nodePrefixSID.SID
        routerDocument := &database.Router{
                BGPID:        obj.Prefix,
                RouterIP:     obj.Prefix,
                NodeSIDIndex: nodePrefixSIDIndex,
        }

        if (action == "add") {
                if err := db.Upsert(routerDocument); err != nil {
                        glog.Errorf("Encountered an error while upserting the ls prefix router document: %+v", err)
                        return
                }
                glog.Infof("Successfully added ls prefix router document with router IP: %q and node-SID index: %q\n", routerDocument.RouterIP, routerDocument.NodeSIDIndex)
        } else {
                if err := db.Delete(routerDocument); err != nil {
                        glog.Errorf("Encountered an error while deleting the ls prefix router document: %+v", err)
                        return
                } else {
                        glog.Infof("Successfully deleted ls prefix router document with router IP: %q and node-SID index: %q\n", routerDocument.RouterIP, routerDocument.NodeSIDIndex)
                }
        }
}

