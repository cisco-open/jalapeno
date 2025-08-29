package arangodb

import (
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/bgpls"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

type duplicateNode struct {
	Key         string       `json:"_key,omitempty"`
	DomainID    int64        `json:"domain_id"`
	IGPRouterID string       `json:"igp_router_id,omitempty"`
	Protocol    string       `json:"protocol,omitempty"`
	ProtocolID  base.ProtoID `json:"protocol_id,omitempty"`
	Name        string       `json:"name,omitempty"`
}

type srObject struct {
	PrefixAttrTLVs *bgpls.PrefixAttrTLVs `json:"prefix_attr_tlvs,omitempty"`
}

type igpNode struct {
	Key                  string                          `json:"_key,omitempty"`
	ID                   string                          `json:"_id,omitempty"`
	Rev                  string                          `json:"_rev,omitempty"`
	Action               string                          `json:"action,omitempty"` // Action can be "add" or "del"
	Sequence             int                             `json:"sequence,omitempty"`
	Hash                 string                          `json:"hash,omitempty"`
	RouterHash           string                          `json:"router_hash,omitempty"`
	DomainID             int64                           `json:"domain_id"`
	RouterIP             string                          `json:"router_ip,omitempty"`
	PeerHash             string                          `json:"peer_hash,omitempty"`
	PeerIP               string                          `json:"peer_ip,omitempty"`
	PeerASN              uint32                          `json:"peer_asn,omitempty"`
	Timestamp            string                          `json:"timestamp,omitempty"`
	IGPRouterID          string                          `json:"igp_router_id,omitempty"`
	RouterID             string                          `json:"router_id,omitempty"`
	ASN                  uint32                          `json:"asn,omitempty"`
	LSID                 uint32                          `json:"ls_id,omitempty"`
	MTID                 []*base.MultiTopologyIdentifier `json:"mt_id_tlv,omitempty"`
	Protocol             string                          `json:"protocol,omitempty"`
	ProtocolID           base.ProtoID                    `json:"protocol_id,omitempty"`
	NodeFlags            *bgpls.NodeAttrFlags            `json:"node_flags,omitempty"`
	Name                 string                          `json:"name,omitempty"`
	SRCapabilities       *sr.Capability                  `json:"ls_sr_capabilities,omitempty"`
	SRAlgorithm          []int                           `json:"sr_algorithm,omitempty"`
	SRLocalBlock         *sr.LocalBlock                  `json:"sr_local_block,omitempty"`
	SRv6CapabilitiesTLV  *srv6.CapabilityTLV             `json:"srv6_capabilities_tlv,omitempty"`
	NodeMSD              []*base.MSDTV                   `json:"node_msd,omitempty"`
	FlexAlgoDefinition   []*bgpls.FlexAlgoDefinition     `json:"flex_algo_definition,omitempty"`
	IsPrepolicy          bool                            `json:"is_prepolicy"`
	IsAdjRIBIn           bool                            `json:"is_adj_rib_in"`
	Prefix               string                          `json:"prefix,omitempty"`
	PrefixLen            int32                           `json:"prefix_len,omitempty"`
	PrefixAttrTLVs       *bgpls.PrefixAttrTLVs           `json:"prefix_attr_tlvs,omitempty"`
	PrefixSID            []*sr.PrefixSIDTLV              `json:"prefix_sid_tlv,omitempty"`
	FlexAlgoPrefixMetric []*bgpls.FlexAlgoPrefixMetric   `json:"flex_algo_prefix_metric,omitempty"`
	SRv6SID              string                          `json:"srv6_sid,omitempty"`
	SIDS                 []SID                           `json:"sids,omitempty"`
}

type SID struct {
	SRv6SID              string                 `json:"srv6_sid,omitempty"`
	SRv6EndpointBehavior *srv6.EndpointBehavior `json:"srv6_endpoint_behavior,omitempty"`
	SRv6BGPPeerNodeSID   *srv6.BGPPeerNodeSID   `json:"srv6_bgp_peer_node_sid,omitempty"`
	SRv6SIDStructure     *srv6.SIDStructure     `json:"srv6_sid_structure,omitempty"`
}

type peerObject struct {
	BGPRouterID string `json:"bgp_router_id,omitempty"`
}

type lsGraphObject struct {
	Key                   string                `json:"_key"`
	From                  string                `json:"_from"`
	To                    string                `json:"_to"`
	Link                  string                `json:"link"`
	ProtocolID            base.ProtoID          `json:"protocol_id"`
	DomainID              int64                 `json:"domain_id"`
	MTID                  uint16                `json:"mt_id"`
	AreaID                string                `json:"area_id"`
	Protocol              string                `json:"protocol"`
	LocalLinkID           uint32                `json:"local_link_id"`
	RemoteLinkID          uint32                `json:"remote_link_id"`
	LocalLinkIP           string                `json:"local_link_ip"`
	RemoteLinkIP          string                `json:"remote_link_ip"`
	LocalNodeASN          uint32                `json:"local_node_asn"`
	RemoteNodeASN         uint32                `json:"remote_node_asn"`
	PeerNodeSID           *sr.PeerSID           `json:"peer_node_sid,omitempty"`
	PeerAdjSID            *sr.PeerSID           `json:"peer_adj_sid,omitempty"`
	PeerSetSID            *sr.PeerSID           `json:"peer_set_sid,omitempty"`
	SRv6BGPPeerNodeSID    *srv6.BGPPeerNodeSID  `json:"srv6_bgp_peer_node_sid,omitempty"`
	SRv6ENDXSID           []*srv6.EndXSIDTLV    `json:"srv6_endx_sid,omitempty"`
	LSAdjacencySID        []*sr.AdjacencySIDTLV `json:"ls_adjacency_sid,omitempty"`
	UnidirLinkDelay       uint32                `json:"unidir_link_delay"`
	UnidirLinkDelayMinMax []uint32              `json:"unidir_link_delay_min_max"`
	UnidirDelayVariation  uint32                `json:"unidir_delay_variation,omitempty"`
	UnidirPacketLoss      uint32                `json:"unidir_packet_loss,omitempty"`
	UnidirResidualBW      uint32                `json:"unidir_residual_bw,omitempty"`
	UnidirAvailableBW     uint32                `json:"unidir_available_bw,omitempty"`
	UnidirBWUtilization   uint32                `json:"unidir_bw_utilization,omitempty"`
	Prefix                string                `json:"prefix"`
	PrefixLen             int32                 `json:"prefix_len"`
	PrefixMetric          uint32                `json:"prefix_metric"`
	PrefixAttrTLVs        *bgpls.PrefixAttrTLVs `json:"prefix_attr_tlvs"`
}
