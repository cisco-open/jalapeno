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
	igpSyncProcessor  *IGPSyncProcessor

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

	// Initialize and start IGP sync processor for periodic reconciliation
	a.igpSyncProcessor = NewIGPSyncProcessor(a)
	a.igpSyncProcessor.StartReconciliation()
	glog.Info("IGP topology reconciliation started")

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

	if a.igpSyncProcessor != nil {
		a.igpSyncProcessor.StopReconciliation()
		glog.Info("IGP topology reconciliation stopped")
	}

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

	// Use the IGP sync processor to copy edges from igpv4_graph/igpv6_graph to ipv4_graph/ipv6_graph
	// This syncs the already-processed IGP topology from igp-graph processor
	igpSyncProcessor := NewIGPSyncProcessor(a)
	if err := igpSyncProcessor.InitialIGPSync(ctx); err != nil {
		glog.Warningf("Failed to sync IGP graphs (continuing with BGP-only): %v", err)
		return nil // Continue without IGP data rather than failing
	}

	glog.Info("IGP graph data synced to IP graphs successfully")
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

	// Step 1: Create BGP nodes using bulk query (matching original logic)
	if err := a.createInitialBGPNodes(ctx); err != nil {
		glog.Warningf("Failed to create initial BGP nodes (continuing): %v", err)
	}

	// Step 2: Process BGP peer sessions to create session edges
	if err := a.loadInitialPeers(ctx); err != nil {
		glog.Warningf("Failed to load initial peers (continuing): %v", err)
	}

	// Step 2: BGP prefix classification and deduplication (keep existing logic)
	deduplicationProcessor := NewBGPDeduplicationProcessor(a)
	if err := deduplicationProcessor.ProcessInitialBGPDeduplication(ctx); err != nil {
		glog.Warningf("Failed to process BGP prefix deduplication (continuing): %v", err)
	}

	// Step 3: Create prefix-to-node attachments and edges
	if err := a.createPrefixToNodeAttachments(ctx); err != nil {
		glog.Warningf("Failed to create prefix-to-node attachments (continuing): %v", err)
	}

	// Step 4: Apply simple BGP routing precedence (replaces complex unified prefix logic)
	if err := a.applyBGPPrecedence(ctx); err != nil {
		glog.Warningf("Failed to apply BGP precedence (continuing): %v", err)
	}

	// Step 5: Handle iBGP-only nodes (e.g., Cilium) that should attach to subnets
	ibgpSubnetProcessor := NewIBGPSubnetProcessor(a)
	if err := ibgpSubnetProcessor.ProcessIBGPSubnetAttachment(ctx); err != nil {
		glog.Warningf("Failed to process iBGP subnet attachments (continuing): %v", err)
	}

	glog.Info("Initial BGP data loaded successfully")
	return nil
}

// applyBGPPrecedence applies simple routing precedence: BGP beats IGP
// Removes IGP edges where eBGP prefixes exist (iBGP keeps IGP edges for internal redistribution)
func (a *arangoDB) applyBGPPrecedence(ctx context.Context) error {
	glog.V(6).Info("Applying BGP routing precedence (removing conflicting IGP edges)...")

	// Apply precedence for IPv4
	if err := a.applyBGPPrecedenceIPv4(ctx); err != nil {
		return fmt.Errorf("failed to apply IPv4 BGP precedence: %w", err)
	}

	// Apply precedence for IPv6
	if err := a.applyBGPPrecedenceIPv6(ctx); err != nil {
		return fmt.Errorf("failed to apply IPv6 BGP precedence: %w", err)
	}

	glog.V(6).Info("BGP routing precedence applied successfully")
	return nil
}

