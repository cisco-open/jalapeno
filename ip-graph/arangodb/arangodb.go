package arangodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/cisco-open/jalapeno/gobmp-arango/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/tools"
)

type arangoDB struct {
	dbclient.DB
	*ArangoConn
	config Config

	// Collections
	igpv4Graph  driver.Collection
	igpv6Graph  driver.Collection
	igpNode     driver.Collection
	igpDomain   driver.Collection
	ipv4Graph   driver.Collection
	ipv6Graph   driver.Collection
	bgpNode     driver.Collection
	bgpPrefixV4 driver.Collection
	bgpPrefixV6 driver.Collection

	// Graphs
	ipv4GraphDB driver.Graph
	ipv6GraphDB driver.Graph

	// Processing components
	batchProcessor    *BatchProcessor
	updateCoordinator *UpdateCoordinator

	// Control channels
	stop    chan struct{}
	started bool
	mu      sync.RWMutex

	// Event notifier
	notifier kafkanotifier.Event
}

// NewDBSrvClient returns an instance of a DB server client process
func NewDBSrvClient(config Config, notifier kafkanotifier.Event) (dbclient.Srv, error) {
	if err := tools.URLAddrValidation(config.DatabaseServer); err != nil {
		return nil, err
	}

	arangoConn, err := NewArango(ArangoConfig{
		URL:      config.DatabaseServer,
		User:     config.User,
		Password: config.Password,
		Database: config.Database,
	})
	if err != nil {
		return nil, err
	}

	arango := &arangoDB{
		ArangoConn: arangoConn,
		config:     config,
		stop:       make(chan struct{}),
		notifier:   notifier,
	}
	arango.DB = arango

	// Initialize collections and graphs
	if err := arango.initializeCollections(); err != nil {
		return nil, fmt.Errorf("failed to initialize collections: %w", err)
	}

	// Initialize batch processor
	arango.batchProcessor = NewBatchProcessor(arango, config.BatchSize, config.ConcurrentWorkers)

	// Initialize update coordinator
	arango.updateCoordinator = NewUpdateCoordinator(arango)

	glog.Infof("IP Graph processor initialized with %d workers, batch size %d",
		config.ConcurrentWorkers, config.BatchSize)

	return arango, nil
}

func (a *arangoDB) initializeCollections() error {
	ctx := context.TODO()

	// Initialize IGP source collections (read-only access)
	var err error
	a.igpv4Graph, err = a.Collection(ctx, a.config.IGPv4Graph)
	if err != nil {
		return fmt.Errorf("failed to access IGP v4 graph collection %s: %w", a.config.IGPv4Graph, err)
	}

	a.igpv6Graph, err = a.Collection(ctx, a.config.IGPv6Graph)
	if err != nil {
		return fmt.Errorf("failed to access IGP v6 graph collection %s: %w", a.config.IGPv6Graph, err)
	}

	a.igpNode, err = a.Collection(ctx, a.config.IGPNode)
	if err != nil {
		return fmt.Errorf("failed to access IGP node collection %s: %w", a.config.IGPNode, err)
	}

	a.igpDomain, err = a.Collection(ctx, a.config.IGPDomain)
	if err != nil {
		return fmt.Errorf("failed to access IGP domain collection %s: %w", a.config.IGPDomain, err)
	}

	// Initialize IP graph collections (full topology)
	a.ipv4Graph, err = a.EnsureCollection(ctx, a.config.IPv4Graph, true) // edge collection
	if err != nil {
		return fmt.Errorf("failed to create IPv4 graph collection: %w", err)
	}

	a.ipv6Graph, err = a.EnsureCollection(ctx, a.config.IPv6Graph, true) // edge collection
	if err != nil {
		return fmt.Errorf("failed to create IPv6 graph collection: %w", err)
	}

	// Initialize BGP collections
	a.bgpNode, err = a.EnsureCollection(ctx, a.config.BGPNode, false) // document collection
	if err != nil {
		return fmt.Errorf("failed to create BGP node collection: %w", err)
	}

	a.bgpPrefixV4, err = a.EnsureCollection(ctx, a.config.BGPPrefixV4, false) // document collection
	if err != nil {
		return fmt.Errorf("failed to create BGP prefix v4 collection: %w", err)
	}

	a.bgpPrefixV6, err = a.EnsureCollection(ctx, a.config.BGPPrefixV6, false) // document collection
	if err != nil {
		return fmt.Errorf("failed to create BGP prefix v6 collection: %w", err)
	}

	// Create IP topology graphs
	if err := a.ensureIPGraphs(ctx); err != nil {
		return fmt.Errorf("failed to ensure IP graphs: %w", err)
	}

	return nil
}

