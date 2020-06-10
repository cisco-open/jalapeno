package arangodb

import (
        "github.com/golang/glog"
        "github.com/sbezverk/gobmp/pkg/message"
        "github.com/sbezverk/gobmp/pkg/sr"
        "github.com/sbezverk/gobmp/pkg/topology/database"
        "encoding/binary"
)

func (a *arangoDB) lsPrefixHandler(obj *message.LSPrefix) {
        db := a.GetArangoDBInterface()
        action := obj.Action
        igpRouterID := obj.IGPRouterID
        prefixSID := obj.LSPrefixSID

        var algorithm *uint8
        var srFlags []string
        var sid []byte
        if(prefixSID != nil) {
                algorithm = &prefixSID.Algorithm
                srFlags = parseFlags(prefixSID.Flags)
                sid = prefixSID.SID
        } 

        if(sid != nil) {
                prefixSIDIndex := parseSIDIndex(sid)
	        lsPrefixKey := obj.IGPRouterID + "_" + obj.Prefix
                lsPrefixIndexSliceExists := db.CheckExistingLSPrefixIndexSlice(lsPrefixKey)
                if (lsPrefixIndexSliceExists) {
                        db.UpdateExistingLSPrefixIndexSlice(lsPrefixKey, prefixSIDIndex)
                } else {
                        db.CreateLSPrefixIndexSlice(lsPrefixKey, prefixSIDIndex)
                }
        }
         
        lsPrefixDocument := &database.LSPrefix{
                IGPRouterID:  igpRouterID,
                Prefix:       obj.Prefix,
                Length:       obj.PrefixLen,
                Protocol:     obj.Protocol,
                Timestamp:    obj.Timestamp,
                SRFlags:      srFlags,
                Algorithm:    algorithm,
        }
        if (action == "add") {
                if err := db.Upsert(lsPrefixDocument); err != nil {
                        glog.Errorf("Encountered an error while upserting the ls prefix document: %+v", err)
                        return
                }
                glog.Infof("Successfully added ls prefix document with IGP router ID: %q, prefix: %q and SRFlag: %q\n", lsPrefixDocument.IGPRouterID, lsPrefixDocument.Prefix, lsPrefixDocument.SRFlags)
        } else {
                if err := db.Delete(lsPrefixDocument); err != nil {
                        glog.Errorf("Encountered an error while deleting the ls prefix document: %+v", err)
                        return
                } else {
                        glog.Infof("Successfully deleted ls prefix document with IGP Router ID: %q\n", lsPrefixDocument.IGPRouterID)
                }
        }
}


func parseSIDIndex(SID []byte) int {
        var data []byte
        if(len(SID) != 4) {
                data = make([]byte,4)
                copy(data[4-len(SID):], SID)
        } else {
                data = SID
        }
        sidIndex := binary.BigEndian.Uint32(data)
        return int(sidIndex)
}

func parseFlags(flags *sr.Flags) []string{
	var srFlags []string
	flagMap := map[string]bool{
		"r": flags.R,
		"n": flags.N,
		"p": flags.P,
		"e": flags.E,
		"v": flags.V,
		"l": flags.L,
        }
	for k, v:= range flagMap {
		if(v == true) {
			srFlags = append(srFlags, k)		
		}
	}
	return srFlags
}