// applyBGPPrecedenceIPv4 removes conflicting IPv4 IGP edges where eBGP takes precedence
func (a *arangoDB) applyBGPPrecedenceIPv4(ctx context.Context) error {
	glog.V(7).Info("Applying IPv4 BGP precedence...")

	// Remove IGP edges where eBGP takes precedence
	// iBGP prefixes keep their IGP edges (internal redistribution)
	query := fmt.Sprintf(`
		FOR bgp IN %s
		FILTER bgp.prefix_type IN ["ebgp_private", "ebgp_private_4byte", "ebgp_public"]
		FOR igp_edge IN %s
		FILTER igp_edge.prefix == bgp.prefix 
		FILTER igp_edge.prefix_len == bgp.prefix_len
		FILTER STARTS_WITH(igp_edge._to, "ls_prefix/") OR STARTS_WITH(igp_edge._from, "ls_prefix/")
		
		// BGP takes precedence over IGP - remove IGP edge
		REMOVE igp_edge IN %s
	`, a.config.BGPPrefixV4, a.config.IPv4Graph, a.config.IPv4Graph)

	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to apply IPv4 BGP precedence: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("IPv4 BGP precedence applied")
	return nil
}

// applyBGPPrecedenceIPv6 removes conflicting IPv6 IGP edges where eBGP takes precedence
func (a *arangoDB) applyBGPPrecedenceIPv6(ctx context.Context) error {
	glog.V(7).Info("Applying IPv6 BGP precedence...")

	// Remove IGP edges where eBGP takes precedence
	// iBGP prefixes keep their IGP edges (internal redistribution)
	query := fmt.Sprintf(`
		FOR bgp IN %s
		FILTER bgp.prefix_type IN ["ebgp_private", "ebgp_private_4byte", "ebgp_public"]
		FOR igp_edge IN %s
		FILTER igp_edge.prefix == bgp.prefix 
		FILTER igp_edge.prefix_len == bgp.prefix_len
		FILTER STARTS_WITH(igp_edge._to, "ls_prefix/") OR STARTS_WITH(igp_edge._from, "ls_prefix/")
		
		// BGP takes precedence over IGP - remove IGP edge
		REMOVE igp_edge IN %s
	`, a.config.BGPPrefixV6, a.config.IPv6Graph, a.config.IPv6Graph)

	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to apply IPv6 BGP precedence: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("IPv6 BGP precedence applied")
	return nil
}

func (a *arangoDB) createInitialBGPNodes(ctx context.Context) error {
	glog.V(6).Info("Creating initial BGP nodes using bulk query (matching original logic)...")

	// Create BGP nodes for all remote peers NOT in IGP domain (using peer_asn)
	// This matches the original query exactly
	bgpNodeQuery := fmt.Sprintf(`
		FOR p IN peer 
		LET igp_asns = (FOR n IN %s RETURN n.peer_asn)
		FILTER p.remote_asn NOT IN igp_asns
		INSERT { 
			_key: CONCAT_SEPARATOR("_", p.remote_bgp_id, p.remote_asn),
			router_id: p.remote_bgp_id,
			asn: p.remote_asn
		} INTO %s OPTIONS { ignoreErrors: true }
	`, a.config.IGPNode, a.config.BGPNode)

	cursor, err := a.db.Query(ctx, bgpNodeQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to create BGP nodes: %w", err)
	}
	defer cursor.Close()

	glog.V(6).Info("Initial BGP nodes created successfully")
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

// createPrefixToNodeAttachments creates bidirectional edges between deduplicated prefixes and their origin nodes
func (a *arangoDB) createPrefixToNodeAttachments(ctx context.Context) error {
	glog.V(6).Info("Creating prefix-to-node attachments...")

	// Process IPv4 prefix attachments
	if err := a.createIPv4PrefixAttachments(ctx); err != nil {
		return fmt.Errorf("failed to create IPv4 prefix attachments: %w", err)
	}

	// Process IPv6 prefix attachments
	if err := a.createIPv6PrefixAttachments(ctx); err != nil {
		return fmt.Errorf("failed to create IPv6 prefix attachments: %w", err)
	}

	glog.V(6).Info("Prefix-to-node attachments created successfully")
	return nil
}

// createIPv4PrefixAttachments creates IPv4 prefix-to-node attachments
func (a *arangoDB) createIPv4PrefixAttachments(ctx context.Context) error {
	glog.V(7).Info("Creating IPv4 prefix-to-node attachments...")

	// Query all deduplicated IPv4 prefixes
	query := fmt.Sprintf("FOR p IN %s RETURN p", a.config.BGPPrefixV4)
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query IPv4 prefixes: %w", err)
	}
	defer cursor.Close()

	attachmentCount := 0
	for cursor.HasMore() {
		var prefixData map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &prefixData); err != nil {
			return fmt.Errorf("failed to read IPv4 prefix: %w", err)
		}

		// Create prefix-to-node attachment
		if err := a.createPrefixAttachment(ctx, prefixData, true); err != nil {
			glog.Warningf("Failed to create IPv4 prefix attachment for %s: %v", getString(prefixData, "_key"), err)
			continue
		}

		attachmentCount++
	}

	glog.V(7).Infof("Created %d IPv4 prefix-to-node attachments", attachmentCount)
	return nil
}

