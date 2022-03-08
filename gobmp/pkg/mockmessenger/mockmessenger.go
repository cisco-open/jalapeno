// Copyright (c) 2022 Cisco Systems, Inc. and its affiliates
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

package mockmessenger

import (
	"time"

	"github.com/golang/glog"
	"github.com/jalapeno/topology/pkg/dbclient"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// Srv defines required method of a processor server
type Srv interface {
	Start() error
	Stop() error
}

type mockMsg struct {
	added bool
	msgs  [2][]byte
}

type mockMessenger struct {
	store map[int]*mockMsg
	stop  chan struct{}
	db    dbclient.DB
}

// NewMockMessenger returns an instance of a mock messenger acting as a messenger server
func NewMockMessenger(db dbclient.DB) (Srv, error) {
	return &mockMessenger{
		db:   db,
		stop: make(chan struct{}),
		store: map[int]*mockMsg{
			bmp.PeerStateChangeMsg: {
				added: false,
				msgs: [2][]byte{
					[]byte(`{"action":"up","router_hash":"5efeb5cc6a67ccead6660f7c43e7a8c3","remote_bgp_id":"192.168.8.8","router_ip":"192.168.9.9","timestamp":"Apr  4 14:08:33.000006","remote_asn":5070,"remote_ip":"192.168.8.8","peer_rd":"0:0","remote_port":37772,"local_asn":5070,"local_ip":"192.168.9.9","local_port":179,"local_bgp_id":"192.168.9.9","adv_cap":"MPBGP (1) : afi=1 safi=1 : Unicast IPv4 , MPBGP (1) : afi=1 safi=4 : MPLS Labels IPv4 , MPBGP (1) : afi=1 safi=128 : MPLS-labeled VPN IPv4 , MPBGP (1) : afi=2 safi=1 : Unicast IPv6 , MPBGP (1) : afi=16388 safi=71 : BGP-LS BGP-LS , MPBGP (1) : afi=1 safi=73 :  IPv4 , Route Refresh Old (128), Route Refresh (2), 4 Octet ASN (65), ADD-PATH (69), Extended Next Hop Encoding (5)","recv_cap":"MPBGP (1) : afi=1 safi=1 : Unicast IPv4 , MPBGP (1) : afi=1 safi=4 : MPLS Labels IPv4 , MPBGP (1) : afi=1 safi=128 : MPLS-labeled VPN IPv4 , MPBGP (1) : afi=2 safi=1 : Unicast IPv6 , MPBGP (1) : afi=16388 safi=71 : BGP-LS BGP-LS , Route Refresh Old (128), Route Refresh (2), 4 Octet ASN (65), ADD-PATH (69), Extended Next Hop Encoding (5)","remote_holddown":180,"adv_holddown":180,"is_ipv4":true}`),
					[]byte(`{"action":"down","router_hash":"5efeb5cc6a67ccead6660f7c43e7a8c3","remote_bgp_id":"192.168.8.8","router_ip":"192.168.9.9","timestamp":"Apr  4 14:08:33.000006","remote_asn":5070,"remote_ip":"192.168.8.8","peer_rd":"0:0","remote_port":37772,"local_asn":5070,"local_ip":"192.168.9.9","local_port":179,"local_bgp_id":"192.168.9.9","adv_cap":"MPBGP (1) : afi=1 safi=1 : Unicast IPv4 , MPBGP (1) : afi=1 safi=4 : MPLS Labels IPv4 , MPBGP (1) : afi=1 safi=128 : MPLS-labeled VPN IPv4 , MPBGP (1) : afi=2 safi=1 : Unicast IPv6 , MPBGP (1) : afi=16388 safi=71 : BGP-LS BGP-LS , MPBGP (1) : afi=1 safi=73 :  IPv4 , Route Refresh Old (128), Route Refresh (2), 4 Octet ASN (65), ADD-PATH (69), Extended Next Hop Encoding (5)","recv_cap":"MPBGP (1) : afi=1 safi=1 : Unicast IPv4 , MPBGP (1) : afi=1 safi=4 : MPLS Labels IPv4 , MPBGP (1) : afi=1 safi=128 : MPLS-labeled VPN IPv4 , MPBGP (1) : afi=2 safi=1 : Unicast IPv6 , MPBGP (1) : afi=16388 safi=71 : BGP-LS BGP-LS , Route Refresh Old (128), Route Refresh (2), 4 Octet ASN (65), ADD-PATH (69), Extended Next Hop Encoding (5)","remote_holddown":180,"adv_holddown":180,"is_ipv4":true}`),
				},
			},
			bmp.LSNodeMsg: {
				added: false,
				msgs: [2][]byte{
					[]byte(`{"action":"add","router_hash":"5efeb5cc6a67ccead6660f7c43e7a8c3","router_ip":"192.168.9.9","base_attr_hash":"4933686d17a7b5b935252efb2cd41f16","peer_hash":"222065841d3956cc669453a009daf611","peer_ip":"192.168.8.8","peer_asn":5070,"timestamp":"Apr  4 13:55:34.000485","igp_router_id":"0000.0000.0008","router_id":"192.168.8.8","asn":5070,"mt_id":"0,2","isis_area_id":"49.0901","protocol":"IS-IS Level 2","nexthop":"192.168.8.8","local_pref":100,"name":"xrv9k-r1","ls_sr_capabilities":"80 64000:[1 134 160] ","sr_algorithm":[0,1],"sr_local_block":"00 1000:[0 58 152] ","srv6_capabilities_tlv":"0000","node_msd":"1:10"}`),
					[]byte(`{"action":"del","router_hash":"5efeb5cc6a67ccead6660f7c43e7a8c3","router_ip":"192.168.9.9","base_attr_hash":"4933686d17a7b5b935252efb2cd41f16","peer_hash":"222065841d3956cc669453a009daf611","peer_ip":"192.168.8.8","peer_asn":5070,"timestamp":"Apr  4 13:55:34.000485","igp_router_id":"0000.0000.0008","router_id":"192.168.8.8","asn":5070,"mt_id":"0,2","isis_area_id":"49.0901","protocol":"IS-IS Level 2","nexthop":"192.168.8.8","local_pref":100,"name":"xrv9k-r1","ls_sr_capabilities":"80 64000:[1 134 160] ","sr_algorithm":[0,1],"sr_local_block":"00 1000:[0 58 152] ","srv6_capabilities_tlv":"0000","node_msd":"1:10"}`),
				},
			},
			bmp.LSLinkMsg: {
				added: false,
				msgs: [2][]byte{
					[]byte(`{"action":"add","router_hash":"5efeb5cc6a67ccead6660f7c43e7a8c3","router_ip":"192.168.9.9","base_attr_hash":"4933686d17a7b5b935252efb2cd41f16","peer_hash":"222065841d3956cc669453a009daf611","peer_ip":"192.168.8.8","peer_asn":5070,"timestamp":"Apr  4 13:55:34.000485","igp_router_id":"0000.0000.0009","router_id":"192.168.9.9","protocol":"IS-IS Level 2","local_pref":100,"nexthop":"192.168.8.8","igp_metric":1,"remote_node_hash":"b0ca71813b4508962008be1bb3b73d8d","local_node_hash":"ae68e174edda04ddf80610d2bec9c522","remote_igp_router_id":"0000.0000.0008","remote_router_id":"192.168.8.8","local_node_asn":5070,"remote_node_asn":5070}`),
					[]byte(`{"action":"del","router_hash":"5efeb5cc6a67ccead6660f7c43e7a8c3","router_ip":"192.168.9.9","base_attr_hash":"4933686d17a7b5b935252efb2cd41f16","peer_hash":"222065841d3956cc669453a009daf611","peer_ip":"192.168.8.8","peer_asn":5070,"timestamp":"Apr  4 13:55:34.000485","igp_router_id":"0000.0000.0009","router_id":"192.168.9.9","protocol":"IS-IS Level 2","local_pref":100,"nexthop":"192.168.8.8","igp_metric":1,"remote_node_hash":"b0ca71813b4508962008be1bb3b73d8d","local_node_hash":"ae68e174edda04ddf80610d2bec9c522","remote_igp_router_id":"0000.0000.0008","remote_router_id":"192.168.8.8","local_node_asn":5070,"remote_node_asn":5070}`),
				},
			},
		},
	}, nil
}

func (m *mockMessenger) Start() error {
	glog.Infof("Starting mock messenger server")
	go m.messenger()

	return nil
}

func (m *mockMessenger) Stop() error {
	close(m.stop)

	return nil
}

func (m *mockMessenger) messenger() {
	peer := time.NewTicker(time.Second * 5)
	lsNode := time.NewTicker(time.Second * 2)
	lsLink := time.NewTicker(time.Second * 1)
	// unicastPrefix := time.NewTicker(time.Second * 30)
	defer func() {
		peer.Stop()
		lsNode.Stop()
		lsLink.Stop()
		// unicastPrefix.Stop()
	}()
	for {
		select {
		case <-peer.C:
			if m.store[bmp.PeerStateChangeMsg].added {
				glog.V(5).Infof("peer delete")
				m.db.StoreMessage(bmp.PeerStateChangeMsg, m.store[bmp.PeerStateChangeMsg].msgs[1])
				m.store[bmp.PeerStateChangeMsg].added = false
			} else {
				glog.V(5).Infof("peer add")
				m.db.StoreMessage(bmp.PeerStateChangeMsg, m.store[bmp.PeerStateChangeMsg].msgs[0])
				m.store[bmp.PeerStateChangeMsg].added = true
			}
		case <-lsNode.C:
			if m.store[bmp.LSNodeMsg].added {
				glog.V(5).Infof("ls node delete")
				m.db.StoreMessage(bmp.LSNodeMsg, m.store[bmp.LSNodeMsg].msgs[1])
				m.store[bmp.LSNodeMsg].added = false
			} else {
				glog.V(5).Infof("ls node add")
				m.db.StoreMessage(bmp.LSNodeMsg, m.store[bmp.LSNodeMsg].msgs[0])
				m.store[bmp.LSNodeMsg].added = true
			}
		case <-lsLink.C:
			if m.store[bmp.LSLinkMsg].added {
				glog.V(5).Infof("ls link delete")
				m.db.StoreMessage(bmp.LSLinkMsg, m.store[bmp.LSLinkMsg].msgs[1])
				m.store[bmp.LSLinkMsg].added = false
			} else {
				glog.V(5).Infof("ls link add")
				m.db.StoreMessage(bmp.LSLinkMsg, m.store[bmp.LSLinkMsg].msgs[0])
				m.store[bmp.LSLinkMsg].added = true
			}
		// case <-unicastPrefix.C:
		// 	if m.store[bmp.UnicastPrefixMsg].added {
		// 		m.db.StoreMessage(bmp.UnicastPrefixMsg, m.store[bmp.UnicastPrefixMsg].msgs[1])
		// 		m.store[bmp.UnicastPrefixMsg].added = false
		// 	} else {
		// 		m.db.StoreMessage(bmp.UnicastPrefixMsg, m.store[bmp.UnicastPrefixMsg].msgs[0])
		// 		m.store[bmp.UnicastPrefixMsg].added = true
		// 	}
		case <-m.stop:
			return
		}
	}
}