func (a *arangoDB) ensureIPGraphs(ctx context.Context) error {
	// Create IPv4 full topology graph
	ipv4GraphOptions := driver.CreateGraphOptions{
		EdgeDefinitions: []driver.EdgeDefinition{
			{
				Collection: a.config.IPv4Graph,
				From:       []string{a.config.IGPNode, a.config.BGPNode, a.config.BGPPrefixV4},
				To:         []string{a.config.IGPNode, a.config.BGPNode, a.config.BGPPrefixV4},
			},
		},
	}

	ipv4Graph, err := a.CreateGraph(ctx, a.config.IPv4Graph, ipv4GraphOptions)
	if err != nil {
		return fmt.Errorf("failed to create IPv4 graph: %w", err)
	}
	a.ipv4GraphDB = ipv4Graph
	glog.V(6).Infof("Ensured IPv4 full topology graph: %s", a.config.IPv4Graph)

	// Create IPv6 full topology graph
	ipv6GraphOptions := driver.CreateGraphOptions{
		EdgeDefinitions: []driver.EdgeDefinition{
			{
				Collection: a.config.IPv6Graph,
				From:       []string{a.config.IGPNode, a.config.BGPNode, a.config.BGPPrefixV6},
				To:         []string{a.config.IGPNode, a.config.BGPNode, a.config.BGPPrefixV6},
			},
		},
	}

	ipv6Graph, err := a.CreateGraph(ctx, a.config.IPv6Graph, ipv6GraphOptions)
	if err != nil {
		return fmt.Errorf("failed to create IPv6 graph: %w", err)
	}
	a.ipv6GraphDB = ipv6Graph
	glog.V(6).Infof("Ensured IPv6 full topology graph: %s", a.config.IPv6Graph)

	return nil
}

func (a *arangoDB) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.started {
		return fmt.Errorf("processor already started")
	}

	// Load initial data
	if err := a.loadInitialData(); err != nil {
		return fmt.Errorf("failed to load initial data: %w", err)
	}

	glog.Info("Starting IP Graph processor components...")

	// Start batch processor
	if err := a.batchProcessor.Start(); err != nil {
		return fmt.Errorf("failed to start batch processor: %w", err)
	}

	// Start update coordinator
	if err := a.updateCoordinator.Start(); err != nil {
		return fmt.Errorf("failed to start update coordinator: %w", err)
	}

	// Start monitoring goroutine
	go a.monitor()

	a.started = true
	glog.Info("IP Graph processor started successfully")

	return nil
}

func (a *arangoDB) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.started {
		return nil
	}

	glog.Info("Stopping IP Graph processor...")

	// Stop components
	close(a.stop)

	if a.updateCoordinator != nil {
		a.updateCoordinator.Stop()
	}

	if a.batchProcessor != nil {
		a.batchProcessor.Stop()
	}

	a.started = false
	glog.Info("IP Graph processor stopped")

	return nil
}

func (a *arangoDB) GetInterface() dbclient.DB {
	return a.DB
}

func (a *arangoDB) StoreMessage(msgType dbclient.CollectionType, msg []byte) error {
	if a.updateCoordinator == nil {
		return ErrProcessorNotStarted
	}
	return a.updateCoordinator.ProcessMessage(msgType, msg)
}

func (a *arangoDB) loadInitialData() error {
	glog.Info("Loading initial IP topology data...")
	ctx := context.TODO()

	// Step 1: Copy IGP graph data to IP graphs
	if err := a.copyIGPGraphData(ctx); err != nil {
		return fmt.Errorf("failed to copy IGP graph data: %w", err)
	}

	// Step 2: Load existing BGP data
	if err := a.loadInitialBGPData(ctx); err != nil {
		return fmt.Errorf("failed to load initial BGP data: %w", err)
	}

	glog.Info("Initial IP topology data loaded successfully")
	return nil
}

func (a *arangoDB) copyIGPGraphData(ctx context.Context) error {
	glog.V(6).Info("Copying IGP graph data to IP graphs...")

	// Check if IGP collections exist
	igpDataAvailable := a.checkIGPDataAvailability(ctx)
	if !igpDataAvailable {
		glog.Warning("No IGP data found - ip-graph will populate graphs with BGP data only")
		glog.Info("Note: For full topology, ensure igp-graph processor is running and has processed network data")
		return nil // Continue without IGP data
	}

	// Use the dedicated IGP copy processor for enhanced copying logic
	igpCopyProcessor := NewIGPCopyProcessor(a)
	if err := igpCopyProcessor.CopyIGPGraphsToFullTopology(ctx); err != nil {
		glog.Warningf("Failed to copy IGP graphs (continuing with BGP-only): %v", err)
		return nil // Continue without IGP data rather than failing
	}

	glog.Info("IGP graph data copied to IP graphs successfully")
	return nil
}

