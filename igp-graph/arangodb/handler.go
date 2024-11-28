// Copyright (c) 2024 Cisco Systems, Inc. and its affiliates
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//   - Redistributions of source code must retain the above copyright
//
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
	"context"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"

	"github.com/cisco-open/jalapeno/topology/kafkanotifier"
	notifier "github.com/cisco-open/jalapeno/topology/kafkanotifier"
	"github.com/sbezverk/gobmp/pkg/message"
)

func (a *arangoDB) lsNodeHandler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	// Check if Collection encoded in ls_node message ID exists
	//glog.Infof("handler received: %+v", obj)
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.lsnode.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.lsnode.Name(), c)
	}
	glog.Infof("Processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)
	var o message.LSNode
	_, err := a.lsnode.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		// In case of a LSNode removal notification, reading it will return Not Found error
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
		}
		// If operation matches to "del" then it is confirmed delete operation, otherwise return error
		if obj.Action != "del" {
			return fmt.Errorf("document %s not found but Action is not \"del\", possible stale event", obj.Key)
		}
		return a.processLSNodeExtRemoval(ctx, obj.Key)
	}
	switch obj.Action {
	case "add":
		if err := a.processLSNodeExt(ctx, obj.Key, &o); err != nil {
			return fmt.Errorf("failed to process action %s for vertex %s with error: %+v", obj.Action, obj.Key, err)
		}
	default:
		// NOOP for update
	}

	return nil
}

func (a *arangoDB) lsSRv6SIDHandler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	// Check if Collection encoded in ls_srv6_sid message ID exists
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.lssrv6sid.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.lssrv6sid.Name(), c)
	}
	glog.V(6).Infof("Processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)
	var o message.LSSRv6SID
	_, err := a.lssrv6sid.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		// In case of a ls_srv6_sid removal notification, reading it will return Not Found error
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
		}
		// If operation matches to "del" then it is confirmed delete operation, otherwise return error
		if obj.Action != "del" {
			return fmt.Errorf("document %s not found but Action is not \"del\", possible stale event", obj.Key)
		}
		glog.V(6).Infof("SRv6 SID deleted: %s for ls_node_extended key: %s ", obj.Action, obj.Key)
		return nil
	}
	switch obj.Action {
	case "add":
		if err := a.processLSSRv6SID(ctx, obj.Key, obj.ID, &o); err != nil {
			return fmt.Errorf("failed to process action %s for edge %s with error: %+v", obj.Action, obj.Key, err)
		}
	default:
		// NOOP
	}

	return nil
}

func (a *arangoDB) lsPrefixHandler(obj *kafkanotifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	// Check if Collection encoded in ls_prefix message ID exists
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.lsprefix.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.lsprefix.Name(), c)
	}
	//glog.V(5).Infof("Processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)
	var o message.LSPrefix
	_, err := a.lsprefix.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		// In case of a ls_link removal notification, reading it will return Not Found error
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
		}
		// If operation matches to "del" then it is confirmed delete operation, otherwise return error
		if obj.Action != "del" {
			return fmt.Errorf("document %s not found but Action is not \"del\", possible stale event", obj.Key)
		}

		// Detect IPv6 link by checking for ":" in the key
		if strings.Contains(obj.Key, ":") {
			return a.processv6PrefixRemoval(ctx, obj.Key, obj.Action)
		}

		err := a.processPrefixRemoval(ctx, obj.Key, obj.Action)
		if err != nil {
			return err
		}
		// write event into ls_node_edge topic
		// a.notifier.EventNotification(obj)
		// return nil
	}
	switch obj.Action {
	case "add":
		fallthrough
	case "update":
		if err := a.processPrefixSID(ctx, obj.Key, obj.ID, o); err != nil {
			return fmt.Errorf("failed to process action %s for vertex %s with error: %+v", obj.Action, obj.Key, err)
		}
		if err := a.processLSPrefixEdge(ctx, obj.Key, &o); err != nil {
			return fmt.Errorf("failed to process action %s for edge %s with error: %+v", obj.Action, obj.Key, err)
		}
	}

	return nil
}

func (a *arangoDB) lsLinkHandler(obj *kafkanotifier.EventMessage) error {
	ctx := context.TODO()
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}
	glog.Infof("Processing eventmessage: %+v", obj)
	// Check if Collection encoded in ls_link message ID exists
	c := strings.Split(obj.ID, "/")[0]
	if strings.Compare(c, a.lslink.Name()) != 0 {
		return fmt.Errorf("configured collection name %s and received in event collection name %s do not match", a.lslink.Name(), c)
	}
	var o message.LSLink
	_, err := a.lslink.ReadDocument(ctx, obj.Key, &o)
	if err != nil {
		// In case of a ls_link removal notification, reading it will return Not Found error
		if !driver.IsNotFound(err) {
			return fmt.Errorf("failed to read existing document %s with error: %+v", obj.Key, err)
		}
		// If operation matches to "del" then it is confirmed delete operation, otherwise return error
		if obj.Action != "del" {
			return fmt.Errorf("document %s not found but Action is not \"del\", possible stale event", obj.Key)
		}

		// Detect IPv6 link by checking for ":" in the key
		if strings.Contains(obj.Key, ":") {
			return a.processv6LinkRemoval(ctx, obj.Key, obj.Action)
		}

		return a.processLinkRemoval(ctx, obj.Key, obj.Action)
	}
	switch obj.Action {
	case "add":
		fallthrough
	case "update":
		if err := a.processLSLinkEdge(ctx, obj.Key, &o); err != nil {
			return fmt.Errorf("failed to process action %s for edge %s with error: %+v", obj.Action, obj.Key, err)
		}
	}
	//glog.V(5).Infof("Complete processing action: %s for key: %s ID: %s", obj.Action, obj.Key, obj.ID)

	// write event into ls_topoogy_v4 topic
	//a.notifier.EventNotification(obj)

	return nil
}
