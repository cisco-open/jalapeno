package arangodb

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	driver "github.com/arangodb/go-driver"
	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/cisco-open/jalapeno/gobmp-arango/kafkanotifier"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// UpdateCoordinator manages incoming topology changes and coordinates updates
type UpdateCoordinator struct {
	db        *arangoDB
	batchSize int

	// Processing channels
	nodeUpdates   chan *kafkanotifier.EventMessage
	linkUpdates   chan *kafkanotifier.EventMessage
	prefixUpdates chan *kafkanotifier.EventMessage
	srv6Updates   chan *kafkanotifier.EventMessage

	// Control
	stop    chan struct{}
	wg      sync.WaitGroup
	started bool
}

// NewUpdateCoordinator creates a new update coordinator
func NewUpdateCoordinator(db *arangoDB, batchSize int) *UpdateCoordinator {
	return &UpdateCoordinator{
		db:            db,
		batchSize:     batchSize,
		nodeUpdates:   make(chan *kafkanotifier.EventMessage, batchSize*2),
		linkUpdates:   make(chan *kafkanotifier.EventMessage, batchSize*2),
		prefixUpdates: make(chan *kafkanotifier.EventMessage, batchSize*2),
		srv6Updates:   make(chan *kafkanotifier.EventMessage, batchSize*2),
		stop:          make(chan struct{}),
	}
}

// Start begins the update coordination
func (uc *UpdateCoordinator) Start() error {
	if uc.started {
		return nil
	}

	glog.Info("Starting update coordinator...")

	// Start processors for each message type
	uc.wg.Add(4)
	go uc.nodeUpdateProcessor()
	go uc.linkUpdateProcessor()
	go uc.prefixUpdateProcessor()
	go uc.srv6UpdateProcessor()

	uc.started = true
	glog.Info("Update coordinator started")
	return nil
}

// Stop shuts down the update coordinator
func (uc *UpdateCoordinator) Stop() {
	if !uc.started {
		return
	}

	glog.Info("Stopping update coordinator...")
	close(uc.stop)
	uc.wg.Wait()
	uc.started = false
	glog.Info("Update coordinator stopped")
}

// ProcessMessage processes an incoming raw BMP message and routes it to the appropriate handler
func (uc *UpdateCoordinator) ProcessMessage(msgType dbclient.CollectionType, msg []byte) error {
	if !uc.started {
		return ErrProcessorNotStarted
	}

	// Parse raw BMP data
	var bmpData map[string]interface{}
	if err := json.Unmarshal(msg, &bmpData); err != nil {
		return fmt.Errorf("failed to unmarshal BMP message: %w", err)
	}

	// Create a pseudo-event message for processing
	event := &kafkanotifier.EventMessage{
		TopicType: msgType,
		Key:       getBMPKeyForMessageType(bmpData, msgType),
		Action:    getBMPAction(bmpData),
		ID:        getBMPID(bmpData, msgType),
	}

	glog.V(8).Infof("Processing BMP message: type=%d, key=%s, action=%s", msgType, event.Key, event.Action)

	// Route message to appropriate channel
	switch msgType {
	case bmp.LSNodeMsg:
		select {
		case uc.nodeUpdates <- event:
			return nil
		case <-uc.stop:
			return ErrProcessorStopped
		default:
			return ErrQueueFull
		}

	case bmp.LSLinkMsg:
		select {
		case uc.linkUpdates <- event:
			return nil
		case <-uc.stop:
			return ErrProcessorStopped
		default:
			return ErrQueueFull
		}

	case bmp.LSPrefixMsg:
		select {
		case uc.prefixUpdates <- event:
			return nil
		case <-uc.stop:
			return ErrProcessorStopped
		default:
			return ErrQueueFull
		}

	case bmp.LSSRv6SIDMsg:
		select {
		case uc.srv6Updates <- event:
			return nil
		case <-uc.stop:
			return ErrProcessorStopped
		default:
			return ErrQueueFull
		}

	default:
		glog.V(5).Infof("Unsupported message type: %d", msgType)
		return nil
	}
}

