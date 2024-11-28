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
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/tools"
)

type arangoDB struct {
	dbclient.DB
	*ArangoConn
	stop      chan struct{}
	lsprefix  driver.Collection
	lslink    driver.Collection
	lssrv6sid driver.Collection
	lsnode    driver.Collection
	lsnodeExt driver.Collection
	igpDomain driver.Collection
	graphv4   driver.Collection
	graphv6   driver.Collection
	lsv4Graph driver.Graph
	lsv6Graph driver.Graph
	notifier  kafkanotifier.Event
}

// NewDBSrvClient returns an instance of a DB server client process
func NewDBSrvClient(arangoSrv, user, pass, dbname, lsprefix, lslink, lssrv6sid, lsnode,
	lsnodeExt string, igpDomain string, lsv4Graph string, lsv6Graph string, notifier kafkanotifier.Event) (dbclient.Srv, error) {
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

	// check for ls_node_extended collection
	found, err := arango.db.CollectionExists(context.TODO(), lsnodeExt)
	if err != nil {
		return nil, err
	}
	if found {
		c, err := arango.db.Collection(context.TODO(), lsnodeExt)
		if err != nil {
			return nil, err
		}
		if err := c.Remove(context.TODO()); err != nil {
			return nil, err
		}
	}
	// create ls_node_extended collection
	var lsnodeExt_options = &driver.CreateCollectionOptions{ /* ... */ }
	glog.V(5).Infof("ls_node_extended not found, creating")
	arango.lsnodeExt, err = arango.db.CreateCollection(context.TODO(), "ls_node_extended", lsnodeExt_options)
	if err != nil {
		return nil, err
	}
	// check if collection exists, if not fail as processor has failed to create collection
	arango.lsnodeExt, err = arango.db.Collection(context.TODO(), lsnodeExt)
	if err != nil {
		return nil, err
	}

	// check for igp_domain collection
	found, err = arango.db.CollectionExists(context.TODO(), igpDomain)
	if err != nil {
		return nil, err
	}
	if found {
		c, err := arango.db.Collection(context.TODO(), igpDomain)
		if err != nil {
			return nil, err
		}
		if err := c.Remove(context.TODO()); err != nil {
			return nil, err
		}
	}
	// create igp_domain collection
	var igpdomain_options = &driver.CreateCollectionOptions{ /* ... */ }
	glog.V(5).Infof("igp_domain collection not found, creating")
	arango.igpDomain, err = arango.db.CreateCollection(context.TODO(), "igp_domain", igpdomain_options)
	if err != nil {
		return nil, err
	}
	// check if collection exists, if not fail as processor has failed to create collection
	arango.igpDomain, err = arango.db.Collection(context.TODO(), igpDomain)
	if err != nil {
		return nil, err
	}

	// check for lsv4 topology graph
	found, err = arango.db.GraphExists(context.TODO(), lsv4Graph)
	if err != nil {
		return nil, err
	}
	if found {
		c, err := arango.db.Graph(context.TODO(), lsv4Graph)
		if err != nil {
			return nil, err
		}
		glog.Infof("found graph %s", c)

	} else {
		// create graph
		var edgeDefinition driver.EdgeDefinition
		edgeDefinition.Collection = "lsv4_graph"
		edgeDefinition.From = []string{"ls_node_extended"}
		edgeDefinition.To = []string{"ls_node_extended"}
		var options driver.CreateGraphOptions
		options.OrphanVertexCollections = []string{"ls_srv6_sid", "ls_prefix"}
		options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}

		arango.lsv4Graph, err = arango.db.CreateGraph(context.TODO(), lsv4Graph, &options)
		if err != nil {
			return nil, err
		}
	}

	// check if lsv4_graph exists, if not fail as processor has failed to create graph
	arango.graphv4, err = arango.db.Collection(context.TODO(), "lsv4_graph")
	if err != nil {
		return nil, err
	}

	// check for lsv6 topology graph
	found, err = arango.db.GraphExists(context.TODO(), lsv6Graph)
	if err != nil {
		return nil, err
	}
	if found {
		c, err := arango.db.Graph(context.TODO(), lsv6Graph)
		if err != nil {
			return nil, err
		}
		glog.Infof("found graph %s", c)

	} else {
		// create graph
		var edgeDefinition driver.EdgeDefinition
		edgeDefinition.Collection = "lsv6_graph"
		edgeDefinition.From = []string{"ls_node_extended"}
		edgeDefinition.To = []string{"ls_node_extended"}
		var options driver.CreateGraphOptions
		options.OrphanVertexCollections = []string{"ls_srv6_sid", "ls_prefix"}
		options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}

		arango.lsv6Graph, err = arango.db.CreateGraph(context.TODO(), lsv6Graph, &options)
		if err != nil {
			return nil, err
		}
	}

	// check if lsv6_graph exists, if not fail as processor has failed to create graph
	arango.graphv6, err = arango.db.Collection(context.TODO(), "lsv6_graph")
	if err != nil {
		return nil, err
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

// loadCollections calls a series of subfunctions to perform ArangoDB operations including populating
// ls_node_extended collection and querying other link state collections to build the lsv4_graph and lsv6_graph

func (a *arangoDB) loadCollections() error {
	ctx := context.TODO()

	if err := a.lsExtendedNodes(ctx); err != nil {
		return err
	}
	if err := a.processDuplicateNodes(ctx); err != nil {
		return err
	}
	if err := a.loadPrefixSIDs(ctx); err != nil {
		return err
	}
	if err := a.loadSRv6SIDs(ctx); err != nil {
		return err
	}
	if err := a.processIBGPv6Peering(ctx); err != nil {
		return err
	}
	if err := a.createIGPDomains(ctx); err != nil {
		return err
	}
	if err := a.lsv4LinkEdges(ctx); err != nil {
		return err
	}
	if err := a.lsv4PrefixEdges(ctx); err != nil {
		return err
	}
	if err := a.lsv6LinkEdges(ctx); err != nil {
		return err
	}
	if err := a.lsv6PrefixEdges(ctx); err != nil {
		return err
	}

	return nil
}
func (a *arangoDB) lsExtendedNodes(ctx context.Context) error {
	lsn_query := "for l in " + a.lsnode.Name() + " insert l in " + a.lsnodeExt.Name() + ""
	cursor, err := a.db.Query(ctx, lsn_query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	return nil
}

func (a *arangoDB) processDuplicateNodes(ctx context.Context) error {
	// BGP-LS generates both a level-1 and a level-2 entry for level-1-2 nodes
	// Here we remove duplicate entries in the ls_node_extended collection
	dup_query := "LET duplicates = ( FOR d IN " + a.lsnodeExt.Name() +
		" COLLECT id = d.igp_router_id, domain = d.domain_id, area = d.area_id WITH COUNT INTO count " +
		" FILTER count > 1 RETURN { id: id, domain: domain, area: area, count: count }) " +
		"FOR d IN duplicates FOR m IN ls_node_extended " +
		"FILTER d.id == m.igp_router_id filter d.domain == m.domain_id RETURN m "
	pcursor, err := a.db.Query(ctx, dup_query, nil)
	if err != nil {
		return err
	}
	defer pcursor.Close()
	for {
		var doc duplicateNode
		dupe, err := pcursor.ReadDocument(ctx, &doc)

		if err != nil {
			if !driver.IsNoMoreDocuments(err) {
				return err
			}
			break
		}
		glog.Infof("Got doc with key '%s' from query\n", dupe.Key)

		if doc.ProtocolID == 1 {
			glog.Infof("remove level-1 duplicate node: %s + igp id: %s protocol id: %v +  ", doc.Key, doc.IGPRouterID, doc.ProtocolID)
			if _, err := a.lsnodeExt.RemoveDocument(ctx, doc.Key); err != nil {
				if !driver.IsConflict(err) {
					return err
				}
			}
		}
		if doc.ProtocolID == 2 {
			update_query := "for l in " + a.lsnodeExt.Name() + " filter l._key == " + "\"" + doc.Key + "\"" +
				" UPDATE l with { protocol: " + "\"" + "ISIS Level 1-2" + "\"" + " } in " + a.lsnodeExt.Name() + ""
			cursor, err := a.db.Query(ctx, update_query, nil)
			glog.Infof("update query: %s ", update_query)
			if err != nil {
				return err
			}
			defer cursor.Close()
		}
	}

	return nil
}

func (a *arangoDB) loadPrefixSIDs(ctx context.Context) error {
	// Find and add sr-mpls prefix sids to nodes in the ls_node_extended collection
	sr_query := "for p in  " + a.lsprefix.Name() +
		" filter p.mt_id_tlv.mt_id != 2 && p.prefix_attr_tlvs.ls_prefix_sid != null return p "
	cursor, err := a.db.Query(ctx, sr_query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var p message.LSPrefix
		meta, err := cursor.ReadDocument(ctx, &p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		if err := a.processPrefixSID(ctx, meta.Key, meta.ID.String(), p); err != nil {
			glog.Errorf("Failed to process ls_prefix_sid %s with error: %+v", p.ID, err)
		}
	}

	return nil
}

func (a *arangoDB) loadSRv6SIDs(ctx context.Context) error {
	// Find and add srv6 sids to nodes in the ls_node_extended collection
	srv6_query := "for s in  " + a.lssrv6sid.Name() + " return s "
	cursor, err := a.db.Query(ctx, srv6_query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var p message.LSSRv6SID
		meta, err := cursor.ReadDocument(ctx, &p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		if err := a.processLSSRv6SID(ctx, meta.Key, meta.ID.String(), &p); err != nil {
			glog.Errorf("Failed to process ls_srv6_sid %s with error: %+v", p.ID, err)
		}
	}

	return nil
}

func (a *arangoDB) processIBGPv6Peering(ctx context.Context) error {
	// add ipv6 iBGP peering address and ipv4 bgp router-id
	ibgp6_query := "for s in peer filter s.remote_ip like " + "\"%:%\"" + " return s "
	cursor, err := a.db.Query(ctx, ibgp6_query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var p message.PeerStateChange
		meta, err := cursor.ReadDocument(ctx, &p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		if err := a.processbgp6(ctx, meta.Key, meta.ID.String(), &p); err != nil {
			glog.Errorf("Failed to process ibgp peering %s with error: %+v", p.ID, err)
		}
	}

	return nil
}

func (a *arangoDB) createIGPDomains(ctx context.Context) error {
	// create igp_domain collection - useful in scaled multi-domain environments
	igpdomain_query := "for l in ls_node_extended insert " +
		"{ _key: CONCAT_SEPARATOR(" + "\"_\", l.protocol_id, l.domain_id, l.asn), " +
		"asn: l.asn, protocol_id: l.protocol_id, domain_id: l.domain_id, protocol: l.protocol } " +
		"into igp_domain OPTIONS { ignoreErrors: true } return l"
	cursor, err := a.db.Query(ctx, igpdomain_query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()

	return nil
}

func (a *arangoDB) lsv4LinkEdges(ctx context.Context) error {
	// Find ipv4 ls_link entries to create edges in the lsv4_graph
	lsv4linkquery := "for l in " + a.lslink.Name() + " filter l.protocol_id != 7 RETURN l"
	cursor, err := a.db.Query(ctx, lsv4linkquery, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var p message.LSLink
		meta, err := cursor.ReadDocument(ctx, &p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		if err := a.processLSLinkEdge(ctx, meta.Key, &p); err != nil {
			glog.Errorf("failed to process key: %s with error: %+v", meta.Key, err)
			continue
		}
	}

	return nil
}

func (a *arangoDB) lsv4PrefixEdges(ctx context.Context) error {
	// Find ls_prefix entries to create prefix or subnet edges in the lsv4_graph
	lsv4pfxquery := "for l in " + a.lsprefix.Name() + //" filter l.mt_id_tlv == null return l"
		" filter l.mt_id_tlv.mt_id != 2 && l.prefix_len != 30 && " +
		"l.prefix_len != 31 && l.prefix_len != 32 return l"
	cursor, err := a.db.Query(ctx, lsv4pfxquery, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var p message.LSPrefix
		meta, err := cursor.ReadDocument(ctx, &p)
		//glog.Infof("processing lsprefix document: %+v", p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		if err := a.processLSPrefixEdge(ctx, meta.Key, &p); err != nil {
			glog.Errorf("failed to process key: %s with error: %+v", meta.Key, err)
			continue
		}
	}

	return nil
}

func (a *arangoDB) lsv6LinkEdges(ctx context.Context) error {
	// Find ipv6 ls_link entries to create edges in the lsv6_graph
	lsv6linkquery := "for l in " + a.lslink.Name() + " filter l.protocol_id != 7 RETURN l"
	cursor, err := a.db.Query(ctx, lsv6linkquery, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var p message.LSLink
		meta, err := cursor.ReadDocument(ctx, &p)
		//glog.Infof("processing lslink document: %+v", p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		if err := a.processLSv6LinkEdge(ctx, meta.Key, &p); err != nil {
			glog.Errorf("failed to process key: %s with error: %+v", meta.Key, err)
			continue
		}
	}

	return nil
}

func (a *arangoDB) lsv6PrefixEdges(ctx context.Context) error {
	// Find ipv6 ls_prefix entries to create prefix or subnet edges in the lsv6_graph
	lsv6pfxquery := "for l in " + a.lsprefix.Name() +
		" filter l.mt_id_tlv.mt_id == 2 && l.prefix_len != 126 && " +
		"l.prefix_len != 127 && l.prefix_len != 128 return l"
	cursor, err := a.db.Query(ctx, lsv6pfxquery, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for {
		var p message.LSPrefix
		meta, err := cursor.ReadDocument(ctx, &p)
		//glog.Infof("processing lsprefix document: %+v", p)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		if err := a.processLSv6PrefixEdge(ctx, meta.Key, &p); err != nil {
			glog.Errorf("failed to process key: %s with error: %+v", meta.Key, err)
			continue
		}
	}

	return nil

}
