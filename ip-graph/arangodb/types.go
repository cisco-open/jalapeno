// Copyright (c) 2022-2025 Cisco Systems, Inc. and its affiliates
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// The contents of this file are licensed under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with the
// License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package arangodb

import (
	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/bgp"
	"github.com/sbezverk/gobmp/pkg/bgpls"
	"github.com/sbezverk/gobmp/pkg/sr"
	"github.com/sbezverk/gobmp/pkg/srv6"
)

// Config holds configuration for the IP graph processor
type Config struct {
	DatabaseServer string
	User           string
	Password       string
	Database       string
	// IGP source collections
	IGPv4Graph string
	IGPv6Graph string
	IGPNode    string
	IGPDomain  string
	// IP graph collections (full topology)
	IPv4Graph string
	IPv6Graph string
	// BGP collections
	BGPNode     string
	BGPPrefixV4 string
	BGPPrefixV6 string
	// Performance settings
	BatchSize         int
	ConcurrentWorkers int
}

// IPGraphObject represents an edge in the full IP topology graph (extends IGP graph)
type IPGraphObject struct {
	Key                   string      `json:"_key"`
	From                  string      `json:"_from"`
	To                    string      `json:"_to"`
	Link                  string      `json:"link"`
	ProtocolID            interface{} `json:"protocol_id"`
	DomainID              interface{} `json:"domain_id"`
	MTID                  uint16      `json:"mt_id"`
	AreaID                string      `json:"area_id"`
	Protocol              string      `json:"protocol"`
	LocalNodeASN          uint32      `json:"local_node_asn"`
	RemoteNodeASN         uint32      `json:"remote_node_asn"`
	LocalLinkID           uint32      `json:"local_link_id"`
	RemoteLinkID          uint32      `json:"remote_link_id"`
	LocalLinkIP           string      `json:"local_link_ip"`
	RemoteLinkIP          string      `json:"remote_link_ip"`
	IGPMetric             uint32      `json:"igp_metric"`
	MaxLinkBWKbps         uint64      `json:"max_link_bw_kbps"`
	SRv6EndXSID           interface{} `json:"srv6_endx_sid"`
	LSAdjacencySID        interface{} `json:"ls_adjacency_sid"`
	UnidirLinkDelayMinMax interface{} `json:"unidir_link_delay_min_max"`
	AppSpecLinkAttr       interface{} `json:"app_spec_link_attr"`
	// BGP-specific fields for BGP peer sessions
	LocalBGPID  string              `json:"local_bgp_id,omitempty"`
	RemoteBGPID string              `json:"remote_bgp_id,omitempty"`
	LocalIP     string              `json:"local_ip,omitempty"`
	RemoteIP    string              `json:"remote_ip,omitempty"`
	BaseAttrs   *bgp.BaseAttributes `json:"base_attrs,omitempty"`
	PeerASN     uint32              `json:"peer_asn,omitempty"`
	OriginAS    int32               `json:"origin_as,omitempty"`
	Nexthop     string              `json:"nexthop,omitempty"`
	Labels      []uint32            `json:"labels,omitempty"`
	// Prefix-specific fields for BGP prefix vertices
	Prefix         string      `json:"prefix,omitempty"`
	PrefixLen      int32       `json:"prefix_len,omitempty"`
	PrefixMetric   uint32      `json:"prefix_metric,omitempty"`
	PrefixAttrTLVs interface{} `json:"prefix_attr_tlvs,omitempty"`
}

// BGPNode represents a BGP peer/router in the topology
type BGPNode struct {
	Key      string `json:"_key,omitempty"`
	ID       string `json:"_id,omitempty"`
	Rev      string `json:"_rev,omitempty"`
	RouterID string `json:"router_id,omitempty"` // Use router_id to match original format
	ASN      uint32 `json:"asn"`                 // Keep as uint32 for compatibility
}