// createIPv6PrefixAttachments creates IPv6 prefix-to-node attachments
func (a *arangoDB) createIPv6PrefixAttachments(ctx context.Context) error {
	glog.V(7).Info("Creating IPv6 prefix-to-node attachments...")

	// Query all deduplicated IPv6 prefixes
	query := fmt.Sprintf("FOR p IN %s RETURN p", a.config.BGPPrefixV6)
	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query IPv6 prefixes: %w", err)
	}
	defer cursor.Close()

	attachmentCount := 0
	for cursor.HasMore() {
		var prefixData map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &prefixData); err != nil {
			return fmt.Errorf("failed to read IPv6 prefix: %w", err)
		}

		// Create prefix-to-node attachment
		if err := a.createPrefixAttachment(ctx, prefixData, false); err != nil {
			glog.Warningf("Failed to create IPv6 prefix attachment for %s: %v", getString(prefixData, "_key"), err)
			continue
		}

		attachmentCount++
	}

	glog.V(7).Infof("Created %d IPv6 prefix-to-node attachments", attachmentCount)
	return nil
}

// createPrefixAttachment creates bidirectional edges between a prefix and its reachable nodes
func (a *arangoDB) createPrefixAttachment(ctx context.Context, prefixData map[string]interface{}, isIPv4 bool) error {
	prefixType := getString(prefixData, "prefix_type")
	routerID := getString(prefixData, "router_id")
	originAS := getUint32FromInterface(prefixData["origin_as"])

	// Determine target collections
	var prefixCollection string
	if isIPv4 {
		prefixCollection = a.config.BGPPrefixV4
	} else {
		prefixCollection = a.config.BGPPrefixV6
	}

	// Find reachable nodes based on prefix type (following original logic)
	switch prefixType {
	case "ibgp":
		// iBGP prefixes: Connect to IGP nodes by router_id and ASN
		return a.attachPrefixToIGPNodes(ctx, prefixData, prefixCollection, isIPv4)
	case "ebgp_private", "ebgp_private_4byte":
		// eBGP private prefixes: Connect to specific BGP node by router_id and origin_as
		return a.attachPrefixToSpecificBGPNode(ctx, prefixData, prefixCollection, isIPv4, routerID, originAS)
	case "ebgp_public":
		// Internet prefixes: Connect to all BGP peer nodes with public ASNs (like original processInetPrefix)
		return a.attachPrefixToInternetPeers(ctx, prefixData, prefixCollection, isIPv4)
	default:
		return fmt.Errorf("unknown prefix type: %s", prefixType)
	}
}