func (a *arangoDB) checkIGPDataAvailability(ctx context.Context) bool {
	// Check if igp_node collection exists and has data
	igpNodeExists, err := a.db.CollectionExists(ctx, a.config.IGPNode)
	if err != nil || !igpNodeExists {
		glog.V(6).Infof("IGP node collection '%s' not found", a.config.IGPNode)
		return false
	}

	// Check if igpv4_graph collection exists and has data
	igpv4GraphExists, err := a.db.CollectionExists(ctx, a.config.IGPv4Graph)
	if err != nil || !igpv4GraphExists {
		glog.V(6).Infof("IGPv4 graph collection '%s' not found", a.config.IGPv4Graph)
		return false
	}

	// Check if igpv6_graph collection exists and has data
	igpv6GraphExists, err := a.db.CollectionExists(ctx, a.config.IGPv6Graph)
	if err != nil || !igpv6GraphExists {
		glog.V(6).Infof("IGPv6 graph collection '%s' not found", a.config.IGPv6Graph)
		return false
	}

	// Check if collections have actual data
	nodeCount, err := a.getCollectionCount(ctx, a.config.IGPNode)
	if err != nil || nodeCount == 0 {
		glog.V(6).Infof("IGP node collection '%s' is empty", a.config.IGPNode)
		return false
	}

	glog.V(6).Infof("IGP data available: %d nodes found", nodeCount)
	return true
}

func (a *arangoDB) getCollectionCount(ctx context.Context, collectionName string) (int64, error) {
	query := fmt.Sprintf("RETURN LENGTH(%s)", collectionName)
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return 0, err
	}
	defer cursor.Close()

	var count int64
	if cursor.HasMore() {
		if _, err := cursor.ReadDocument(ctx, &count); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (a *arangoDB) copyGraphCollection(ctx context.Context, source, target driver.Collection) error {
	// Query all documents from source collection
	query := fmt.Sprintf("FOR doc IN %s RETURN doc", source.Name())
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query source collection %s: %w", source.Name(), err)
	}
	defer cursor.Close()

	count := 0
	for {
		var doc map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			return fmt.Errorf("error reading document: %w", err)
		}

		// Insert into target collection
		if _, err := target.CreateDocument(ctx, doc); err != nil {
			if !driver.IsConflict(err) {
				return fmt.Errorf("failed to create document in target collection: %w", err)
			}
			// Document already exists, update it
			if key, ok := doc["_key"].(string); ok {
				if _, err := target.UpdateDocument(ctx, key, doc); err != nil {
					return fmt.Errorf("failed to update document in target collection: %w", err)
				}
			}
		}

		count++
		if count%1000 == 0 {
			glog.V(3).Infof("Copied %d documents from %s to %s...", count, source.Name(), target.Name())
		}
	}

	glog.Infof("Copied %d documents from %s to %s", count, source.Name(), target.Name())
	return nil
}

func (a *arangoDB) loadInitialBGPData(ctx context.Context) error {
	glog.Info("Loading initial BGP peer and prefix data...")

	// Check if BGP data is available
	bgpDataAvailable := a.checkBGPDataAvailability(ctx)
	if !bgpDataAvailable {
		glog.Warning("No BGP data found - ip-graph will contain IGP topology only")
		glog.Info("Note: BGP data will be populated as peer sessions and prefixes are received")
		return nil // Continue without BGP data
	}

	// Load existing BGP peer data
	if err := a.loadInitialPeers(ctx); err != nil {
		glog.Warningf("Failed to load initial peers (continuing): %v", err)
	}

	// Load existing BGP prefix data
	if err := a.loadInitialPrefixes(ctx); err != nil {
		glog.Warningf("Failed to load initial prefixes (continuing): %v", err)
	}

	glog.Info("Initial BGP data loaded successfully")
	return nil
}

func (a *arangoDB) checkBGPDataAvailability(ctx context.Context) bool {
	// Check if peer collection exists and has data
	peerExists, err := a.db.CollectionExists(ctx, "peer")
	if err != nil || !peerExists {
		glog.V(6).Info("BGP peer collection 'peer' not found")
		return false
	}

	peerCount, err := a.getCollectionCount(ctx, "peer")
	if err != nil || peerCount == 0 {
		glog.V(6).Info("BGP peer collection 'peer' is empty")
		return false
	}

	glog.V(6).Infof("BGP data available: %d peers found", peerCount)
	return true
}

