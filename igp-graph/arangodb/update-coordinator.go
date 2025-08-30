package arangodb

import (
	"encoding/json"
	"fmt"
	"sync"

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

// ProcessMessage processes an incoming message and routes it to the appropriate handler
func (uc *UpdateCoordinator) ProcessMessage(msgType dbclient.CollectionType, msg []byte) error {
	if !uc.started {
		return ErrProcessorNotStarted
	}

	event := &kafkanotifier.EventMessage{}
	if err := json.Unmarshal(msg, event); err != nil {
		return fmt.Errorf("failed to unmarshal event message: %w", err)
	}

	event.TopicType = msgType

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

// Individual update processors - these will be implemented in the next phase
func (uc *UpdateCoordinator) processNodeUpdate(event *kafkanotifier.EventMessage) error {
	glog.V(7).Infof("Processing node update: %s action: %s", event.Key, event.Action)

	// Create operation for batch processor
	op := &NodeOperation{
		Type: event.Action,
		Key:  event.Key,
		Data: make(map[string]interface{}),
		Done: make(chan error, 1),
	}

	// Submit to batch processor
	if err := uc.db.batchProcessor.SubmitNodeOperation(op); err != nil {
		return fmt.Errorf("failed to submit node operation: %w", err)
	}

	// Wait for completion (optional - could be async)
	select {
	case err := <-op.Done:
		return err
	case <-uc.stop:
		return ErrProcessorStopped
	}
}

func (uc *UpdateCoordinator) processLinkUpdate(event *kafkanotifier.EventMessage) error {
	glog.V(7).Infof("Processing link update: %s action: %s", event.Key, event.Action)

	op := &LinkOperation{
		Type: event.Action,
		Key:  event.Key,
		Data: make(map[string]interface{}),
		Done: make(chan error, 1),
	}

	if err := uc.db.batchProcessor.SubmitLinkOperation(op); err != nil {
		return fmt.Errorf("failed to submit link operation: %w", err)
	}

	select {
	case err := <-op.Done:
		return err
	case <-uc.stop:
		return ErrProcessorStopped
	}
}

func (uc *UpdateCoordinator) processPrefixUpdate(event *kafkanotifier.EventMessage) error {
	glog.V(7).Infof("Processing prefix update: %s action: %s", event.Key, event.Action)

	op := &PrefixOperation{
		Type: event.Action,
		Key:  event.Key,
		Data: make(map[string]interface{}),
		Done: make(chan error, 1),
	}

	if err := uc.db.batchProcessor.SubmitPrefixOperation(op); err != nil {
		return fmt.Errorf("failed to submit prefix operation: %w", err)
	}

	select {
	case err := <-op.Done:
		return err
	case <-uc.stop:
		return ErrProcessorStopped
	}
}

func (uc *UpdateCoordinator) processSRv6Update(event *kafkanotifier.EventMessage) error {
	glog.V(7).Infof("Processing SRv6 update: %s action: %s", event.Key, event.Action)

	// SRv6 updates are typically merged into node data
	op := &NodeOperation{
		Type: "srv6_update",
		Key:  event.Key,
		Data: make(map[string]interface{}),
		Done: make(chan error, 1),
	}

	if err := uc.db.batchProcessor.SubmitNodeOperation(op); err != nil {
		return fmt.Errorf("failed to submit SRv6 operation: %w", err)
	}

	select {
	case err := <-op.Done:
		return err
	case <-uc.stop:
		return ErrProcessorStopped
	}
}
