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

	ipv4Graph, err := a.CreateGraph(ctx, a.config.IPv4Graph+"_graph", ipv4GraphOptions)
	if err != nil {
		return fmt.Errorf("failed to create IPv4 graph: %w", err)
	}
	a.ipv4GraphDB = ipv4Graph
	glog.V(6).Infof("Ensured IPv4 full topology graph: %s", a.config.IPv4Graph+"_graph")

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

	ipv6Graph, err := a.CreateGraph(ctx, a.config.IPv6Graph+"_graph", ipv6GraphOptions)
	if err != nil {
		return fmt.Errorf("failed to create IPv6 graph: %w", err)
	}
	a.ipv6GraphDB = ipv6Graph
	glog.V(6).Infof("Ensured IPv6 full topology graph: %s", a.config.IPv6Graph+"_graph")

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

	// Use the dedicated IGP copy processor for enhanced copying logic
	igpCopyProcessor := NewIGPCopyProcessor(a)
	if err := igpCopyProcessor.CopyIGPGraphsToFullTopology(ctx); err != nil {
		return fmt.Errorf("failed to copy IGP graphs: %w", err)
	}

	glog.Info("IGP graph data copied to IP graphs successfully")
	return nil
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
	// TODO: Load existing BGP peer and prefix data
	// This will be implemented in the next phase
	glog.V(6).Info("BGP data loading will be implemented in next phase")
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