func (uc *UpdateCoordinator) nodeUpdateProcessor() {
	defer uc.wg.Done()
	glog.V(6).Info("Node update processor started")

	for {
		select {
		case <-uc.stop:
			glog.V(6).Info("Node update processor stopped")
			return

		case event := <-uc.nodeUpdates:
			if err := uc.processNodeUpdate(event); err != nil {
				glog.Errorf("Failed to process node update %s: %v", event.Key, err)
			}
		}
	}
}

func (uc *UpdateCoordinator) linkUpdateProcessor() {
	defer uc.wg.Done()
	glog.V(6).Info("Link update processor started")

	for {
		select {
		case <-uc.stop:
			glog.V(6).Info("Link update processor stopped")
			return

		case event := <-uc.linkUpdates:
			if err := uc.processLinkUpdate(event); err != nil {
				glog.Errorf("Failed to process link update %s: %v", event.Key, err)
			}
		}
	}
}

func (uc *UpdateCoordinator) prefixUpdateProcessor() {
	defer uc.wg.Done()
	glog.V(6).Info("Prefix update processor started")

	for {
		select {
		case <-uc.stop:
			glog.V(6).Info("Prefix update processor stopped")
			return

		case event := <-uc.prefixUpdates:
			if err := uc.processPrefixUpdate(event); err != nil {
				glog.Errorf("Failed to process prefix update %s: %v", event.Key, err)
			}
		}
	}
}

func (uc *UpdateCoordinator) srv6UpdateProcessor() {
	defer uc.wg.Done()
	glog.V(6).Info("SRv6 update processor started")

	for {
		select {
		case <-uc.stop:
			glog.V(6).Info("SRv6 update processor stopped")
			return

		case event := <-uc.srv6Updates:
			if err := uc.processSRv6Update(event); err != nil {
				glog.Errorf("Failed to process SRv6 update %s: %v", event.Key, err)
			}
		}
	}
}

// Individual update processors - now with real implementation
func (uc *UpdateCoordinator) processNodeUpdate(event *kafkanotifier.EventMessage) error {
	// Validate event message
	if event == nil {
		return fmt.Errorf("event message is nil")
	}

	if event.Key == "" {
		glog.Errorf("Node event has empty key: ID=%s, Action=%s, TopicType=%d", event.ID, event.Action, event.TopicType)
		return fmt.Errorf("key is empty")
	}

	glog.V(7).Infof("Processing node update: %s action: %s ID: %s", event.Key, event.Action, event.ID)

	ctx := context.TODO()

	switch event.Action {
	case "del":
		// Handle node deletion
		return uc.processNodeDeletion(ctx, event.Key)

	case "add", "update":
		// Handle node addition/update - fetch the actual node data
		return uc.processNodeAddUpdate(ctx, event.Key, event.Action)

	default:
		glog.V(5).Infof("Unknown node action: %s for key: %s", event.Action, event.Key)
		return nil
	}
}

func (uc *UpdateCoordinator) processLinkUpdate(event *kafkanotifier.EventMessage) error {
	// Validate event message
	if event == nil {
		return fmt.Errorf("event message is nil")
	}

	if event.Key == "" {
		glog.Errorf("Link event has empty key: ID=%s, Action=%s, TopicType=%d", event.ID, event.Action, event.TopicType)
		return fmt.Errorf("key is empty")
	}

	glog.V(7).Infof("Processing link update: %s action: %s ID: %s", event.Key, event.Action, event.ID)

	ctx := context.TODO()

	switch event.Action {
	case "del":
		// Handle link deletion
		return uc.processLinkDeletion(ctx, event.Key)

	case "add", "update":
		// Handle link addition/update - fetch the actual link data
		return uc.processLinkAddUpdate(ctx, event.Key, event.Action)

	default:
		glog.V(5).Infof("Unknown link action: %s for key: %s", event.Action, event.Key)
		return nil
	}
}