// attachPrefixToIGPNodes attaches iBGP prefixes to IGP nodes (original processIbgpPrefix logic)
func (a *arangoDB) attachPrefixToIGPNodes(ctx context.Context, prefixData map[string]interface{}, prefixCollection string, isIPv4 bool) error {
	routerID := getString(prefixData, "router_id")
	asn := getUint32FromInterface(prefixData["asn"])
	prefixKey := getString(prefixData, "_key")

	// Query IGP nodes by router_id and ASN (matching original logic)
	query := fmt.Sprintf(`
		FOR node IN %s 
		FILTER node.router_id == @routerId AND node.asn == @asn 
		RETURN node
	`, a.config.IGPNode)

	bindVars := map[string]interface{}{
		"routerId": routerID,
		"asn":      asn,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to query IGP nodes: %w", err)
	}
	defer cursor.Close()

	// Create edges to all matching IGP nodes
	for cursor.HasMore() {
		var igpNode map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &igpNode); err != nil {
			return fmt.Errorf("failed to read IGP node: %w", err)
		}

		if err := a.createBidirectionalPrefixEdges(ctx, prefixData, igpNode, prefixCollection, isIPv4); err != nil {
			glog.Warningf("Failed to create iBGP prefix edges for %s: %v", prefixKey, err)
		}
	}

	return nil
}

// attachPrefixToSpecificBGPNode attaches eBGP prefixes to specific BGP node (original processeBgpPrefix logic)
func (a *arangoDB) attachPrefixToSpecificBGPNode(ctx context.Context, prefixData map[string]interface{}, prefixCollection string, isIPv4 bool, routerID string, originAS uint32) error {
	prefixKey := getString(prefixData, "_key")

	// Query specific BGP node by router_id and origin_as (matching original logic)
	query := fmt.Sprintf(`
		FOR node IN %s 
		FILTER node.router_id == @routerId AND node.asn == @originAs 
		RETURN node
	`, a.config.BGPNode)

	bindVars := map[string]interface{}{
		"routerId": routerID,
		"originAs": originAS,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to query BGP nodes: %w", err)
	}
	defer cursor.Close()

	// Create edges to all matching BGP nodes
	for cursor.HasMore() {
		var bgpNode map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &bgpNode); err != nil {
			return fmt.Errorf("failed to read BGP node: %w", err)
		}

		if err := a.createBidirectionalPrefixEdges(ctx, prefixData, bgpNode, prefixCollection, isIPv4); err != nil {
			glog.Warningf("Failed to create eBGP prefix edges for %s: %v", prefixKey, err)
		}
	}

	return nil
}

// attachPrefixToInternetPeers attaches Internet prefixes to BGP peer nodes (original processInetPrefix logic)
func (a *arangoDB) attachPrefixToInternetPeers(ctx context.Context, prefixData map[string]interface{}, prefixCollection string, isIPv4 bool) error {
	prefixKey := getString(prefixData, "_key")

	// Query all BGP nodes with public ASNs (matching original logic)
	query := fmt.Sprintf(`
		FOR node IN %s 
		FILTER node.asn NOT IN 64512..65535 
		FILTER node.asn NOT IN 4200000000..4294967294 
		RETURN node
	`, a.config.BGPNode)

	cursor, err := a.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to query Internet BGP peers: %w", err)
	}
	defer cursor.Close()

	// Create edges to all Internet BGP peer nodes
	for cursor.HasMore() {
		var bgpNode map[string]interface{}
		if _, err := cursor.ReadDocument(ctx, &bgpNode); err != nil {
			return fmt.Errorf("failed to read BGP peer node: %w", err)
		}

		if err := a.createBidirectionalPrefixEdges(ctx, prefixData, bgpNode, prefixCollection, isIPv4); err != nil {
			glog.Warningf("Failed to create Internet prefix edges for %s: %v", prefixKey, err)
		}
	}

	return nil
}