// BGPPrefix represents a BGP prefix in the topology
type BGPPrefix struct {
	Key        string              `json:"_key"`
	ID         string              `json:"_id,omitempty"`
	Rev        string              `json:"_rev,omitempty"`
	Prefix     string              `json:"prefix"`
	PrefixLen  int32               `json:"prefix_len"`
	RouterID   string              `json:"router_id,omitempty"`
	LocalIP    string              `json:"local_ip,omitempty"`
	PeerIP     string              `json:"peer_ip,omitempty"`
	PeerASN    uint32              `json:"peer_asn,omitempty"`
	OriginAS   int32               `json:"origin_as"`
	ASN        uint32              `json:"asn,omitempty"`
	LocalPref  int32               `json:"local_pref,omitempty"`
	BaseAttrs  *bgp.BaseAttributes `json:"base_attrs,omitempty"`
	ProtocolID base.ProtoID        `json:"protocol_id,omitempty"`
	Nexthop    string              `json:"nexthop,omitempty"`
	Labels     []uint32            `json:"labels,omitempty"`
	Name       string              `json:"name,omitempty"`
	PeerName   string              `json:"peer_name,omitempty"`
	PrefixType string              `json:"prefix_type"` // "ibgp", "ebgp_private", "ebgp_public", "inet"
	IsHost     bool                `json:"is_host"`     // true for /32 and /128 prefixes
}

// IPNode represents a node in the full IP topology (can be IGP, BGP, or hybrid)
type IPNode struct {
	Key                  string                          `json:"_key,omitempty"`
	ID                   string                          `json:"_id,omitempty"`
	Rev                  string                          `json:"_rev,omitempty"`
	Action               string                          `json:"action,omitempty"`
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
	BGPRouterID          string                          `json:"bgp_router_id,omitempty"`
	ASN                  uint32                          `json:"asn,omitempty"`
	LSID                 uint32                          `json:"ls_id,omitempty"`
	MTID                 []*base.MultiTopologyIdentifier `json:"mt_id_tlv,omitempty"`
	AreaID               string                          `json:"area_id"`
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
	PrefixSID            []*sr.PrefixSIDTLV              `json:"prefix_sid_tlv,omitempty"`
	FlexAlgoPrefixMetric []*bgpls.FlexAlgoPrefixMetric   `json:"flex_algo_prefix_metric,omitempty"`
	SRv6SID              string                          `json:"srv6_sid,omitempty"`
	SIDS                 []SID                           `json:"sids,omitempty"`
	Prefixes             []interface{}                   `json:"prefixes,omitempty"`
	NodeType             string                          `json:"node_type,omitempty"` // "igp", "bgp", "hybrid"
	// BGP-specific fields
	LocalBGPID      string          `json:"local_bgp_id,omitempty"`
	RemoteBGPID     string          `json:"remote_bgp_id,omitempty"`
	LocalIP         string          `json:"local_ip,omitempty"`
	RemoteIP        string          `json:"remote_ip,omitempty"`
	AdvCapabilities *bgp.Capability `json:"adv_cap,omitempty"`
	Tier            string          `json:"tier,omitempty"`
}

// SID represents a Segment Routing v6 SID associated with a node
type SID struct {
	SRv6SID              string                 `json:"srv6_sid,omitempty"`
	SRv6EndpointBehavior *srv6.EndpointBehavior `json:"srv6_endpoint_behavior,omitempty"`
	SRv6BGPPeerNodeSID   *srv6.BGPPeerNodeSID   `json:"srv6_bgp_peer_node_sid,omitempty"`
	SRv6SIDStructure     *srv6.SIDStructure     `json:"srv6_sid_structure,omitempty"`
}

// BatchOperation represents a batch operation for database writes
type BatchOperation struct {
	Type       string      `json:"type"`       // "node", "edge", "prefix"
	Action     string      `json:"action"`     // "add", "update", "delete"
	Collection string      `json:"collection"` // target collection
	Document   interface{} `json:"document"`   // document to process
	Key        string      `json:"key"`        // document key
}

// ProcessingStats represents processing statistics
type ProcessingStats struct {
	NodesProcessed    int64 `json:"nodes_processed"`
	EdgesProcessed    int64 `json:"edges_processed"`
	PrefixesProcessed int64 `json:"prefixes_processed"`
	ErrorsEncountered int64 `json:"errors_encountered"`
}
