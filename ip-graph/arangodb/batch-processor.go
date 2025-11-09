package arangodb

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
)

// BatchProcessor handles batch operations for database writes
type BatchProcessor struct {
	db        *arangoDB
	batchSize int
	workers   int

	// Processing channels
	nodeOps   chan *BatchOperation
	edgeOps   chan *BatchOperation
	prefixOps chan *BatchOperation

	// Control channels
	stop    chan struct{}
	wg      sync.WaitGroup
	started bool
	mu      sync.RWMutex

	// Statistics
	processed atomic.Int64
	pending   atomic.Int64
}

// BatchStats represents batch processor statistics
type BatchStats struct {
	Processed atomic.Int64
	Pending   atomic.Int64
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(db *arangoDB, batchSize, workers int) *BatchProcessor {
	return &BatchProcessor{
		db:        db,
		batchSize: batchSize,
		workers:   workers,
		nodeOps:   make(chan *BatchOperation, batchSize*2),
		edgeOps:   make(chan *BatchOperation, batchSize*2),
		prefixOps: make(chan *BatchOperation, batchSize*2),
		stop:      make(chan struct{}),
	}
}

// Start starts the batch processor
func (bp *BatchProcessor) Start() error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.started {
		return nil
	}

	// Start worker goroutines
	for i := 0; i < bp.workers; i++ {
		bp.wg.Add(3)
		go bp.nodeWorker()
		go bp.edgeWorker()
		go bp.prefixWorker()
	}

	// Start flush timer
	bp.wg.Add(1)
	go bp.flushTimer()

	bp.started = true
	glog.Infof("Starting batch processor with %d workers, batch size %d", bp.workers, bp.batchSize)

	return nil
}

// Stop stops the batch processor
func (bp *BatchProcessor) Stop() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if !bp.started {
		return
	}

	close(bp.stop)
	bp.wg.Wait()
	bp.started = false

	glog.Info("Batch processor stopped")
}

// SubmitNodeOperation submits a node operation for batch processing
func (bp *BatchProcessor) SubmitNodeOperation(op *BatchOperation) error {
	select {
	case bp.nodeOps <- op:
		bp.pending.Add(1)
		return nil
	case <-bp.stop:
		return ErrProcessorStopped
	default:
		return ErrQueueFull
	}
}

// SubmitEdgeOperation submits an edge operation for batch processing
func (bp *BatchProcessor) SubmitEdgeOperation(op *BatchOperation) error {
	select {
	case bp.edgeOps <- op:
		bp.pending.Add(1)
		return nil
	case <-bp.stop:
		return ErrProcessorStopped
	default:
		return ErrQueueFull
	}
}

// SubmitPrefixOperation submits a prefix operation for batch processing
func (bp *BatchProcessor) SubmitPrefixOperation(op *BatchOperation) error {
	select {
	case bp.prefixOps <- op:
		bp.pending.Add(1)
		return nil
	case <-bp.stop:
		return ErrProcessorStopped
	default:
		return ErrQueueFull
	}
}

// GetStats returns current processing statistics
func (bp *BatchProcessor) GetStats() BatchStats {
	var stats BatchStats
	stats.Processed.Store(bp.processed.Load())
	stats.Pending.Store(bp.pending.Load())
	return stats
}

func (bp *BatchProcessor) nodeWorker() {
	defer bp.wg.Done()

	batch := make([]*BatchOperation, 0, bp.batchSize)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bp.stop:
			if len(batch) > 0 {
				bp.processBatch("node", batch)
			}
			return

		case op := <-bp.nodeOps:
			batch = append(batch, op)

			if len(batch) >= bp.batchSize {
				bp.processBatch("node", batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				bp.processBatch("node", batch)
				batch = batch[:0]
			}
		}
	}
}

func (bp *BatchProcessor) edgeWorker() {
	defer bp.wg.Done()

	batch := make([]*BatchOperation, 0, bp.batchSize)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bp.stop:
			if len(batch) > 0 {
				bp.processBatch("edge", batch)
			}
			return

		case op := <-bp.edgeOps:
			batch = append(batch, op)

			if len(batch) >= bp.batchSize {
				bp.processBatch("edge", batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				bp.processBatch("edge", batch)
				batch = batch[:0]
			}
		}
	}
}

func (bp *BatchProcessor) prefixWorker() {
	defer bp.wg.Done()

	batch := make([]*BatchOperation, 0, bp.batchSize)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bp.stop:
			if len(batch) > 0 {
				bp.processBatch("prefix", batch)
			}
			return

		case op := <-bp.prefixOps:
			batch = append(batch, op)

			if len(batch) >= bp.batchSize {
				bp.processBatch("prefix", batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				bp.processBatch("prefix", batch)
				batch = batch[:0]
			}
		}
	}
}

func (bp *BatchProcessor) processBatch(batchType string, batch []*BatchOperation) {
	if len(batch) == 0 {
		return
	}

	ctx := context.TODO()
	processed := 0

	for _, op := range batch {
		if err := bp.processOperation(ctx, op); err != nil {
			glog.Errorf("Failed to process %s operation %s: %v", batchType, op.Key, err)
		} else {
			processed++
		}

		bp.pending.Add(-1)
		bp.processed.Add(1)
	}

	glog.V(7).Infof("Processed %d/%d %s operations in batch", processed, len(batch), batchType)
}

func (bp *BatchProcessor) processOperation(ctx context.Context, op *BatchOperation) error {
	// TODO: Implement actual operation processing
	// This will be expanded when we implement specific BGP processing logic
	glog.V(9).Infof("Processing %s operation: %s %s", op.Type, op.Action, op.Key)
	return nil
}

func (bp *BatchProcessor) flushTimer() {
	defer bp.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bp.stop:
			return
		case <-ticker.C:
			// Periodic flush statistics logging
			stats := bp.GetStats()
			glog.V(6).Infof("Batch processor stats: processed=%d, pending=%d",
				stats.Processed.Load(), stats.Pending.Load())
		}
	}
}
