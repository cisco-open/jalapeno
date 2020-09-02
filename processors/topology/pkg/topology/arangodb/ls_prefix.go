package arangodb

import (
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/jalapeno-sdn/jalapeno/pkg/topology/database"
)

func (a *arangoDB) lsPrefixHandler(obj *message.LSPrefix) {
	db := a.GetArangoDBInterface()
	action := obj.Action
	igpRouterID := obj.IGPRouterID

	//var prefixSID []*sr.PrefixSIDTLV
	//if obj.LSPrefixSID != nil {
	//	prefixSID = obj.LSPrefixSID
	//	}

	//var prefixAttrFlags base.PrefixAttrFlags
	//if obj.PrefixAttrFlags != nil {
	//	prefixAttrFlags = obj.PrefixAttrFlags
	//	}

	lsPrefixDocument := &database.LSPrefix{
		IGPRouterID: igpRouterID,
		Prefix:		      obj.Prefix,
		Length:		      obj.PrefixLen,
		Protocol:	      obj.Protocol,
		Timestamp:	      obj.Timestamp,
		PrefixSID:            obj.LSPrefixSID,
		PrefixAttrFlags:      obj.PrefixAttrFlags,
		FlexAlgoPrefixMetric: obj.FlexAlgoPrefixMetric,
	}

	if action == "add" {
		if err := db.Upsert(lsPrefixDocument); err != nil {
			glog.Errorf("Encountered an error while upserting the ls prefix document: %+v", err)
			return
		}
		glog.Infof("Successfully added ls prefix document with IGP router ID: %q, prefix: %q and length: %q\n", lsPrefixDocument.IGPRouterID, lsPrefixDocument.Prefix, lsPrefixDocument.Length)
	} else {
		if err := db.Delete(lsPrefixDocument); err != nil {
			glog.Errorf("Encountered an error while deleting the ls prefix document: %+v", err)
			return
		} else {
			glog.Infof("Successfully deleted ls prefix document with IGP Router ID: %q\n", lsPrefixDocument.IGPRouterID)
		}
	}
}

//func parseSIDIndex(SID []byte) int {
//	var data []byte
//	if len(SID) != 4 {
//		data = make([]byte, 4)
//		copy(data[4-len(SID):], SID)
//	} else {
//		data = SID
//	}
//	sidIndex := binary.BigEndian.Uint32(data)
//	return int(sidIndex)
//}

//func parseFlags(flags *sr.Flags) []string {
//	var srFlags []string
//	flagMap := map[string]bool{
//		"r": flags.R,
//		"n": flags.N,
//		"p": flags.P,
//		"e": flags.E,
//		"v": flags.V,
//		"l": flags.L,
//	}
//	for k, v := range flagMap {
//		if v == true {
//			srFlags = append(srFlags, k)
//		}
//	}
//	return srFlags
//}
