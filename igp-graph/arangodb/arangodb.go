package arangodb

import (
	"context"
	"fmt"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/cisco-open/jalapeno/gobmp-arango/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/tools"
)

// Config holds the configuration for the IGP Graph processor
type Config struct {
	URL               string
	User              string
	Password          string
	Database          string
	LSPrefix          string
	LSLink            string
	LSSRv6SID         string
	LSNode            string
	IGPDomain         string
	IGPNode           string
	IGPv4Graph        string
	IGPv6Graph        string
	LSNodeEdge        string
	BatchSize         int
	ConcurrentWorkers int
	Notifier          kafkanotifier.Event
}

type arangoDB struct {
	dbclient.DB
	*ArangoConn
	config Config

	// Collections
	lsprefix   driver.Collection
	lslink     driver.Collection
	lssrv6sid  driver.Collection
	lsnode     driver.Collection
	igpDomain  driver.Collection
	igpNode    driver.Collection
	lsNodeEdge driver.Collection

	// Graphs
	igpv4Graph driver.Graph
	igpv6Graph driver.Graph

	// Performance components
	batchProcessor    *BatchProcessor
	updateCoordinator *UpdateCoordinator

	// Control
	stop     chan struct{}
	notifier kafkanotifier.Event
}

// NewDBSrvClient creates a new unified IGP Graph database client
func NewDBSrvClient(config Config) (dbclient.Srv, error) {
	if err := tools.URLAddrValidation(config.URL); err != nil {
		return nil, err
	}

	arangoConn, err := NewArango(ArangoConfig{
		URL:      config.URL,
		User:     config.User,
		Password: config.Password,
		Database: config.Database,
	})
	if err != nil {
		return nil, err
	}

	arango := &arangoDB{
		config:   config,
		stop:     make(chan struct{}),
		notifier: config.Notifier,
	}
	arango.DB = arango
	arango.ArangoConn = arangoConn

	// Initialize collections
	if err := arango.initializeCollections(); err != nil {
		return nil, fmt.Errorf("failed to initialize collections: %w", err)
	}

	// Initialize graphs
	if err := arango.initializeGraphs(); err != nil {
		return nil, fmt.Errorf("failed to initialize graphs: %w", err)
	}

	// Initialize performance components
	arango.batchProcessor = NewBatchProcessor(config.BatchSize, config.ConcurrentWorkers)
	arango.updateCoordinator = NewUpdateCoordinator(arango, config.BatchSize)

	glog.Infof("IGP Graph processor initialized with %d workers, batch size %d",
		config.ConcurrentWorkers, config.BatchSize)

	return arango, nil
}

func (a *arangoDB) initializeCollections() error {
	ctx := context.TODO()
	var err error

	// Check if base link state collections exist
	a.lsprefix, err = a.db.Collection(ctx, a.config.LSPrefix)
	if err != nil {
		return fmt.Errorf("ls_prefix collection not found: %w", err)
	}

	a.lslink, err = a.db.Collection(ctx, a.config.LSLink)
	if err != nil {
		return fmt.Errorf("ls_link collection not found: %w", err)
	}

	a.lssrv6sid, err = a.db.Collection(ctx, a.config.LSSRv6SID)
	if err != nil {
		return fmt.Errorf("ls_srv6_sid collection not found: %w", err)
	}

	a.lsnode, err = a.db.Collection(ctx, a.config.LSNode)
	if err != nil {
		return fmt.Errorf("ls_node collection not found: %w", err)
	}

	// Initialize or create IGP collections
	if err := a.ensureCollection(a.config.IGPDomain, false); err != nil {
		return err
	}
	a.igpDomain, err = a.db.Collection(ctx, a.config.IGPDomain)
	if err != nil {
		return err
	}

	if err := a.ensureCollection(a.config.IGPNode, false); err != nil {
		return err
	}
	a.igpNode, err = a.db.Collection(ctx, a.config.IGPNode)
	if err != nil {
		return err
	}

	// Create ls_node_edge collection for backward compatibility
	if err := a.ensureCollection(a.config.LSNodeEdge, true); err != nil {
		return err
	}
	a.lsNodeEdge, err = a.db.Collection(ctx, a.config.LSNodeEdge)
	if err != nil {
		return err
	}

	return nil
}

