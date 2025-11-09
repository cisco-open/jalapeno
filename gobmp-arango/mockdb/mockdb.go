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

package mockdb

import (
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/message"
)

type mockDB struct {
	stop chan struct{}
	dbclient.DB
}

// NewDBSrvClient returns an instance of a mock DB server client process
func NewDBSrvClient() (dbclient.Srv, error) {
	m := &mockDB{
		stop: make(chan struct{}),
	}
	m.DB = m

	return m, nil
}

func (m *mockDB) Start() error {
	glog.Info("Starting Mock DB Client")
	return nil
}

func (m *mockDB) Stop() error {
	close(m.stop)

	return nil
}

func (m *mockDB) GetInterface() dbclient.DB {
	return m.DB
}

func (m *mockDB) StoreMessage(msgType dbclient.CollectionType, msg []byte) error {
	// switch msgType {
	// case bmp.PeerStateChangeMsg:
	// 	p, ok := msg.(*message.PeerStateChange)
	// 	if !ok {
	// 		return fmt.Errorf("malformed PeerStateChange message")
	// 	}
	// 	m.peerChangeHandler(p)
	// case bmp.UnicastPrefixMsg:
	// 	un, ok := msg.(*message.UnicastPrefix)
	// 	if !ok {
	// 		return fmt.Errorf("malformed UnicastPrefix message")
	// 	}
	// 	m.unicastPrefixHandler(un)
	// case bmp.LSNodeMsg:
	// 	ln, ok := msg.(*message.LSNode)
	// 	if !ok {
	// 		return fmt.Errorf("malformed LSNode message")
	// 	}
	// 	m.lsNodeHandler(ln)
	// case bmp.LSLinkMsg:
	// 	ll, ok := msg.(*message.LSLink)
	// 	if !ok {
	// 		return fmt.Errorf("malformed LSLink message")
	// 	}
	// 	m.lsLinkHandler(ll)
	// case bmp.L3VPNMsg:
	// 	l3, ok := msg.(*message.L3VPNPrefix)
	// 	if !ok {
	// 		return fmt.Errorf("malformed L3VPN message")
	// 	}
	// 	m.l3vpnPrefixHandler(l3)
	// }

	return nil
}

func (m *mockDB) peerChangeHandler(obj *message.PeerStateChange) {
	glog.V(5).Infof("peer change handler")
}

func (m *mockDB) unicastPrefixHandler(obj *message.UnicastPrefix) {
	glog.V(5).Infof("unicast prefix handler")
}

func (m *mockDB) lsNodeHandler(obj *message.LSNode) {
	glog.V(5).Infof("LS Node handler")
}

func (m *mockDB) lsLinkHandler(obj *message.LSLink) {
	glog.V(5).Infof("LS Link handler")
}

func (m *mockDB) l3vpnPrefixHandler(obj *message.L3VPNPrefix) {
	glog.V(5).Infof("L3VPNPrefix handler")
}
