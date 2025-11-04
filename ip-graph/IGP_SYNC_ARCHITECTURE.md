# IGP Sync Architecture

## Overview

The `ip-graph` processor now uses a **producer-consumer pattern** for IGP topology synchronization, eliminating race conditions and code duplication.

## Architecture

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│ GoBMP       │ Kafka   │  igp-graph   │ ArangoDB │  ip-graph   │
│ Collector   ├────────►│  Processor   ├─────────►│  Processor  │
└─────────────┘ (IGP)   └──────────────┘          └─────────────┘
                            │                          ▲
                            │ Writes to:               │ Syncs from:
                            │ - igpv4_graph            │ - igpv4_graph
                            │ - igpv6_graph            │ - igpv6_graph
                            └──────────────────────────┘
```

## Components

### 1. Kafka Topic Subscription

**`ip-graph` subscribes to:**
- `gobmp.parsed.peer` - BGP peer sessions
- `gobmp.parsed.unicast_prefix_v4` - BGP IPv4 prefixes  
- `gobmp.parsed.unicast_prefix_v6` - BGP IPv6 prefixes

**`ip-graph` does NOT subscribe to:**
- ~~`gobmp.parsed.ls_node`~~ - Handled by `igp-graph`
- ~~`gobmp.parsed.ls_link`~~ - Handled by `igp-graph`
- ~~`gobmp.parsed.ls_prefix`~~ - Handled by `igp-graph`
- ~~`gobmp.parsed.ls_srv6_sid`~~ - Handled by `igp-graph`

### 2. IGP Sync Strategy

**Initial Sync (Startup):**
- Runs once when `ip-graph` starts
- Bulk copies all edges from `igpv4_graph` → `ipv4_graph`
- Bulk copies all edges from `igpv6_graph` → `ipv6_graph`
- Fast and efficient (single AQL query per graph)

**Periodic Reconciliation:**
- Runs every 10 seconds (configurable)
- Finds edges in `igpv4_graph`/`igpv6_graph` that don't exist in `ipv4_graph`/`ipv6_graph`
- Copies missing edges
- Self-healing: catches any missed updates
- Logs when edges are synced

### 3. Implementation Details

**Files Modified:**

1. **`kafkamessenger/kafkamessenger.go`**
   - Removed IGP topics from subscription list
   - Removed IGP message type handling

2. **`arangodb/igp-sync-processor.go`**
   - Added `StartReconciliation()` / `StopReconciliation()` methods
   - Added `reconcileIPv4Edges()` / `reconcileIPv6Edges()` methods
   - Uses efficient AQL queries to find and sync missing edges

3. **`arangodb/arangodb.go`**
   - Added `igpSyncProcessor` field to struct
   - Starts reconciliation in `Start()` method
   - Stops reconciliation in `Stop()` method

## Advantages

✅ **No Race Conditions**: `igp-graph` finishes processing before `ip-graph` syncs

✅ **No Code Duplication**: IGP processing logic lives only in `igp-graph`

✅ **Single Source of Truth**: `igpv4_graph`/`igpv6_graph` are authoritative

✅ **Self-Healing**: Periodic reconciliation catches missed updates

✅ **Independent Scaling**: Each processor can scale independently

✅ **Easier Maintenance**: Bug fixes only needed in one place

✅ **Clean Separation**: Each processor has distinct responsibilities

## Performance Characteristics

**IGP Update Latency:**
- Maximum delay: 10 seconds (reconciliation interval)
- Typical delay: 5-10 seconds
- Acceptable because IGP changes are infrequent

**BGP Update Latency:**
- Real-time (no change)
- Processed immediately from Kafka

**Resource Usage:**
- Reconciliation query runs every 10 seconds
- Efficient: Only copies missing edges
- Minimal overhead when topology is stable

## Configuration

**Reconciliation Interval:**
```go
// In igp-sync-processor.go NewIGPSyncProcessor()
interval: 10 * time.Second  // Adjust as needed
```

**Fast Interval (5s)**: More responsive, higher DB load
**Slow Interval (30s)**: Less responsive, lower DB load
**Recommended (10s)**: Good balance for most deployments

## Monitoring

**Log Messages to Watch:**

```
INFO: IGP topology reconciliation started
INFO: Starting IGP topology reconciliation (interval: 10s)
V(6): Reconciled N missing IPv4 IGP edges
V(6): Reconciled N missing IPv6 IGP edges
INFO: IGP topology reconciliation stopped
```

**If you see many reconciled edges:**
- Check if `igp-graph` is running
- Check if `igp-graph` is processing messages
- May indicate `igp-graph` backlog

**If you see zero reconciled edges:**
- Normal during steady state
- Indicates sync is working correctly

## Future Enhancements

**Possible improvements:**

1. **Event-Driven Sync**: 
   - Have `igp-graph` publish "topology changed" events
   - Trigger immediate reconciliation on events
   - Fallback to periodic reconciliation

2. **Adaptive Interval**:
   - Fast interval (5s) when changes detected
   - Slow interval (30s) during steady state

3. **Deletion Sync**:
   - Currently only syncs additions
   - Could add deletion reconciliation (remove edges not in IGP graphs)

4. **Metrics**:
   - Track edges synced per cycle
   - Track reconciliation duration
   - Export Prometheus metrics

## Troubleshooting

**Problem: IGP edges not appearing in `ipv4_graph`/`ipv6_graph`**

Solutions:
1. Check if `igp-graph` is running and processing
2. Check if edges exist in `igpv4_graph`/`igpv6_graph` 
3. Check `ip-graph` logs for reconciliation messages
4. Verify collections exist and are accessible

**Problem: Stale IGP topology**

Solutions:
1. Restart `ip-graph` to force initial sync
2. Manually trigger reconciliation (restart processor)
3. Reduce reconciliation interval for faster updates

**Problem: High database load**

Solutions:
1. Increase reconciliation interval (e.g., 30s)
2. Add indexes on `_key` fields if not present
3. Consider event-driven sync instead of polling

## Testing

**Verify Initial Sync:**
```aql
// Check edge counts match
RETURN {
  igpv4: LENGTH(igpv4_graph),
  ipv4: LENGTH(FOR e IN ipv4_graph FILTER e.protocol_id != null RETURN 1),
  igpv6: LENGTH(igpv6_graph),
  ipv6: LENGTH(FOR e IN ipv6_graph FILTER e.protocol_id != null RETURN 1)
}
```

**Verify Reconciliation:**
1. Stop `ip-graph`
2. Have `igp-graph` process some IGP changes
3. Start `ip-graph`
4. Wait 10-15 seconds
5. Verify edges appeared in `ipv4_graph`/`ipv6_graph`

## Migration Notes

**Upgrading from Previous Version:**

1. Deploy new `ip-graph` code
2. Restart `ip-graph` processor
3. Initial sync will populate missing edges
4. No database migration needed
5. Old IGP routing code is harmless (dead code)

**Rollback:**

1. Revert code changes
2. Restart `ip-graph`
3. Processor will re-subscribe to IGP topics
4. Both sync methods will coexist (redundant but safe)

