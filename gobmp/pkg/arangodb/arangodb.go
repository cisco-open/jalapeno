package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
	"github.com/jalapeno/topology/pkg/dbclient"
	"github.com/jalapeno/topology/pkg/kafkanotifier"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/tools"
)

const (
	concurrentWorkers = 1024
)

var (
	collections = map[dbclient.CollectionType]*collectionProperties{
		dbclient.PeerStateChange: {name: "Node", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.LSLink:          {name: "LSLink", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.LSNode:          {name: "LSNode", isVertex: true, options: &driver.CreateCollectionOptions{}},
		dbclient.LSPrefix:        {name: "LSPrefix", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.LSSRv6SID:       {name: "LSSRv6SID", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.L3VPN:           {name: "L3VPN_Prefix", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.L3VPNV4:         {name: "L3VPNV4_Prefix", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.L3VPNV6:         {name: "L3VPNV6_Prefix", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.UnicastPrefix:   {name: "UnicastPrefix", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.UnicastPrefixV4: {name: "UnicastPrefixV4", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.UnicastPrefixV6: {name: "UnicastPrefixV6", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.SRPolicy:        {name: "SRPolicy", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.SRPolicyV4:      {name: "SRPolicyV4", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.SRPolicyV6:      {name: "SRPolicyV6", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.Flowspec:        {name: "Flowspec_Test", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.FlowspecV4:      {name: "FlowspecV4_Test", isVertex: false, options: &driver.CreateCollectionOptions{}},
		dbclient.FlowspecV6:      {name: "FlowspecV6_Test", isVertex: false, options: &driver.CreateCollectionOptions{}},
	}
)

// collectionProperties defines a collection specific properties
// TODO this information should be configurable without recompiling code.
type collectionProperties struct {
	name     string
	isVertex bool
	options  *driver.CreateCollectionOptions
}

type arangoDB struct {
	dbclient.DB
	*ArangoConn
	stop             chan struct{}
	collections      map[dbclient.CollectionType]*collection
	notifyCompletion bool
	notifier         kafkanotifier.Event
}

// NewDBSrvClient returns an instance of a DB server client process
func NewDBSrvClient(arangoSrv, user, pass, dbname string, notifier kafkanotifier.Event) (dbclient.Srv, error) {
	if err := tools.URLAddrValidation(arangoSrv); err != nil {
		return nil, err
	}
	arangoConn, err := NewArango(ArangoConfig{
		URL:      arangoSrv,
		User:     user,
		Password: pass,
		Database: dbname,
	})
	if err != nil {
		return nil, err
	}
	arango := &arangoDB{
		stop:        make(chan struct{}),
		collections: make(map[dbclient.CollectionType]*collection),
	}
	arango.DB = arango
	arango.ArangoConn = arangoConn
	if notifier != nil {
		arango.notifyCompletion = true
		arango.notifier = notifier
	}
	// Init collections
	for t, n := range collections {
		if err := arango.ensureCollection(n, t); err != nil {
			return nil, err
		}
	}

	return arango, nil
}

func (a *arangoDB) ensureCollection(p *collectionProperties, collectionType dbclient.CollectionType) error {
	if _, ok := a.collections[collectionType]; !ok {
		a.collections[collectionType] = &collection{
			queue:          make(chan *queueMsg),
			stats:          &stats{},
			stop:           a.stop,
			arango:         a,
			collectionType: collectionType,
			properties:     p,
		}
		switch collectionType {
		case bmp.PeerStateChangeMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.LSLinkMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.LSNodeMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.LSPrefixMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.LSSRv6SIDMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.L3VPNMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.L3VPNV4Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.L3VPNV6Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.UnicastPrefixMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.UnicastPrefixV4Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.UnicastPrefixV6Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.SRPolicyMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.SRPolicyV4Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.SRPolicyV6Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.FlowspecMsg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.FlowspecV4Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		case bmp.FlowspecV6Msg:
			a.collections[collectionType].handler = a.collections[collectionType].genericHandler
		default:
			return fmt.Errorf("unknown collection type %d", collectionType)
		}
	}
	var ci driver.Collection
	var err error
	// There are two possible collection types, base type and vertex type
	if !a.collections[collectionType].properties.isVertex {
		ci, err = a.db.Collection(context.TODO(), a.collections[collectionType].properties.name)
		if err != nil {
			if !driver.IsArangoErrorWithErrorNum(err, driver.ErrArangoDataSourceNotFound) {
				return err
			}
			ci, err = a.db.CreateCollection(context.TODO(), a.collections[collectionType].properties.name, a.collections[collectionType].properties.options)
			// log create collection
			glog.Infof("create collection: %s", a.collections[collectionType].properties.name)
		}
	} else {
		graph, err := a.ensureGraph(a.collections[collectionType].properties.name)
		// log ensure graph
		glog.Infof("ensure graph: %s", a.collections[collectionType].properties.name)
		if err != nil {
			return err
		}
		ci, err = graph.VertexCollection(context.TODO(), a.collections[collectionType].properties.name)
		// log ensure vertex
		glog.Infof("ensure graph.vertex: %s", a.collections[collectionType].properties.name)
		if err != nil {
			if !driver.IsArangoErrorWithErrorNum(err, driver.ErrArangoDataSourceNotFound) {
				return err
			}
			ci, err = graph.CreateVertexCollection(context.TODO(), a.collections[collectionType].properties.name)
		}
	}
	a.collections[collectionType].topicCollection = ci

	return nil
}

func (a *arangoDB) ensureGraph(name string) (driver.Graph, error) {
	var edgeDefinition driver.EdgeDefinition
	edgeDefinition.Collection = name + "_Edge"
	edgeDefinition.From = []string{name}
	edgeDefinition.To = []string{name}
	// log graph creation
	glog.Infof("graph created: %s", name)

	var options driver.CreateGraphOptions
	options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}
	graph, err := a.db.Graph(context.TODO(), name)
	if err == nil {
		graph.Remove(context.TODO())
		return a.db.CreateGraph(context.TODO(), name, &options)
	}
	if !driver.IsArangoErrorWithErrorNum(err, 1924) {
		return nil, err
	}

	return a.db.CreateGraph(context.TODO(), name, &options)
}

func (a *arangoDB) Start() error {
	glog.Infof("Connected to arango database, starting monitor")
	go a.monitor()
	for _, c := range a.collections {
		go c.handler()
	}
	return nil
}

func (a *arangoDB) Stop() error {
	close(a.stop)

	return nil
}

func (a *arangoDB) GetInterface() dbclient.DB {
	return a.DB
}

func (a *arangoDB) GetArangoDBInterface() *ArangoConn {
	return a.ArangoConn
}

func (a *arangoDB) StoreMessage(msgType dbclient.CollectionType, msg []byte) error {
	if t, ok := a.collections[msgType]; ok {
		t.queue <- &queueMsg{
			msgType: msgType,
			msgData: msg,
		}
	}

	return nil
}

func (a *arangoDB) monitor() {
	for {
		select {
		case <-a.stop:
			// TODO Add clean up of connection with Arango DB
			return
		}
	}
}
