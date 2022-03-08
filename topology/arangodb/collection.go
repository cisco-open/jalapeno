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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/jalapeno/topology/pkg/dbclient"
	"github.com/jalapeno/topology/pkg/kafkanotifier"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"go.uber.org/atomic"
)

const (
	addAction     actionType = "add"
	updateAction  actionType = "update"
	delAction     actionType = "del"
	unknownAction actionType = "unknown"
)

type actionType string

func newAction(a string) actionType {
	switch strings.ToLower(a) {
	case "add":
		return addAction
	case "update":
		return updateAction
	case "del":
		return delAction
	default:
		return unknownAction
	}
}

type DBRecord interface {
	MakeKey() string
}

type result struct {
	object DBRecord
	key    string
	action actionType
	err    error
}

type queueMsg struct {
	msgType dbclient.CollectionType
	msgData []byte
}

type stats struct {
	total atomic.Int64
}

type collection struct {
	queue           chan *queueMsg
	stats           *stats
	stop            chan struct{}
	topicCollection driver.Collection
	collectionType  dbclient.CollectionType
	handler         func()
	arango          *arangoDB
	properties      *collectionProperties
}

const (
	// ErrArangoGraphNotFound defines Arango DB error to return when requested Gra[h is not found
	// This error is not defined by ArangoDB go driver module.
	ErrArangoGraphNotFound = 1924
)

func (c *collection) processError(r *result) bool {
	switch {
	// Condition when a collection was deleted while the topology was running
	case driver.IsArangoErrorWithErrorNum(r.err, driver.ErrArangoDataSourceNotFound, ErrArangoGraphNotFound):
		if err := c.arango.ensureCollection(c.properties, c.collectionType); err != nil {
			return true
		}
		return false
	case driver.IsPreconditionFailed(r.err):
		glog.Errorf("precondition for %+v failed", r.key)
		return false
	default:
		glog.Errorf("failed to process the document for key %s with error: %+v", r.key, r.err)
		return true
	}
}

var (
	backlogCheckInterval = time.Millisecond * 500
)

func (c *collection) genericHandler() {
	glog.Infof("Starting handler for type: %d", c.collectionType)
	// keyStore is used to track duplicate key in messages, duplicate key means there is already in processing
	// a go routine for the key
	keyStore := make(map[string]bool)
	// backlog is used to store duplicate key entry until the key is released (finished processing)
	backlog := make(map[string]FIFO)
	// tokens are used to control a number of concurrent goroutine accessing the same collection, to prevent
	// conflicting database changes, each go routine processes a message with the unique key.
	tokens := make(chan struct{}, concurrentWorkers)
	done := make(chan *result, concurrentWorkers*2)
	backlogTicker := time.NewTicker(backlogCheckInterval)
	for {
		select {
		case m := <-c.queue:
			backlogTicker.Reset(backlogCheckInterval)
			o, err := newDBRecord(m.msgData, c.collectionType)
			if err != nil {
				glog.Errorf("failed to unmarshal message of type %d with error: %+v", c.collectionType, err)
				continue
			}
			k := o.MakeKey()
			busy, ok := keyStore[k]
			if ok && busy {
				// Check if there is already a backlog for this key, if not then create it
				b, ok := backlog[k]
				if !ok {
					b = newFIFO()
				}
				// Saving message in the backlog
				b.Push(o)
				backlog[k] = b
				continue
			}
			// Depositing one token and calling worker to process message for the key
			tokens <- struct{}{}
			keyStore[k] = true
			go c.genericWorker(k, o, done, tokens)
		case r := <-done:
			backlogTicker.Reset(backlogCheckInterval)
			if r.err != nil {
				// Error was encountered during processing of the key, attempting to correct the error condition
				// and pushing failed DBObject to the backlog queue
				if c.processError(r) {
					glog.Errorf("genericWorker for key: %s reported a fatal error: %+v", r.key, r.err)
				} else {
					glog.Errorf("genericWorker for key: %s reported a non fatal error: %+v", r.key, r.err)
				}
				glog.Infof("Pushing key: %s to the backlog after the failure", r.key)
				b, ok := backlog[r.key]
				if !ok {
					b = newFIFO()
				}
				// Saving message in the backlog
				b.Push(r.object)
				backlog[r.key] = b
				continue
			}
			if c.arango.notifyCompletion && c.arango.notifier != nil {
				go c.reliableNotifier(r)
				// Need to dispatch a go routine which should ensure that the record has been stored in DB and only then to send the notification

				// m := &kafkanotifier.EventMessage{
				// 	TopicType: c.collectionType,
				// 	Key:       r.key,
				// 	ID:        c.properties.name + "/" + r.key,
				// 	Action:    r.action,
				// }
				// if err := c.arango.notifier.EventNotification(m); err != nil {
				// 	glog.Errorf("genericWorker for key: %s failed to send notification with error: %+v", r.key, r.err)
				// }
			}
			delete(keyStore, r.key)
			// Check if there an entry for this key in the backlog, if there is, retrieve it and process it
			b, ok := backlog[r.key]
			if !ok {
				continue
			}
			bo := b.Pop()
			if bo != nil {
				tokens <- struct{}{}
				keyStore[r.key] = true
				go c.genericWorker(r.key, bo, done, tokens)
			}
			// If Backlog for a specific key is empty, remove it from the backlog
			if b.Len() == 0 {
				delete(backlog, r.key)
			}
		case <-c.stop:
			backlogTicker.Stop()
			return
		case <-backlogTicker.C:
			// If loop is idle for backlogCheckInterval, trying to process any outstanding items stored in the backlog
			for key, b := range backlog {
				glog.Infof("Extracted key %s from backlog (backlogTicker)", key)
				bo := b.Pop()
				if bo != nil {
					tokens <- struct{}{}
					keyStore[key] = true
					go c.genericWorker(key, bo, done, tokens)
				}
				// If Backlog for a specific key is empty, remove it from the backlog
				if b.Len() == 0 {
					delete(backlog, key)
				}
				// Processing a single item per backlogTicker cycle
				break
			}
		}
	}
}