func (a *arangoDB) initializeGraphs() error {
	var err error

	// Initialize IGPv4 graph
	a.igpv4Graph, err = a.ensureGraph(a.config.IGPv4Graph, a.config.IGPNode)
	if err != nil {
		return fmt.Errorf("failed to initialize IGPv4 graph: %w", err)
	}

	// Initialize IGPv6 graph
	a.igpv6Graph, err = a.ensureGraph(a.config.IGPv6Graph, a.config.IGPNode)
	if err != nil {
		return fmt.Errorf("failed to initialize IGPv6 graph: %w", err)
	}

	return nil
}

func (a *arangoDB) ensureCollection(name string, isEdge bool) error {
	ctx := context.TODO()

	found, err := a.db.CollectionExists(ctx, name)
	if err != nil {
		return err
	}

	if !found {
		options := &driver.CreateCollectionOptions{}
		if isEdge {
			options.Type = driver.CollectionTypeEdge
		}

		glog.V(5).Infof("Creating collection: %s", name)
		_, err = a.db.CreateCollection(ctx, name, options)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *arangoDB) ensureGraph(graphName, vertexCollection string) (driver.Graph, error) {
	ctx := context.TODO()

	found, err := a.db.GraphExists(ctx, graphName)
	if err != nil {
		return nil, err
	}

	if found {
		graph, err := a.db.Graph(ctx, graphName)
		if err != nil {
			return nil, err
		}
		glog.V(5).Infof("Found existing graph: %s", graphName)
		return graph, nil
	}

	// Create edge collection for the graph
	edgeCollectionName := graphName
	if err := a.ensureCollection(edgeCollectionName, true); err != nil {
		return nil, err
	}

	// Create graph with edge definition
	var edgeDefinition driver.EdgeDefinition
	edgeDefinition.Collection = edgeCollectionName
	edgeDefinition.From = []string{vertexCollection}
	edgeDefinition.To = []string{vertexCollection}

	var options driver.CreateGraphOptions
	options.OrphanVertexCollections = []string{a.config.LSPrefix, a.config.LSSRv6SID}
	options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}

	graph, err := a.db.CreateGraphV2(ctx, graphName, &options)
	if err != nil {
		return nil, err
	}

	glog.V(5).Infof("Created new graph: %s", graphName)
	return graph, nil
}

func (a *arangoDB) Start() error {
	if err := a.loadInitialData(); err != nil {
		return fmt.Errorf("failed to load initial data: %w", err)
	}

	glog.Info("Starting IGP Graph processor components...")

	// Start batch processor
	if err := a.batchProcessor.Start(); err != nil {
		return fmt.Errorf("failed to start batch processor: %w", err)
	}

	// Start update coordinator
	if err := a.updateCoordinator.Start(); err != nil {
		return fmt.Errorf("failed to start update coordinator: %w", err)
	}

	glog.Info("IGP Graph processor started successfully")
	go a.monitor()

	return nil
}

func (a *arangoDB) Stop() error {
	glog.Info("Stopping IGP Graph processor...")

	close(a.stop)

	// Stop components
	if a.updateCoordinator != nil {
		a.updateCoordinator.Stop()
	}

	if a.batchProcessor != nil {
		a.batchProcessor.Stop()
	}

	glog.Info("IGP Graph processor stopped")
	return nil
}

func (a *arangoDB) GetInterface() dbclient.DB {
	return a.DB
}

func (a *arangoDB) GetArangoDBInterface() *ArangoConn {
	return a.ArangoConn
}

func (a *arangoDB) StoreMessage(msgType dbclient.CollectionType, msg []byte) error {
	return a.updateCoordinator.ProcessMessage(msgType, msg)
}

func (a *arangoDB) loadInitialData() error {
	glog.Info("Loading initial IGP topology data...")
	ctx := context.TODO()

	// Load initial nodes
	if err := a.loadInitialNodes(ctx); err != nil {
		return fmt.Errorf("failed to load initial nodes: %w", err)
	}

	// Load initial links
	if err := a.loadInitialLinks(ctx); err != nil {
		return fmt.Errorf("failed to load initial links: %w", err)
	}

	glog.Info("Initial IGP topology data loaded successfully")
	return nil
}

