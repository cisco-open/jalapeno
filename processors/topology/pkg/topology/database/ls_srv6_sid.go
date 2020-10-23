package database

import (
	"fmt"

	"github.com/sbezverk/gobmp/pkg/srv6"
)

const LSSRv6SIDName = "LSSRv6SIDDemo"

type LSSRv6SID struct {
	Key                  string                 `json:"_key,omitempty"`
	ID                   string                 `json:"_id,omitempty"`
	Rev                  string                 `json:"_rev,omitempty"`
	Action               string                 `json:"action,omitempty"`
	Sequence             int                    `json:"sequence,omitempty"`
	Hash                 string                 `json:"hash,omitempty"`
	RouterHash           string                 `json:"router_hash,omitempty"`
	RouterIP             string                 `json:"router_ip,omitempty"`
	DomainID             int64                  `json:"domain_id"`
	PeerHash             string                 `json:"peer_hash,omitempty"`
	PeerIP               string                 `json:"peer_ip,omitempty"`
	PeerASN              int32                  `json:"peer_asn,omitempty"`
	Timestamp            string                 `json:"timestamp,omitempty"`
	IGPRouterID          string                 `json:"igp_router_id,omitempty"`
	LocalNodeASN         uint32                 `json:"local_node_asn,omitempty"`
	RouterID             string                 `json:"router_id,omitempty"`
	LSID                 uint32                 `json:"ls_id,omitempty"`
	OSPFAreaID           string                 `json:"ospf_area_id,omitempty"`
	ISISAreaID           string                 `json:"isis_area_id,omitempty"`
	Protocol             string                 `json:"protocol,omitempty"`
	Nexthop              string                 `json:"nexthop,omitempty"`
	LocalNodeHash        string                 `json:"local_node_hash,omitempty"`
	MTID                 uint16                 `json:"mt_id,omitempty"`
	IGPFlags             uint8                  `json:"igp_flags"`
	IGPRouteTag          uint8                  `json:"route_tag,omitempty"`
	IGPExtRouteTag       uint8                  `json:"ext_route_tag,omitempty"`
	OSPFFwdAddr          string                 `json:"ospf_fwd_addr,omitempty"`
	IGPMetric            uint32                 `json:"igp_metric,omitempty"`
	Prefix               string                 `json:"prefix,omitempty"`
	PrefixLen            int32                  `json:"prefix_len,omitempty"`
	IsPrepolicy          bool                   `json:"isprepolicy"`
	IsAdjRIBIn           bool                   `json:"is_adj_rib_in"`
	SRv6SID              string                 `json:"srv6_sid,omitempty"`
	SRv6EndpointBehavior *srv6.EndpointBehavior `json:"srv6_endpoint_behavior,omitempty"`
	SRv6BGPPeerNodeSID   *srv6.BGPPeerNodeSID   `json:"srv6_bgp_peer_node_sid,omitempty"`
	SRv6SIDStructure     *srv6.SIDStructure     `json:"srv6_sid_structure,omitempty"`
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
