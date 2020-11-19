package arangodb

import (
	"context"
	"encoding/json"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/topology/pkg/dbclient"
	"github.com/sbezverk/topology/pkg/kafkanotifier"
	"go.uber.org/atomic"
)

type DBRecord interface {
	MakeKey() string
}

type result struct {
	key    string
	action string
	err    error
}

type queueMsg struct {
	msgType dbclient.CollectionType
	msgData []byte
}

type stats struct {
	total  atomic.Int64
	failed atomic.Int64
}

type collection struct {
	queue           chan *queueMsg
	stats           *stats
	stop            chan struct{}
	topicCollection driver.Collection
	collectionType dbclient.CollectionType
	handler    func()
	arango     *arangoDB
	properties *collectionProperties
}

func (c *collection) processError(r *result) bool {
	switch {
	// Condition when a collection was deleted while the topology was running
	case driver.IsArangoErrorWithErrorNum(r.err, driver.ErrArangoDataSourceNotFound):
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
	for {
		select {
		case m := <-c.queue:
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
			if r.err != nil {
				// Error was encountered during processing of the key
				if c.processError(r) {
					glog.Errorf("genericWorker for key: %s reported a fatal error: %+v", r.key, r.err)
				}
				glog.Errorf("genericWorker for key: %s reported a non fatal error: %+v", r.key, r.err)
			}
			if c.arango.notifyCompletion && c.arango.notifier != nil {
				if err := c.arango.notifier.EventNotification(&kafkanotifier.EventMessage{
					TopicType: c.collectionType,
					Key:       r.key,
					ID:        c.properties.name + "/" + r.key,
					Action:    r.action,
				}); err != nil {
					glog.Errorf("genericWorker for key: %s failed to send notification with error: %+v", r.key, r.err)
				}
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
				go c.genericWorker(r.key, bo.(DBRecord), done, tokens)
			}
			// If Backlog for a specific key is empty, remove it from the backlog
			if b.Len() == 0 {
				delete(backlog, r.key)
			}
		case <-c.stop:
			return
		}
	}
}

func (c *collection) genericWorker(k string, o DBRecord, done chan *result, tokens chan struct{}) {
	var err error
	var action string
	defer func() {
		<-tokens
		done <- &result{key: k, action: action, err: err}
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
		action = obj.(*peerStateChangeArangoMessage).Action
	case bmp.LSLinkMsg:
		obj, ok = o.(*lsLinkArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover lsLinkArangoMessage from DBRecord interface")
			return
		}
		obj.(*lsLinkArangoMessage).Key = k
		obj.(*lsLinkArangoMessage).ID = c.properties.name + "/" + k
		action = obj.(*lsLinkArangoMessage).Action
	case bmp.LSNodeMsg:
		obj, ok = o.(*lsNodeArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover lsNodeArangoMessage from DBRecord interface")
			return
		}
		obj.(*lsNodeArangoMessage).Key = k
		obj.(*lsNodeArangoMessage).ID = c.properties.name + "/" + k
		action = obj.(*lsNodeArangoMessage).Action
	case bmp.LSPrefixMsg:
		obj, ok = o.(*lsPrefixArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover lsPrefixArangoMessage from DBRecord interface")
			return
		}
		obj.(*lsPrefixArangoMessage).Key = k
		obj.(*lsPrefixArangoMessage).ID = c.properties.name + "/" + k
		action = obj.(*lsPrefixArangoMessage).Action
	case bmp.LSSRv6SIDMsg:
		obj, ok = o.(*lsSRv6SIDArangoMessage)
		if !ok {
			err = fmt.Errorf("failed to recover lsSRv6SIDArangoMessage from DBRecord interface")
			return
		}
		obj.(*lsSRv6SIDArangoMessage).Key = k
		obj.(*lsSRv6SIDArangoMessage).ID = c.properties.name + "/" + k
		action = obj.(*lsSRv6SIDArangoMessage).Action
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
		action = obj.(*l3VPNArangoMessage).Action
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
		action = obj.(*unicastPrefixArangoMessage).Action
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
		action = obj.(*srPolicyArangoMessage).Action
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
				break
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

	return
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
		return &o, nil
	}

	return nil, fmt.Errorf("unknown collection type %d", collectionType)
}
