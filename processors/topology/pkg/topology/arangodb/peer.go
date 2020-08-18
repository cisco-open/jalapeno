package arangodb

import (
	"github.com/golang/glog"
	"github.com/jalapeno-sdn/jalapeno/pkg/topology/database"
	"github.com/sbezverk/gobmp/pkg/message"
	//        "github.com/sbezverk/gobmp/pkg/topology/database"
	//"github.com/cisco-ie/jalapeno/processors/topology/pkg/database"
)

func (a *arangoDB) peerChangeHandler(obj *message.PeerStateChange) {
	db := a.GetArangoDBInterface()
	action := obj.Action
	nodeID := obj.LocalBGPID
	nodeASN := obj.LocalASN
	peerIP := obj.RemoteIP
	peerASN := obj.RemoteASN

	// break case: neighboring peer is internal -- this is not an EPENode
	remoteHasInternalASN := checkASNLocation(peerASN)
	if peerASN == int32(db.ASN) || remoteHasInternalASN == true {
		glog.Infof("Current peer message's neighbor ASN is a local ASN: this is not an EPENode -- skipping")
		return
	}

	epeNodeExists := db.CheckExistingEPENode(nodeID)
	if epeNodeExists {
		db.UpdateExistingPeerIP(nodeID, peerIP)
	} else {
		epeNodeDocument := &database.EPENode{
			RouterID: nodeID,
			PeerIP:   []string{peerIP},
			ASN:      nodeASN,
		}
		if action == "up" {
			if err := db.Upsert(epeNodeDocument); err != nil {
				glog.Errorf("Encountered an error while upserting the epe node document: %+v", err)
				return
			}
			glog.Infof("Successfully added epe node document: %q with peer: %q\n", nodeID, peerIP)
		} else {
			if err := db.Delete(epeNodeDocument); err != nil {
				glog.Errorf("Encountered an error while deleting the epe node document: %+v", err)
				return
			} else {
				glog.Infof("Successfully deleted epe node document: %q with peer: %q\n", nodeID, peerIP)
			}
		}

	}
}