func (uc *UpdateCoordinator) processPrefixUpdate(event *kafkanotifier.EventMessage) error {
	glog.V(7).Infof("Processing prefix update: %s action: %s", event.Key, event.Action)

	// For now, prefix processing is simplified
	// In future phases, we'll implement the fixed prefix handling strategy:
	// - /32, /128 prefixes -> node metadata
	// - Transit networks -> separate vertices

	ctx := context.TODO()

	switch event.Action {
	case "del":
		glog.V(7).Infof("Prefix deleted: %s", event.Key)
		// TODO: Implement prefix deletion logic

	case "add", "update":
		// Read prefix data
		var prefixData map[string]interface{}
		_, err := uc.db.lsprefix.ReadDocument(ctx, event.Key, &prefixData)
		if err != nil {
			if driver.IsNotFoundGeneral(err) {
				glog.V(6).Infof("Prefix %s not found in ls_prefix collection, skipping", event.Key)
				return nil
			}
			return fmt.Errorf("failed to read prefix %s: %w", event.Key, err)
		}

		glog.V(7).Infof("Prefix %s action %s processed (simplified)", event.Key, event.Action)
		// TODO: Implement actual prefix processing in next phase

	default:
		glog.V(5).Infof("Unknown prefix action: %s for key: %s", event.Action, event.Key)
	}

	return nil
}

func (uc *UpdateCoordinator) processSRv6Update(event *kafkanotifier.EventMessage) error {
	glog.V(7).Infof("Processing SRv6 update: %s action: %s", event.Key, event.Action)

	ctx := context.TODO()

	switch event.Action {
	case "del":
		// Read SRv6 SID data to get router ID for removal
		var srv6Data map[string]interface{}
		_, err := uc.db.lssrv6sid.ReadDocument(ctx, event.Key, &srv6Data)
		if err != nil {
			if driver.IsNotFoundGeneral(err) {
				glog.V(6).Infof("SRv6 SID %s not found for deletion, skipping", event.Key)
				return nil
			}
			return fmt.Errorf("failed to read SRv6 SID %s: %w", event.Key, err)
		}

		return uc.db.removeSRv6SIDFromIGPNode(ctx, event.Key, srv6Data)

	case "add", "update":
		// Read SRv6 SID data
		var srv6Data map[string]interface{}
		_, err := uc.db.lssrv6sid.ReadDocument(ctx, event.Key, &srv6Data)
		if err != nil {
			if driver.IsNotFoundGeneral(err) {
				glog.V(6).Infof("SRv6 SID %s not found, skipping", event.Key)
				return nil
			}
			return fmt.Errorf("failed to read SRv6 SID %s: %w", event.Key, err)
		}

		return uc.db.processSRv6SIDUpdate(ctx, event.Action, event.Key, srv6Data)

	default:
		glog.V(5).Infof("Unknown SRv6 action: %s for key: %s", event.Action, event.Key)
		return nil
	}
}

// Helper functions for real-time processing

func (uc *UpdateCoordinator) processNodeAddUpdate(ctx context.Context, key, action string) error {
	// Read the actual node data from ls_node collection
	var nodeData map[string]interface{}
	_, err := uc.db.lsnode.ReadDocument(ctx, key, &nodeData)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			glog.V(6).Infof("Node %s not found in ls_node collection, skipping", key)
			return nil
		}
		return fmt.Errorf("failed to read node %s: %w", key, err)
	}

	// Filter out BGP nodes (protocol_id = 7)
	if protocolID, ok := nodeData["protocol_id"].(float64); ok && protocolID == 7 {
		glog.V(7).Infof("Skipping BGP node update (protocol_id=7): %s", key)
		return nil
	}

	// Process the node using the same logic as initial loading
	if err := uc.db.processInitialNode(ctx, nodeData); err != nil {
		return fmt.Errorf("failed to process node %s: %w", key, err)
	}

	glog.V(6).Infof("Successfully processed node %s action %s", key, action)
	return nil
}

func (uc *UpdateCoordinator) processNodeDeletion(ctx context.Context, key string) error {
	// Remove from igp_node collection
	if _, err := uc.db.igpNode.RemoveDocument(ctx, key); err != nil {
		if !driver.IsNotFoundGeneral(err) {
			return fmt.Errorf("failed to remove node %s from igp_node: %w", key, err)
		}
	}

	// Remove all edges where this node is referenced
	if err := uc.removeNodeEdges(ctx, key); err != nil {
		return fmt.Errorf("failed to remove edges for node %s: %w", key, err)
	}

	glog.V(6).Infof("Successfully removed node %s", key)
	return nil
}

