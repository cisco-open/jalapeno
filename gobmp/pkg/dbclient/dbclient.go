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

package dbclient

import "github.com/sbezverk/gobmp/pkg/bmp"

// DB defines required methods for a database client to support
type DB interface {
	StoreMessage(msgType CollectionType, msg []byte) error
}

// Srv defines required method of a database server
type Srv interface {
	Start() error
	Stop() error
	GetInterface() DB
}

type CollectionType int

const (
	PeerStateChange CollectionType = bmp.PeerStateChangeMsg
	LSLink          CollectionType = bmp.LSLinkMsg
	LSNode          CollectionType = bmp.LSNodeMsg
	LSPrefix        CollectionType = bmp.LSPrefixMsg
	LSSRv6SID       CollectionType = bmp.LSSRv6SIDMsg
	L3VPN           CollectionType = bmp.L3VPNMsg
	L3VPNV4         CollectionType = bmp.L3VPNV4Msg
	L3VPNV6         CollectionType = bmp.L3VPNV6Msg
	UnicastPrefix   CollectionType = bmp.UnicastPrefixMsg
	UnicastPrefixV4 CollectionType = bmp.UnicastPrefixV4Msg
	UnicastPrefixV6 CollectionType = bmp.UnicastPrefixV6Msg
	SRPolicy        CollectionType = bmp.SRPolicyMsg
	SRPolicyV4      CollectionType = bmp.SRPolicyV4Msg
	SRPolicyV6      CollectionType = bmp.SRPolicyV6Msg
	Flowspec        CollectionType = bmp.FlowspecMsg
	FlowspecV4      CollectionType = bmp.FlowspecV4Msg
	FlowspecV6      CollectionType = bmp.FlowspecV6Msg
)
