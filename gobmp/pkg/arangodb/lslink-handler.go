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

package arangodb

import (
	"strconv"

	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/message"
)

type lsLinkArangoMessage struct {
	*message.LSLink
}

func (l *lsLinkArangoMessage) MakeKey() string {
	var localID, remoteID string
	mtid := 0
	if l.MTID != nil {
		mtid = int(l.MTID.MTID)
	}
	if l.LocalLinkIP != "" && l.RemoteLinkIP != "" {
		localID = l.LocalLinkIP
		remoteID = l.RemoteLinkIP
	} else {
		// If Local/Remote IP are missing as in the case of unnumbered links,
		// then links are identified by Local/Remote Link IDs which are presented in the inverse dotted notation
		// Example: Link ID 14 is presented as 0.0.0.14
		localID = strconv.Itoa(int(l.LocalLinkID>>24)&0x000000ff) + "." +
			strconv.Itoa(int(l.LocalLinkID>>16)&0x000000ff) + "." +
			strconv.Itoa(int(l.LocalLinkID>>8)&0x000000ff) + "." +
			strconv.Itoa(int(l.LocalLinkID)&0x000000ff)
		remoteID = strconv.Itoa(int(l.RemoteLinkID>>24)&0x000000ff) + "." +
			strconv.Itoa(int(l.RemoteLinkID>>16)&0x000000ff) + "." +
			strconv.Itoa(int(l.RemoteLinkID>>8)&0x000000ff) + "." +
			strconv.Itoa(int(l.RemoteLinkID)&0x000000ff)
	}
	routerID := l.IGPRouterID
	remoteRouterID := l.RemoteIGPRouterID
	// If Protocol ID == 7 (BGP) LS Link does not carry IGP Router ID field, instead BGP Router ID must be used
	if l.ProtocolID == base.BGP {
		routerID = l.BGPRouterID
		remoteRouterID = l.BGPRemoteRouterID
	}

	// if len(l.LSAdjacencySID) != 0 {
	// 	glog.Infof("handling lslink local IP: %+v, remote IP: %+v, protocol: %+v, with Adj SID: %+v ", l.LocalLinkIP, l.RemoteLinkIP, l.Protocol, l.LSAdjacencySID[0].SID)
	// }

	// The ls_link Key uses ProtocolID, DomainID, and Multi-Topology ID
	// to create unique Keys for DB entries in multi-area / multi-topology scenarios

	return strconv.Itoa(int(l.ProtocolID)) + "_" + strconv.Itoa(int(l.DomainID)) + "_" + strconv.Itoa(mtid) + "_" + l.AreaID + "_" + routerID + "_" + localID + "_" + remoteRouterID + "_" + remoteID
}
