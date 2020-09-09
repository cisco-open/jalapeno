package database

import (
	"fmt"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

const LSSRv6SIDName = "LSSRv6SID"

type LSSRv6SID struct {
	Key                       string                 `json:"_key,omitempty"`
	RouterIP                  string                 `json:"RouterIP,omitempty"`
	PeerIP                    string                 `json:"PeerIP,omitempty"`
	PeerASN                   int32                  `json:"PeerASN,omitempty"`
	Timestamp                 string                 `json:"timestamp,omitempty"`
	IGPRouterID               string                 `json:"IGPRouterID,omitempty"`
	LocalASN                  uint32                 `json:"LocalASN,omitempty"`
	Protocol                  string                 `json:"Protocol,omitempty"`
	RouterID                  string                 `json:"RouterID,omitempty"`
	IGPFlags                  uint8                  `json:"IGPFlags,omitempty"`
	MTID                      uint16               `json:"MT_ID,omitempty"`
	Nexthop                   string                 `json:"Nexthop,omitempty"`
	SRv6SID                   []string               `json:"SRv6_SID,omitempty"`
	SRv6EndpointBehaviorRaw   *srv6.EndpointBehavior `json:"SRv6_Endpoint_Behavior,omitempty"`
	SRv6BGPPeerNodeSIDRaw     *srv6.BGPPeerNodeSID   `json:"SRv6_BGP_Peer_Node_SID,omitempty"`
	SRv6SIDStructureRaw       *srv6.SIDStructure     `json:"SRv6_SID_Structure,omitempty"`
	SRv6EndpointBehavior      uint16                 `json:"SRv6EndpointBehavior,omitempty"`
	SRv6Flag                  uint8                  `json:"SRv6Flag,omitempty"`
	SRv6Algorithm             uint8                  `json:"SRv6Algorithm,omitempty"`
	SRv6BGPPeerNodeSIDFlag    uint8                  `json:"SRv6BGPPeerNodeSIDFlag,omitempty"`
	SRv6BGPPeerNodeSIDWeight  uint8                  `json:"SRv6BGPPeerNodeSIDWeight,omitempty"`
	SRv6BGPPeerNodeSIDPeerASN uint32                 `json:"SRv6BGPPeerNodeSIDPeerASN,omitempty"`
	SRv6BGPPeerNodeSIDID      []byte                 `json:"SRv6BGPPeerNodeSIDID,omitempty"`
	SRv6SIDStructureLBLength  uint8                  `json:"SRv6SIDStructureLBLength,omitempty"`
	SRv6SIDStructureLNLength  uint8                  `json:"SRv6SIDStructureLNLength,omitempty"`
	SRv6SIDStructureFunLength uint8                  `json:"SRv6SIDStructureFunLength,omitempty"`
	SRv6SIDStructureArgLength uint8                  `json:"SRv6SIDStructureArgLength,omitempty"`
}

func (r LSSRv6SID) GetKey() (string, error) {
	if r.Key == "" {
		return r.makeKey()
	}
	return r.Key, nil
}

func (r *LSSRv6SID) SetKey() error {
	k, err := r.makeKey()
	if err != nil {
		return err
	}
	r.Key = k
	return nil
}

func (r *LSSRv6SID) makeKey() (string, error) {
	err := ErrKeyInvalid
	ret := ""
	if r.IGPRouterID != "" {
		ret = fmt.Sprintf("%s", r.IGPRouterID)
		err = nil
	}
	return ret, err
}

func (r LSSRv6SID) GetType() string {
	return LSSRv6SIDName
}
