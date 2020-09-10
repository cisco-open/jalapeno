package database

import (
	"fmt"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

const LSSRv6SIDName = "LSSRv6SID"

type LSSRv6SID struct {
	Key                       string                 `json:"_key,omitempty"`
	RouterIP                  string                 `json:"router_ip,omitempty"`
	PeerIP                    string                 `json:"peer_ip,omitempty"`
	PeerASN                   int32                  `json:"peer_asn,omitempty"`
	Timestamp                 string                 `json:"timestamp,omitempty"`
	IGPRouterID               string                 `json:"igp_router_id,omitempty"`
	LocalNodeASN              uint32                 `json:"local_asn,omitempty"`
	RouterID                  string                 `json:"router_id,omitempty"`
	OSPFAreaID                string                 `json:"ospf_area_id,omitempty"`
	ISISAreaID                string                 `json:"isis_area_id,omitempty"`
	Protocol                  string                 `json:"Protocol,omitempty"`
	Nexthop                   string                 `json:"nexthop,omitempty"`
	IGPFlags                  uint8                  `json:"igp_flags,omitempty"`
	MTID                      uint16                 `json:"mtid,omitempty"`
	OSPFRouteType             uint8                  `json:"ospf_route_type,omitempty"`
	IGPRouteTag               uint8                  `json:"route_tag,omitempty"`
	IGPExtRouteTag            uint8                  `json:"ext_route_tag,omitempty"`
	OSPFFwdAddr               string                 `json:"ospf_fwd_addr,omitempty"`
	IGPMetric                 uint32                 `json:"igp_metric,omitempty"`
	Prefix                    string                 `json:"prefix,omitempty"`
	PrefixLen                 int32                  `json:"prefix_len,omitempty"`
	SRv6SID                   []string               `json:"srv6_sid,omitempty"`
	SRv6EndpointBehavior      *srv6.EndpointBehavior `json:"srv6_endpoint_Behavior,omitempty"`
	SRv6BGPPeerNodeSID        *srv6.BGPPeerNodeSID   `json:"srv6_bgo_peer_node_sid,omitempty"`
	SRv6SIDStructure          *srv6.SIDStructure     `json:"srv6_sid_structure,omitempty"`
	//SRv6EndpointBehaviorRaw   *srv6.EndpointBehavior `json:"srv6_endpoint_Behavior,omitempty"`
	//SRv6BGPPeerNodeSIDRaw     *srv6.BGPPeerNodeSID   `json:"SRv6_BGP_Peer_Node_SID,omitempty"`
	//SRv6SIDStructureRaw       *srv6.SIDStructure     `json:"srv6_sid_structure,omitempty"`
	//SRv6EndpointBehavior      uint16                 `json:"srv6_endpoint_behavior,omitempty"`
	//SRv6Flag                  uint8                  `json:"srv6_flag,omitempty"`
	//SRv6Algorithm             uint8                  `json:"srv6_algorithm,omitempty"`
	//SRv6BGPPeerNodeSIDFlag    uint8                  `json:"SRv6BGPPeerNodeSIDFlag,omitempty"`
	//SRv6BGPPeerNodeSIDWeight  uint8                  `json:"SRv6BGPPeerNodeSIDWeight,omitempty"`
	//SRv6BGPPeerNodeSIDPeerASN uint32                 `json:"SRv6BGPPeerNodeSIDPeerASN,omitempty"`
	//SRv6BGPPeerNodeSIDID      []byte                 `json:"SRv6BGPPeerNodeSIDID,omitempty"`
	//SRv6SIDStructureLBLength  uint8                  `json:"SRv6SIDStructureLBLength,omitempty"`
	//SRv6SIDStructureLNLength  uint8                  `json:"SRv6SIDStructureLNLength,omitempty"`
	//SRv6SIDStructureFunLength uint8                  `json:"SRv6SIDStructureFunLength,omitempty"`
	//SRv6SIDStructureArgLength uint8                  `json:"SRv6SIDStructureArgLength,omitempty"`
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