func (a *arangoDB) loadInitialPeers(ctx context.Context) error {
	glog.V(6).Info("Loading initial BGP peers...")

	// Query all BGP peer sessions
	peerQuery := "FOR p IN peer RETURN p"
	cursor, err := a.db.Query(ctx, peerQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to query peers: %w", err)
	}
	defer cursor.Close()

	peerCount := 0
	for cursor.HasMore() {
		var peerData map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &peerData); err != nil {
			return fmt.Errorf("failed to read peer document: %w", err)
		}

		// Process the peer session (create BGP nodes and session edges)
		if err := a.processInitialPeer(ctx, peerData); err != nil {
			glog.Warningf("Failed to process initial peer %s: %v", getString(peerData, "_key"), err)
			continue
		}

		peerCount++
	}

	glog.V(6).Infof("Loaded %d initial BGP peers", peerCount)
	return nil
}

func (a *arangoDB) loadInitialPrefixes(ctx context.Context) error {
	glog.V(6).Info("Loading initial BGP prefixes...")

	// Load IPv4 unicast prefixes
	if err := a.loadInitialUnicastPrefixes(ctx, "unicast_prefix_v4", true); err != nil {
		return fmt.Errorf("failed to load IPv4 prefixes: %w", err)
	}

	// Load IPv6 unicast prefixes
	if err := a.loadInitialUnicastPrefixes(ctx, "unicast_prefix_v6", false); err != nil {
		return fmt.Errorf("failed to load IPv6 prefixes: %w", err)
	}

	glog.V(6).Info("Initial BGP prefixes loaded successfully")
	return nil
}

func (a *arangoDB) loadInitialUnicastPrefixes(ctx context.Context, collection string, isIPv4 bool) error {
	ipVersion := "IPv6"
	if isIPv4 {
		ipVersion = "IPv4"
	}
	glog.V(6).Infof("Loading initial %s unicast prefixes from %s...", ipVersion, collection)

	// Query all unicast prefixes
	prefixQuery := fmt.Sprintf("FOR u IN %s RETURN u", collection)
	cursor, err := a.db.Query(ctx, prefixQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to query %s: %w", collection, err)
	}
	defer cursor.Close()

	prefixCount := 0
	for cursor.HasMore() {
		var prefixData map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &prefixData); err != nil {
			return fmt.Errorf("failed to read prefix document: %w", err)
		}

		// Add is_ipv4 field to match the processing logic
		prefixData["is_ipv4"] = isIPv4

		// Process the prefix (create BGP prefix vertices and edges)
		if err := a.processInitialPrefix(ctx, prefixData); err != nil {
			glog.Warningf("Failed to process initial prefix %s: %v", getString(prefixData, "_key"), err)
			continue
		}

		prefixCount++
	}

	glog.V(6).Infof("Loaded %d initial %s unicast prefixes", prefixCount, ipVersion)
	return nil
}

func (a *arangoDB) processInitialPeer(ctx context.Context, peerData map[string]interface{}) error {
	// Create a pseudo-message for the BGP peer processor
	procMsg := &ProcessingMessage{
		Key:    getString(peerData, "_key"),
		Action: "add", // Initial load is always "add"
		Data:   peerData,
	}

	// Use the existing BGP peer processor
	uc := &UpdateCoordinator{db: a}
	return uc.processBGPPeerUpdate(procMsg)
}

func (a *arangoDB) processInitialPrefix(ctx context.Context, prefixData map[string]interface{}) error {
	// Create a pseudo-message for the BGP prefix processor
	procMsg := &ProcessingMessage{
		Key:    getString(prefixData, "_key"),
		Action: "add", // Initial load is always "add"
		Data:   prefixData,
	}

	// Use the existing BGP prefix processor
	uc := &UpdateCoordinator{db: a}
	return uc.processBGPPrefixUpdate(procMsg)
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
				processedCount := stats.Processed.Load()
				pendingCount := stats.Pending.Load()
				glog.V(5).Infof("Batch processor stats: processed=%d, pending=%d",
					processedCount, pendingCount)
			}
		}
	}
}

// Helper function to get string value from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// Helper function to get uint32 value from map
func getUint32(v interface{}) uint32 {
	switch val := v.(type) {
	case float64:
		return uint32(val)
	case uint32:
		return val
	case int:
		return uint32(val)
	default:
		return 0
	}
}