// createBidirectionalPrefixEdges creates bidirectional edges between prefix and node (matching original edge structure)
func (a *arangoDB) createBidirectionalPrefixEdges(ctx context.Context, prefixData, nodeData map[string]interface{}, prefixCollection string, isIPv4 bool) error {
	prefixKey := getString(prefixData, "_key")
	nodeKey := getString(nodeData, "_key")
	nodeID := getString(nodeData, "_id")
	prefixVertexID := fmt.Sprintf("%s/%s", prefixCollection, prefixKey)

	// Extract edge metadata
	prefix := getString(prefixData, "prefix")
	prefixLen := getUint32FromInterface(prefixData["prefix_len"])
	originAS := getUint32FromInterface(prefixData["origin_as"])
	prefixType := getString(prefixData, "prefix_type")

	// Edge from node to prefix (matching original unicastPrefixEdgeObject structure)
	nodeToPrefix := &IPGraphObject{
		Key:       fmt.Sprintf("%s_%s", nodeKey, prefixKey),
		From:      nodeID,
		To:        prefixVertexID,
		Protocol:  fmt.Sprintf("BGP_%s", prefixType),
		Link:      prefixKey,
		Prefix:    prefix,
		PrefixLen: int32(prefixLen),
		OriginAS:  int32(originAS),
	}

	// Edge from prefix to node
	prefixToNode := &IPGraphObject{
		Key:       fmt.Sprintf("%s_%s", prefixKey, nodeKey),
		From:      prefixVertexID,
		To:        nodeID,
		Protocol:  fmt.Sprintf("BGP_%s", prefixType),
		Link:      prefixKey,
		Prefix:    prefix,
		PrefixLen: int32(prefixLen),
		OriginAS:  int32(originAS),
	}

	// Insert edges into the appropriate graph collection
	var targetCollection driver.Collection
	if isIPv4 {
		targetCollection = a.ipv4Graph
	} else {
		targetCollection = a.ipv6Graph
	}

	// Create both edges (with conflict handling like original)
	for _, edge := range []*IPGraphObject{nodeToPrefix, prefixToNode} {
		if _, err := targetCollection.CreateDocument(ctx, edge); err != nil {
			if !driver.IsConflict(err) {
				return fmt.Errorf("failed to create edge %s: %w", edge.Key, err)
			}
			// Update existing edge (matching original logic)
			if _, err := targetCollection.UpdateDocument(ctx, edge.Key, edge); err != nil {
				return fmt.Errorf("failed to update edge %s: %w", edge.Key, err)
			}
		}
	}

	return nil
}

// findIGPNodeByRouterIDAndASN finds an IGP node by router ID and ASN
func (a *arangoDB) findIGPNodeByRouterIDAndASN(ctx context.Context, routerID string, asn uint32) (string, error) {
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.router_id == @routerId AND node.asn == @asn
		LIMIT 1
		RETURN node._id
	`, a.config.IGPNode)

	bindVars := map[string]interface{}{
		"routerId": routerID,
		"asn":      asn,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return "", fmt.Errorf("failed to query IGP nodes: %w", err)
	}
	defer cursor.Close()

	if cursor.HasMore() {
		var nodeID string
		if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
			return "", fmt.Errorf("failed to read IGP node ID: %w", err)
		}
		return nodeID, nil
	}

	return "", nil // Not found
}

// findBGPNodeByRouterIDAndASN finds a BGP node by router ID and ASN
func (a *arangoDB) findBGPNodeByRouterIDAndASN(ctx context.Context, routerID string, asn uint32) (string, error) {
	query := fmt.Sprintf(`
		FOR node IN %s
		FILTER node.router_id == @routerId AND node.asn == @asn
		LIMIT 1
		RETURN node._id
	`, a.config.BGPNode)

	bindVars := map[string]interface{}{
		"routerId": routerID,
		"asn":      asn,
	}

	cursor, err := a.db.Query(ctx, query, bindVars)
	if err != nil {
		return "", fmt.Errorf("failed to query BGP nodes: %w", err)
	}
	defer cursor.Close()

	if cursor.HasMore() {
		var nodeID string
		if _, err := cursor.ReadDocument(ctx, &nodeID); err != nil {
			return "", fmt.Errorf("failed to read BGP node ID: %w", err)
		}
		return nodeID, nil
	}

	return "", nil // Not found
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