// reliableNotifier is function which is called if Topology needs to send an event out notifying
// about topology change event.
func (c *collection) reliableNotifier(r *result) {
	// The check for the DB record state will be done every 1000 milliseconds
	ticker := time.NewTicker(time.Millisecond * 1000)
	defer func() {
		ticker.Stop()
	}()
	m := &kafkanotifier.EventMessage{
		TopicType: c.collectionType,
		Key:       r.key,
		ID:        c.properties.name + "/" + r.key,
		Action:    string(r.action),
	}
notify:
	for {
		found, err := c.topicCollection.DocumentExists(context.TODO(), r.key)
		if err != nil {
			glog.Errorf("reliableNotifier for key: %s failed to check the state of the document with error: %+v", r.key, err)
			return
		}
		switch r.action {
		case addAction:
			fallthrough
		case updateAction:
			// For add and update actions, the record must be "readable" from the collection,
			// if collection read error Document not found, it means DB has not yet written it
			// any other error terminates notifier and generate the error message
			if found {
				break notify
			}
		case delAction:
			// For del action, the read opration should return Document not found, if no error received
			//
			if !found {
				break notify
			}
		case unknownAction:
			return
		}
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			continue
		}
	}
	if err := c.arango.notifier.EventNotification(m); err != nil {
		glog.Errorf("reliableNotifier for key: %s failed to send notification with error: %+v", r.key, err)
	}
}