func (a *arangoDB) loadInitialNodes(ctx context.Context) error {
	// Query all ls_node documents and process them
	query := fmt.Sprintf("FOR doc IN %s RETURN doc", a.config.LSNode)
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()

	count := 0
	for {
		var node map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &node)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return err
		}

		// Process node (simplified for now)
		if err := a.processInitialNode(ctx, node); err != nil {
			glog.Warningf("Failed to process initial node %v: %v", node["_key"], err)
			continue
		}
		count++

		if count%1000 == 0 {
			glog.V(3).Infof("Loaded %d nodes...", count)
		}
	}

	glog.Infof("Loaded %d initial nodes", count)
	return nil
}

func (a *arangoDB) loadInitialLinks(ctx context.Context) error {
	// Query all ls_link documents and process them
	query := fmt.Sprintf("FOR doc IN %s RETURN doc", a.config.LSLink)
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()

	count := 0
	for {
		var link map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &link)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return err
		}

		// Process link (simplified for now)
		if err := a.processInitialLink(ctx, link); err != nil {
			glog.Warningf("Failed to process initial link %v: %v", link["_key"], err)
			continue
		}
		count++

		if count%1000 == 0 {
			glog.V(3).Infof("Loaded %d links...", count)
		}
	}

	glog.Infof("Loaded %d initial links", count)
	return nil
}

func (a *arangoDB) processInitialNode(ctx context.Context, node map[string]interface{}) error {
	// Convert map to LSNode-like structure for processing
	key, ok := node["_key"].(string)
	if !ok {
		return fmt.Errorf("invalid node key")
	}

	// Create IGP node entry
	igpNodeDoc := map[string]interface{}{
		"_key":          key,
		"igp_router_id": node["igp_router_id"],
		"protocol_id":   node["protocol_id"],
		"domain_id":     node["domain_id"],
		"area_id":       node["area_id"],
		"asn":           node["asn"],
		"name":          node["name"],
		"timestamp":     node["timestamp"],
	}

	// Try to create the document
	_, err := a.igpNode.CreateDocument(ctx, igpNodeDoc)
	if err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create igp_node document: %w", err)
		}
		// Document exists, update it
		if _, err := a.igpNode.UpdateDocument(ctx, key, igpNodeDoc); err != nil {
			return fmt.Errorf("failed to update igp_node document: %w", err)
		}
	}

	glog.V(9).Infof("Processed initial node: %s", key)
	return nil
}

func (a *arangoDB) processInitialLink(ctx context.Context, link map[string]interface{}) error {
	// Convert map to LSLink-like structure for processing
	key, ok := link["_key"].(string)
	if !ok {
		return fmt.Errorf("invalid link key")
	}

	// Create ls_node_edge entry for backward compatibility
	lsNodeEdgeDoc := map[string]interface{}{
		"_key":                  key,
		"_from":                 fmt.Sprintf("%s/%s", a.config.LSNode, link["igp_router_id"]),
		"_to":                   fmt.Sprintf("%s/%s", a.config.LSNode, link["remote_igp_router_id"]),
		"link":                  key,
		"protocol_id":           link["protocol_id"],
		"domain_id":             link["domain_id"],
		"area_id":               link["area_id"],
		"local_link_ip":         link["local_link_ip"],
		"remote_link_ip":        link["remote_link_ip"],
		"igp_metric":            link["igp_metric"],
		"local_node_asn":        link["local_node_asn"],
		"remote_node_asn":       link["remote_node_asn"],
		"max_link_bw":           link["max_link_bw"],
		"max_resv_bw":           link["max_resv_bw"],
		"te_default_metric":     link["te_default_metric"],
		"unidir_link_delay":     link["unidir_link_delay"],
		"unidir_packet_loss":    link["unidir_packet_loss"],
		"unidir_available_bw":   link["unidir_available_bw"],
		"unidir_bw_utilization": link["unidir_bw_utilization"],
	}

	// Create ls_node_edge document
	_, err := a.lsNodeEdge.CreateDocument(ctx, lsNodeEdgeDoc)
	if err != nil {
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create ls_node_edge document: %w", err)
		}
		// Document exists, update it
		if _, err := a.lsNodeEdge.UpdateDocument(ctx, key, lsNodeEdgeDoc); err != nil {
			return fmt.Errorf("failed to update ls_node_edge document: %w", err)
		}
	}

	// Create IGP graph edges
	if err := a.createIGPGraphEdges(ctx, link); err != nil {
		return fmt.Errorf("failed to create IGP graph edges: %w", err)
	}

	glog.V(9).Infof("Processed initial link: %s", key)
	return nil
}

