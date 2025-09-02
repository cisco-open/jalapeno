package arangodb

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/cisco-open/jalapeno/gobmp-arango/dbclient"
	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
)

// UpdateCoordinator coordinates real-time updates from Kafka messages
type UpdateCoordinator struct {
	db *arangoDB

	// Processing channels
	igpUpdates    chan *ProcessingMessage
	bgpUpdates    chan *ProcessingMessage
	prefixUpdates chan *ProcessingMessage

	// Control
	stop    chan struct{}
	wg      sync.WaitGroup
	started bool
	mu      sync.RWMutex
}

// ProcessingMessage represents a message for processing
type ProcessingMessage struct {
	Type   dbclient.CollectionType
	Key    string
	Action string
	ID     string
	Data   map[string]interface{}
}

// NewUpdateCoordinator creates a new update coordinator
func NewUpdateCoordinator(db *arangoDB) *UpdateCoordinator {
	return &UpdateCoordinator{
		db:            db,
		igpUpdates:    make(chan *ProcessingMessage, 1000),
		bgpUpdates:    make(chan *ProcessingMessage, 1000),
		prefixUpdates: make(chan *ProcessingMessage, 1000),
		stop:          make(chan struct{}),
	}
}

// Start starts the update coordinator
func (uc *UpdateCoordinator) Start() error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.started {
		return nil
	}

	// Start processing workers
	uc.wg.Add(3)
	go uc.igpUpdateWorker()
	go uc.bgpUpdateWorker()
	go uc.prefixUpdateWorker()

	uc.started = true
	glog.Info("Starting update coordinator...")

	return nil
}

// Stop stops the update coordinator
func (uc *UpdateCoordinator) Stop() {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if !uc.started {
		return
	}

	close(uc.stop)
	uc.wg.Wait()
	uc.started = false

	glog.Info("Update coordinator stopped")
}

// ProcessMessage processes a raw BMP message
func (uc *UpdateCoordinator) ProcessMessage(msgType dbclient.CollectionType, msg []byte) error {
	if !uc.started {
		return ErrProcessorNotStarted
	}

	// Parse raw BMP data
	var bmpData map[string]interface{}
	if err := json.Unmarshal(msg, &bmpData); err != nil {
		return fmt.Errorf("failed to unmarshal BMP message: %w", err)
	}

	// Create processing message
	procMsg := &ProcessingMessage{
		Type:   msgType,
		Key:    getBMPKeyForMessageType(bmpData, msgType),
		Action: getBMPAction(bmpData),
		ID:     getBMPID(bmpData, msgType),
		Data:   bmpData,
	}

	glog.V(8).Infof("Processing BMP message: type=%d, key=%s, action=%s", msgType, procMsg.Key, procMsg.Action)

	// Route message to appropriate worker
	switch msgType {
	case bmp.LSNodeMsg, bmp.LSLinkMsg, bmp.LSPrefixMsg, bmp.LSSRv6SIDMsg:
		// IGP sync messages
		select {
		case uc.igpUpdates <- procMsg:
			return nil
		case <-uc.stop:
			return ErrProcessorStopped
		default:
			return ErrQueueFull
		}

	case bmp.PeerStateChangeMsg:
		// BGP peer messages
		select {
		case uc.bgpUpdates <- procMsg:
			return nil
		case <-uc.stop:
			return ErrProcessorStopped
		default:
			return ErrQueueFull
		}

	case bmp.UnicastPrefixV4Msg, bmp.UnicastPrefixV6Msg:
		// BGP prefix messages
		select {
		case uc.prefixUpdates <- procMsg:
			return nil
		case <-uc.stop:
			return ErrProcessorStopped
		default:
			return ErrQueueFull
		}

	default:
		glog.V(5).Infof("Unknown message type: %d", msgType)
		return nil
	}
}