func (c *collection) genericWorker(k string, o DBRecord, done chan *result, tokens chan struct{}) {
	var err error
	var action actionType
	defer func() {
		<-tokens
		done <- &result{object: o, key: k, action: action, err: err}
		if err == nil {
			c.stats.total.Add(1)
		}
		glog.V(6).Infof("done key: %s, type: %d total messages: %s", k, c.collectionType, c.stats.total.String())
	}()
	ctx := context.TODO()
	var obj interface{}
	var ok bool
	switch c.collectionType {
	case bmp.PeerStateChangeMsg:
		obj, ok = o.(*peerStateChangeArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover peerStateChangeArangoMessage from DBRecord interface")
			return
		}
		obj.(*peerStateChangeArangoMessage).Key = k
		obj.(*peerStateChangeArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*peerStateChangeArangoMessage).Action)
	case bmp.LSLinkMsg:
		obj, ok = o.(*lsLinkArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover lsLinkArangoMessage from DBRecord interface")
			return
		}
		obj.(*lsLinkArangoMessage).Key = k
		obj.(*lsLinkArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*lsLinkArangoMessage).Action)
	case bmp.LSNodeMsg:
		obj, ok = o.(*lsNodeArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover lsNodeArangoMessage from DBRecord interface")
			return
		}
		obj.(*lsNodeArangoMessage).Key = k
		obj.(*lsNodeArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*lsNodeArangoMessage).Action)
	case bmp.LSPrefixMsg:
		obj, ok = o.(*lsPrefixArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover lsPrefixArangoMessage from DBRecord interface")
			return
		}
		obj.(*lsPrefixArangoMessage).Key = k
		obj.(*lsPrefixArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*lsPrefixArangoMessage).Action)
	case bmp.LSSRv6SIDMsg:
		obj, ok = o.(*lsSRv6SIDArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover lsSRv6SIDArangoMessage from DBRecord interface")
			return
		}
		obj.(*lsSRv6SIDArangoMessage).Key = k
		obj.(*lsSRv6SIDArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*lsSRv6SIDArangoMessage).Action)
	case bmp.L3VPNMsg:
		fallthrough
	case bmp.L3VPNV4Msg:
		fallthrough
	case bmp.L3VPNV6Msg:
		obj, ok = o.(*l3VPNArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover l3VPNArangoMessage from DBRecord interface")
			return
		}
		obj.(*l3VPNArangoMessage).Key = k
		obj.(*l3VPNArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*l3VPNArangoMessage).Action)
	case bmp.UnicastPrefixMsg:
		fallthrough
	case bmp.UnicastPrefixV4Msg:
		fallthrough
	case bmp.UnicastPrefixV6Msg:
		obj, ok = o.(*unicastPrefixArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover unicastPrefixArangoMessage from DBRecord interface")
			return
		}
		obj.(*unicastPrefixArangoMessage).Key = k
		obj.(*unicastPrefixArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*unicastPrefixArangoMessage).Action)
	case bmp.SRPolicyMsg:
		fallthrough
	case bmp.SRPolicyV4Msg:
		fallthrough
	case bmp.SRPolicyV6Msg:
		obj, ok = o.(*srPolicyArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover SRPolicy from DBRecord interface")
			return
		}
		obj.(*srPolicyArangoMessage).Key = k
		obj.(*srPolicyArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*srPolicyArangoMessage).Action)
	case bmp.FlowspecMsg:
		fallthrough
	case bmp.FlowspecV4Msg:
		fallthrough
	case bmp.FlowspecV6Msg:
		obj, ok = o.(*flowspecArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover Flowspec Arango Message from DBRecord interface")
			return
		}
		obj.(*flowspecArangoMessage).Key = k
		obj.(*flowspecArangoMessage).ID = c.properties.name + "/" + k
		action = newAction(obj.(*flowspecArangoMessage).Action)
	default:
		err = fmt.Errorf("unknown collection type %d", c.collectionType)
		return
	}
	switch action {
	case "add":
		if _, e := c.topicCollection.CreateDocument(ctx, obj); e != nil {
			switch {
			// The following 2 types of errors inidcate that the document by the key already
			// exists, no need to fail but instead call Update of the document.
			case driver.IsArangoErrorWithErrorNum(e, driver.ErrArangoConflict):
			case driver.IsArangoErrorWithErrorNum(e, driver.ErrArangoUniqueConstraintViolated):
			default:
				err = e
			}
			if _, e := c.topicCollection.UpdateDocument(ctx, k, obj); e != nil {
				err = e
				break
			}
			// Fixing action in result message to the actual action occured
			action = "update"
		}
	case "del":
		if _, e := c.topicCollection.RemoveDocument(ctx, k); e != nil {
			if !driver.IsArangoErrorWithErrorNum(e, driver.ErrArangoDocumentNotFound) {
				err = e
			}
		}
	}
}

func newDBRecord(msgData []byte, collectionType dbclient.CollectionType) (DBRecord, error) {
	switch collectionType {
	case bmp.PeerStateChangeMsg:
		var o peerStateChangeArangoMessage
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
		return &o, nil
	case bmp.LSLinkMsg:
		var o lsLinkArangoMessage
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
		return &o, nil
	case bmp.LSNodeMsg:
		var o lsNodeArangoMessage
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
		return &o, nil
	case bmp.LSPrefixMsg:
		var o lsPrefixArangoMessage
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
		return &o, nil
	case bmp.LSSRv6SIDMsg:
		var o lsSRv6SIDArangoMessage
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
		return &o, nil
	case bmp.L3VPNMsg:
		fallthrough
	case bmp.L3VPNV4Msg:
		fallthrough
	case bmp.L3VPNV6Msg:
		var o l3VPNArangoMessage
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
		return &o, nil
	case bmp.UnicastPrefixMsg:
		fallthrough
	case bmp.UnicastPrefixV4Msg:
		fallthrough
	case bmp.UnicastPrefixV6Msg:
		var o unicastPrefixArangoMessage
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
		return &o, nil
	case bmp.SRPolicyMsg:
		fallthrough
	case bmp.SRPolicyV4Msg:
		fallthrough
	case bmp.SRPolicyV6Msg:
		var o srPolicyArangoMessage
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
	case bmp.FlowspecMsg:
		fallthrough
	case bmp.FlowspecV4Msg:
		fallthrough
	case bmp.FlowspecV6Msg:
		o := flowspecArangoMessage{}
		if err := json.Unmarshal(msgData, &o); err != nil {
			return nil, err
		}
		return &o, nil
	}

	return nil, fmt.Errorf("unknown collection type %d", collectionType)
}
