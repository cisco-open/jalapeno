package arangodb

import (
	"context"
	"encoding/json"

	driver "github.com/arangodb/go-driver"
	"github.com/cisco-open/jalapeno/topology/dbclient"
	"github.com/cisco-open/jalapeno/topology/kafkanotifier"
	notifier "github.com/cisco-open/jalapeno/topology/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/tools"
)

type arangoDB struct {
	dbclient.DB
	*ArangoConn
	stop       chan struct{}
	lsprefix   driver.Collection
	lslink     driver.Collection
	lssrv6sid  driver.Collection
	lsnode     driver.Collection
	igpDomain  driver.Collection
	igpNode    driver.Collection
	graphv4    driver.Collection
	graphv6    driver.Collection
	igpv4Graph driver.Graph
	igpv6Graph driver.Graph
	notifier   kafkanotifier.Event
}

// NewDBSrvClient returns an instance of a DB server client process
func NewDBSrvClient(arangoSrv, user, pass, dbname, lsprefix, lslink, lssrv6sid, lsnode,
	igpDomain string, igpNode string, igpv4Graph string, igpv6Graph string,
	notifier kafkanotifier.Event) (dbclient.Srv, error) {
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
		stop: make(chan struct{}),
	}
	arango.DB = arango
	arango.ArangoConn = arangoConn

	// Check if base link state collections exist, if not fail as Jalapeno topology is not running
	arango.lsprefix, err = arango.db.Collection(context.TODO(), lsprefix)
	if err != nil {
		return nil, err
	}
	arango.lslink, err = arango.db.Collection(context.TODO(), lslink)
	if err != nil {
		return nil, err
	}
	arango.lssrv6sid, err = arango.db.Collection(context.TODO(), lssrv6sid)
	if err != nil {
		return nil, err
	}
	arango.lsnode, err = arango.db.Collection(context.TODO(), lsnode)
	if err != nil {
		return nil, err
	}

	// check for igp_node collection
	found, err := arango.db.CollectionExists(context.TODO(), igpNode)
	if err != nil {
		return nil, err
	}
	if found {
		// Collection exists, get a reference to it
		arango.igpNode, err = arango.db.Collection(context.TODO(), igpNode)
		if err != nil {
			return nil, err
		}
	} else {
		// Collection doesn't exist, create it
		var igpNode_options = &driver.CreateCollectionOptions{ /* ... */ }
		glog.V(5).Infof("igp_node not found, creating")
		arango.igpNode, err = arango.db.CreateCollection(context.TODO(), "igp_node", igpNode_options)
		if err != nil {
			return nil, err
		}
	}

	// check for igp_domain collection
	found, err = arango.db.CollectionExists(context.TODO(), igpDomain)
	if err != nil {
		return nil, err
	}
	if found {
		// Collection exists, get a reference to it
		arango.igpDomain, err = arango.db.Collection(context.TODO(), igpDomain)
		if err != nil {
			return nil, err
		}
	} else {
		// Collection doesn't exist, create it
		var igpDomain_options = &driver.CreateCollectionOptions{ /* ... */ }
		glog.V(5).Infof("igp_domain not found, creating")
		arango.igpDomain, err = arango.db.CreateCollection(context.TODO(), "igp_domain", igpDomain_options)
		if err != nil {
			return nil, err
		}
	}

	// Handle IGPv4 graph and edge collection
	found, err = arango.db.GraphExists(context.TODO(), igpv4Graph)
	if err != nil {
		return nil, err
	}
	if found {
		// Get reference to existing graph
		arango.igpv4Graph, err = arango.db.Graph(context.TODO(), igpv4Graph)
		if err != nil {
			return nil, err
		}
		// Get reference to existing edge collection
		arango.graphv4, err = arango.db.Collection(context.TODO(), "igpv4_graph")
		if err != nil {
			return nil, err
		}
		glog.V(5).Infof("Found existing graph and edge collection: %s", igpv4Graph)
	} else {
		// Create edge collection first
		var edgeOptions = &driver.CreateCollectionOptions{Type: driver.CollectionTypeEdge}
		arango.graphv4, err = arango.db.CreateCollection(context.TODO(), "igpv4_graph", edgeOptions)
		if err != nil {
			return nil, err
		}

		// Create graph with edge definition
		var edgeDefinition driver.EdgeDefinition
		edgeDefinition.Collection = "igpv4_graph"
		edgeDefinition.From = []string{"igp_node"}
		edgeDefinition.To = []string{"igp_node"}

		var options driver.CreateGraphOptions
		options.OrphanVertexCollections = []string{"ls_srv6_sid", "ls_prefix"}
		options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}

		arango.igpv4Graph, err = arango.db.CreateGraphV2(context.TODO(), igpv4Graph, &options)
		if err != nil {
			return nil, err
		}
		glog.V(5).Infof("Created new graph and edge collection: %s", igpv4Graph)
	}

	// Handle IGPv6 graph and edge collection
	found, err = arango.db.GraphExists(context.TODO(), igpv6Graph)
	if err != nil {
		return nil, err
	}
	if found {
		// Get reference to existing graph
		arango.igpv6Graph, err = arango.db.Graph(context.TODO(), igpv6Graph)
		if err != nil {
			return nil, err
		}
		// Get reference to existing edge collection
		arango.graphv6, err = arango.db.Collection(context.TODO(), "igpv6_graph")
		if err != nil {
			return nil, err
		}
		glog.V(5).Infof("Found existing graph and edge collection: %s", igpv6Graph)
	} else {
		// Create edge collection first
		var edgeOptions = &driver.CreateCollectionOptions{Type: driver.CollectionTypeEdge}
		arango.graphv6, err = arango.db.CreateCollection(context.TODO(), "igpv6_graph", edgeOptions)
		if err != nil {
			return nil, err
		}

		// Create graph with edge definition
		var edgeDefinition driver.EdgeDefinition
		edgeDefinition.Collection = "igpv6_graph"
		edgeDefinition.From = []string{"igp_node"}
		edgeDefinition.To = []string{"igp_node"}

		var options driver.CreateGraphOptions
		options.OrphanVertexCollections = []string{"ls_srv6_sid", "ls_prefix"}
		options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}

		arango.igpv6Graph, err = arango.db.CreateGraphV2(context.TODO(), igpv6Graph, &options)
		if err != nil {
			return nil, err
		}
		glog.V(5).Infof("Created new graph and edge collection: %s", igpv6Graph)
	}

	return arango, nil
}

func (a *arangoDB) Start() error {
	if err := a.loadCollections(); err != nil {
		return err
	}
	glog.Infof("Connected to arango database, starting monitor")
	go a.monitor()

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
	event := &notifier.EventMessage{}
	if err := json.Unmarshal(msg, event); err != nil {
		return err
	}
	event.TopicType = msgType
	switch msgType {
	case bmp.LSSRv6SIDMsg:
		return a.lsSRv6SIDHandler(event)
	case bmp.LSNodeMsg:
		return a.lsNodeHandler(event)
	case bmp.LSPrefixMsg:
		return a.lsPrefixHandler(event)
	case bmp.LSLinkMsg:
		return a.lsLinkHandler(event)
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