// createIGPGraphEdges creates edges in appropriate IPv4 or IPv6 graphs based on MTID
func (a *arangoDB) createIGPGraphEdges(ctx context.Context, link map[string]interface{}) error {
	key, _ := link["_key"].(string)

	// Determine if this is IPv4 or IPv6 based on MTID
	// IPv4: no mt_id_tlv field (or mt_id = 0)
	// IPv6: mt_id_tlv.mt_id = 2
	isIPv6 := false

	if mtidTLV, exists := link["mt_id_tlv"]; exists {
		if mtidObj, ok := mtidTLV.(map[string]interface{}); ok {
			if mtID, ok := mtidObj["mt_id"].(float64); ok && mtID == 2 {
				isIPv6 = true
			}
		}
	}

	// Create IGP graph edge document
	igpEdgeDoc := map[string]interface{}{
		"_key":                  key,
		"_from":                 fmt.Sprintf("%s/%s", a.config.IGPNode, link["igp_router_id"]),
		"_to":                   fmt.Sprintf("%s/%s", a.config.IGPNode, link["remote_igp_router_id"]),
		"link":                  key,
		"protocol_id":           link["protocol_id"],
		"domain_id":             link["domain_id"],
		"area_id":               link["area_id"],
		"local_link_ip":         link["local_link_ip"],
		"remote_link_ip":        link["remote_link_ip"],
		"igp_metric":            link["igp_metric"],
		"local_node_asn":        link["local_node_asn"],
		"remote_node_asn":       link["remote_node_asn"],
		"max_link_bw":           link["max_link_bw"],
		"max_resv_bw":           link["max_resv_bw"],
		"te_default_metric":     link["te_default_metric"],
		"unidir_link_delay":     link["unidir_link_delay"],
		"unidir_packet_loss":    link["unidir_packet_loss"],
		"unidir_available_bw":   link["unidir_available_bw"],
		"unidir_bw_utilization": link["unidir_bw_utilization"],
	}

	// Create edge in appropriate graph based on IP version
	if isIPv6 {
		// IPv6 graph (MTID = 2)
		igpv6EdgeCollection, err := a.db.Collection(ctx, a.config.IGPv6Graph)
		if err != nil {
			return fmt.Errorf("failed to get IGPv6 edge collection: %w", err)
		}

		_, err = igpv6EdgeCollection.CreateDocument(ctx, igpEdgeDoc)
		if err != nil && !driver.IsConflict(err) {
			return fmt.Errorf("failed to create IGPv6 edge: %w", err)
		}
		glog.V(6).Infof("Created IPv6 graph edge: %s -> %s", link["igp_router_id"], link["remote_igp_router_id"])
	} else {
		// IPv4 graph (MTID = nil or 0)
		igpv4EdgeCollection, err := a.db.Collection(ctx, a.config.IGPv4Graph)
		if err != nil {
			return fmt.Errorf("failed to get IGPv4 edge collection: %w", err)
		}

		_, err = igpv4EdgeCollection.CreateDocument(ctx, igpEdgeDoc)
		if err != nil && !driver.IsConflict(err) {
			return fmt.Errorf("failed to create IGPv4 edge: %w", err)
		}
		glog.V(6).Infof("Created IPv4 graph edge: %s -> %s", link["igp_router_id"], link["remote_igp_router_id"])
	}

	return nil
}

func (a *arangoDB) monitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stop:
			return
		case <-ticker.C:
			// Log performance statistics
			if a.batchProcessor != nil {
				stats := a.batchProcessor.GetStats()
				processed := stats.Processed
				pending := stats.Pending
				glog.V(5).Infof("Batch processor stats: processed=%d, pending=%d",
					processed, pending)
			}
		}
	}
}