func (uc *UpdateCoordinator) processLinkAddUpdate(ctx context.Context, key, action string) error {
	// Read the actual link data from ls_link collection
	var linkData map[string]interface{}
	_, err := uc.db.lslink.ReadDocument(ctx, key, &linkData)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			glog.V(6).Infof("Link %s not found in ls_link collection, skipping", key)
			return nil
		}
		return fmt.Errorf("failed to read link %s: %w", key, err)
	}

	// Filter out BGP links (protocol_id = 7)
	if protocolID, ok := linkData["protocol_id"].(float64); ok && protocolID == 7 {
		glog.V(7).Infof("Skipping BGP link update (protocol_id=7): %s", key)
		return nil
	}

	// Process the link using the same logic as initial loading
	if err := uc.db.processInitialLink(ctx, linkData); err != nil {
		return fmt.Errorf("failed to process link %s: %w", key, err)
	}

	glog.V(6).Infof("Successfully processed link %s action %s", key, action)
	return nil
}

func (uc *UpdateCoordinator) processLinkDeletion(ctx context.Context, key string) error {
	// Remove from ls_node_edge collection
	if _, err := uc.db.lsNodeEdge.RemoveDocument(ctx, key); err != nil {
		if !driver.IsNotFoundGeneral(err) {
			glog.V(6).Infof("Link %s not found in ls_node_edge collection", key)
		}
	}

	// Remove from IGP graph collections
	collections := []driver.Collection{}

	// Get IGP graph edge collections
	igpv4Coll, err := uc.db.db.Collection(ctx, uc.db.config.IGPv4Graph)
	if err == nil {
		collections = append(collections, igpv4Coll)
	}

	igpv6Coll, err := uc.db.db.Collection(ctx, uc.db.config.IGPv6Graph)
	if err == nil {
		collections = append(collections, igpv6Coll)
	}

	// Remove from all graph collections
	for _, coll := range collections {
		if _, err := coll.RemoveDocument(ctx, key); err != nil {
			if !driver.IsNotFoundGeneral(err) {
				glog.V(6).Infof("Link %s not found in collection %s", key, coll.Name())
			}
		}
	}

	glog.V(6).Infof("Successfully removed link %s", key)
	return nil
}

func (uc *UpdateCoordinator) removeNodeEdges(ctx context.Context, nodeKey string) error {
	// Build node references for both ls_node and igp_node collections
	lsNodeRef := fmt.Sprintf("%s/%s", uc.db.config.LSNode, nodeKey)
	igpNodeRef := fmt.Sprintf("%s/%s", uc.db.config.IGPNode, nodeKey)

	collections := []driver.Collection{uc.db.lsNodeEdge}

	// Get IGP graph edge collections
	igpv4Coll, err := uc.db.db.Collection(ctx, uc.db.config.IGPv4Graph)
	if err == nil {
		collections = append(collections, igpv4Coll)
	}

	igpv6Coll, err := uc.db.db.Collection(ctx, uc.db.config.IGPv6Graph)
	if err == nil {
		collections = append(collections, igpv6Coll)
	}

	// Remove edges from all collections where this node is referenced
	for _, coll := range collections {
		// Query for edges where this node is _from or _to
		query := fmt.Sprintf(`
			FOR doc IN %s
			FILTER doc._from == @lsNodeRef OR doc._from == @igpNodeRef OR 
			       doc._to == @lsNodeRef OR doc._to == @igpNodeRef
			RETURN doc._key`, coll.Name())

		bindVars := map[string]interface{}{
			"lsNodeRef":  lsNodeRef,
			"igpNodeRef": igpNodeRef,
		}

		cursor, err := uc.db.db.Query(ctx, query, bindVars)
		if err != nil {
			glog.Errorf("Failed to query edges for node %s in collection %s: %v", nodeKey, coll.Name(), err)
			continue
		}

		// Remove each edge found
		for {
			var edgeKey string
			_, err := cursor.ReadDocument(ctx, &edgeKey)
			if err != nil {
				if driver.IsNoMoreDocuments(err) {
					break
				}
				glog.Errorf("Error reading edge key: %v", err)
				continue
			}

			if _, err := coll.RemoveDocument(ctx, edgeKey); err != nil {
				glog.Errorf("Failed to remove edge %s: %v", edgeKey, err)
			} else {
				glog.V(7).Infof("Removed edge %s from %s", edgeKey, coll.Name())
			}
		}
		cursor.Close()
	}

	return nil
}