func (uc *UpdateCoordinator) igpUpdateWorker() {
	defer uc.wg.Done()

	for {
		select {
		case <-uc.stop:
			return

		case msg := <-uc.igpUpdates:
			if err := uc.processIGPUpdate(msg); err != nil {
				glog.Errorf("Failed to process IGP update %s: %v", msg.Key, err)
			}
		}
	}
}

func (uc *UpdateCoordinator) bgpUpdateWorker() {
	defer uc.wg.Done()

	for {
		select {
		case <-uc.stop:
			return

		case msg := <-uc.bgpUpdates:
			if err := uc.processBGPUpdate(msg); err != nil {
				glog.Errorf("Failed to process BGP update %s: %v", msg.Key, err)
			}
		}
	}
}

func (uc *UpdateCoordinator) prefixUpdateWorker() {
	defer uc.wg.Done()

	for {
		select {
		case <-uc.stop:
			return

		case msg := <-uc.prefixUpdates:
			if err := uc.processPrefixUpdate(msg); err != nil {
				glog.Errorf("Failed to process prefix update %s: %v", msg.Key, err)
			}
		}
	}
}

func (uc *UpdateCoordinator) processIGPUpdate(msg *ProcessingMessage) error {
	glog.V(7).Infof("Processing IGP update: %s action: %s", msg.Key, msg.Action)

	// TODO: Implement IGP sync logic
	// This will sync changes from igpv4_graph/igpv6_graph to ipv4_graph/ipv6_graph

	switch msg.Type {
	case bmp.LSNodeMsg:
		return uc.processIGPNodeUpdate(msg)
	case bmp.LSLinkMsg:
		return uc.processIGPLinkUpdate(msg)
	case bmp.LSPrefixMsg:
		return uc.processIGPPrefixUpdate(msg)
	case bmp.LSSRv6SIDMsg:
		return uc.processIGPSRv6Update(msg)
	}

	return nil
}

func (uc *UpdateCoordinator) processBGPUpdate(msg *ProcessingMessage) error {
	glog.V(7).Infof("Processing BGP update: %s action: %s", msg.Key, msg.Action)

	// Process BGP peer sessions
	return uc.processBGPPeerUpdate(msg)
}

func (uc *UpdateCoordinator) processPrefixUpdate(msg *ProcessingMessage) error {
	glog.V(7).Infof("Processing prefix update: %s action: %s", msg.Key, msg.Action)

	// Process BGP unicast prefixes
	return uc.processBGPPrefixUpdate(msg)
}

// IGP sync processing methods
func (uc *UpdateCoordinator) processIGPNodeUpdate(msg *ProcessingMessage) error {
	// Sync IGP node changes to full topology
	return uc.syncIGPNodeUpdate(context.TODO(), msg.Key, msg.Action)
}

func (uc *UpdateCoordinator) processIGPLinkUpdate(msg *ProcessingMessage) error {
	// Sync IGP link changes to full topology
	// Determine if this is IPv4 or IPv6 based on message data
	isIPv4 := true
	if mtidData, ok := msg.Data["mt_id_tlv"]; ok {
		if mtidMap, ok := mtidData.(map[string]interface{}); ok {
			if mtid, ok := mtidMap["mt_id"].(float64); ok && mtid == 2 {
				isIPv4 = false
			}
		}
	}

	return uc.syncIGPLinkUpdate(context.TODO(), msg.Key, msg.Action, isIPv4)
}

func (uc *UpdateCoordinator) processIGPPrefixUpdate(msg *ProcessingMessage) error {
	// TODO: Sync IGP prefix changes to full topology
	glog.V(8).Infof("IGP prefix update: %s", msg.Key)
	return nil
}

func (uc *UpdateCoordinator) processIGPSRv6Update(msg *ProcessingMessage) error {
	// TODO: Sync IGP SRv6 changes to full topology
	glog.V(8).Infof("IGP SRv6 update: %s", msg.Key)
	return nil
}
