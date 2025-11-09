# Real-Time BMP Message Update Implementation

## Overview
This document describes how ip-graph handles real-time BMP message updates for IGP topology changes, BGP peer sessions, and BGP prefix advertisements/withdrawals.

## Architecture

### Division of Responsibilities

**igp-graph processor:**
- Processes raw BGP-LS messages (ls_node, ls_link, ls_prefix, ls_srv6_sid)
- Maintains `igp_node` collection
- Maintains `igpv4_graph` and `igpv6_graph` edge collections
- Handles IGP topology updates in real-time

**ip-graph processor:**
- **SYNCS** IGP topology from igpv4_graph/igpv6_graph to ipv4_graph/ipv6_graph
- Processes BGP peer state changes
- Processes BGP unicast prefix advertisements/withdrawals
- Maintains full topology graphs (IGP + BGP)

### Key Principle: Don't Re-Process IGP Data

ip-graph does **NOT** re-process raw BGP-LS messages. Instead, it:
1. **Syncs** already-processed IGP edges from igpv4_graph/igpv6_graph
2. Adds BGP overlay (peers, prefixes) on top of IGP underlay
3. Creates unified full topology in ipv4_graph/ipv6_graph

## Components

### 1. IGPSyncProcessor (`igp-sync-processor.go`)

**Purpose:** Syncs IGP topology changes from igp-graph's output to ip-graph's full topology.

**Key Methods:**

#### `syncIGPLinkUpdate(ctx, linkKey, action, isIPv4)`
Syncs a single IGP link (edge) change.

**Actions:**
- `add`/`update`: Copies edge from igpvX_graph → ipvX_graph
- `del`: Removes edge from ipvX_graph

**Implementation:**
```go
// For add/update:
1. Read edge from igpv4_graph or igpv6_graph
2. Remove ArangoDB metadata (_id, _rev)
3. Insert or update in ipv4_graph or ipv6_graph

// For delete:
1. Remove edge from ipv4_graph or ipv6_graph
```

#### `InitialIGPSync(ctx)`
Performs bulk sync during startup.

**Implementation:**
```aql
FOR edge IN igpv4_graph
INSERT UNSET(edge, "_id", "_rev") INTO ipv4_graph
OPTIONS { overwriteMode: "update" }
```

Does the same for IPv6.

#### `syncIGPNodeUpdate(ctx, nodeKey, action)`
Handles IGP node updates. Currently a no-op because:
- igp_node collection is shared between processors
- Nodes are referenced by edges, not copied
- igp-graph maintains the node collection

### 2. UpdateCoordinator (`update-coordinator.go`)

**Purpose:** Routes incoming BMP messages to appropriate processors.

**Message Flow:**

```
BMP Message → StoreMessage() → ProcessMessage()
                                       ↓
                    ┌──────────────────┼──────────────────┐
                    ↓                  ↓                  ↓
              IGP Updates        BGP Peer           BGP Prefix
                    ↓            Updates              Updates
           igpUpdateWorker()        ↓                    ↓
                    ↓          bgpUpdateWorker()   prefixUpdateWorker()
                    ↓                  ↓                  ↓
         processIGPUpdate()   processBGPUpdate()  processPrefixUpdate()
                    ↓                  ↓                  ↓
          IGPSyncProcessor     BGPPeerProcessor    BGPPrefixProcessor
```

#### Message Type Routing

**IGP Messages** (sync from igp-graph):
- `bmp.LSNodeMsg` → syncIGPNodeUpdate()
- `bmp.LSLinkMsg` → syncIGPLinkUpdate()
- `bmp.LSPrefixMsg` → syncIGPPrefixUpdate()
- `bmp.LSSRv6SIDMsg` → syncIGPSRv6Update()

**BGP Messages** (process directly):
- `bmp.PeerStateChangeMsg` → processBGPPeerUpdate()
- `bmp.UnicastPrefixV4Msg` → processBGPPrefixUpdate()
- `bmp.UnicastPrefixV6Msg` → processBGPPrefixUpdate()

### 3. BGP Peer Processor (`bgp-peer-processor.go`)

**Purpose:** Handles real-time BGP peer session changes.

**Actions:**
- `add`/`update`: Creates/updates BGP nodes and session edges
- `del`: Removes session edges (nodes kept if they have other sessions)

**Key Features:**
- Creates ONE edge per peer message
- Edge key format: `remote_bgp_id_remote_asn_remote_ip`
- Natural bidirectionality from BMP sending messages from both routers

### 4. BGP Prefix Processor (`bgp-prefix-processor.go`)

**Purpose:** Handles real-time BGP prefix advertisements and withdrawals.

**Actions:**
- `add`/`update`: Classifies prefix, creates vertex and edges
- `del`: Removes prefix vertex and associated edges

**Key Features:**
- Best path selection (checks if new path is better)
- Prefix classification (iBGP, eBGP private, eBGP public)
- Loopback detection (/32, /128 as node metadata)
- Prefix-to-node attachment

## Real-Time Update Examples

### Example 1: IGP Link Add

**Scenario:** New link between routers in IGP domain

```
1. BMP sends ls_link message
2. igp-graph processes it → creates edge in igpv4_graph
3. ip-graph receives ls_link notification
4. IGPSyncProcessor reads edge from igpv4_graph
5. IGPSyncProcessor writes edge to ipv4_graph
```

**Result:** New IGP link appears in full topology graph

### Example 2: IGP Link Delete

**Scenario:** Link failure in IGP domain

```
1. BMP sends ls_link delete message
2. igp-graph processes it → removes edge from igpv4_graph
3. ip-graph receives ls_link delete notification
4. IGPSyncProcessor removes edge from ipv4_graph
```

**Result:** Failed IGP link removed from full topology graph

### Example 3: BGP Peer Session Add

**Scenario:** New BGP peer session established

```
1. BMP sends peer state change message (action: add)
2. ip-graph receives peer message
3. BGPPeerProcessor:
   a. Creates bgp_node for remote peer (if not in IGP domain)
   b. Creates session edge with key: remote_bgp_id_asn_ip
4. BMP sends reverse peer message from other router
5. BGPPeerProcessor creates reverse direction edge
```

**Result:** Bidirectional BGP session in topology

### Example 4: BGP Prefix Advertisement

**Scenario:** New BGP prefix announced

```
1. BMP sends unicast_prefix message (action: add)
2. ip-graph receives prefix message
3. BGPPrefixProcessor:
   a. Classifies prefix type (iBGP/eBGP/public)
   b. Creates or updates prefix in bgp_prefix_v4
   c. Creates edges from advertising node to prefix
   d. Creates edges from prefix to next-hop nodes
```

**Result:** New prefix vertex in topology with appropriate edges

### Example 5: BGP Prefix Withdrawal

**Scenario:** BGP prefix withdrawn

```
1. BMP sends unicast_prefix message (action: del)
2. ip-graph receives prefix message
3. BGPPrefixProcessor:
   a. Identifies prefix vertex in bgp_prefix_v4
   b. Removes all edges to/from prefix
   c. Removes prefix vertex
```

**Result:** Prefix and associated edges removed from topology

## Initial Loading vs Real-Time Updates

### Initial Loading (Startup)

**Sequence:**
1. Initialize collections and graphs
2. **Bulk IGP Sync:** Copy all edges from igpv4_graph/igpv6_graph
3. **Bulk BGP Node Creation:** Query peer collection, create bgp_nodes
4. **Bulk BGP Prefix Classification:** Classify all unicast prefixes
5. **Bulk Prefix Attachment:** Create prefix-to-node edges
6. **Apply BGP Precedence:** Remove conflicting IGP edges
7. Start real-time update coordinators

**Advantages:**
- Fast bulk operations using AQL queries
- Efficient initial state setup
- Handles large topologies well

### Real-Time Updates (Running)

**Sequence:**
1. Receive BMP message from Kafka
2. Parse and route to appropriate worker
3. Process individual add/update/delete
4. Update specific vertices/edges
5. Apply conflict resolution if needed

**Advantages:**
- Low latency updates
- Incremental changes
- Maintains consistency

## Configuration

### Channel Buffer Sizes

```go
igpUpdates:    make(chan *ProcessingMessage, 1000),
bgpUpdates:    make(chan *ProcessingMessage, 1000),
prefixUpdates: make(chan *ProcessingMessage, 1000),
```

**Tuning Considerations:**
- Increase for high-rate topologies
- Monitor queue depth
- Balance memory vs message loss

### Worker Concurrency

Currently: 3 workers (IGP, BGP peer, BGP prefix)

**Potential Enhancements:**
- Multiple workers per message type
- Priority queuing for critical updates
- Batch processing for high-volume updates

## Error Handling

### Graceful Degradation

**Principle:** Log errors but continue processing

**Examples:**

1. **IGP Sync Failure:**
   - Log warning
   - Continue with BGP-only topology
   - Topology remains consistent (no partial state)

2. **BGP Node Creation Failure:**
   - Log error
   - Skip this peer session
   - Process other sessions normally

3. **Prefix Processing Failure:**
   - Log error
   - Skip this prefix
   - Process other prefixes normally

### Conflict Resolution

**IGP-BGP Prefix Conflicts:**
- BGP takes precedence over IGP
- Remove IGP edges for conflicting prefixes
- Maintain iBGP internal redistribution

## Monitoring and Debugging

### Log Levels

- `V(5)`: High-level operations
- `V(6)`: Worker status, bulk operations
- `V(7)`: Individual message processing
- `V(8)`: Detailed processing steps
- `V(9)`: Very detailed (edge creation, queries)

### Key Metrics to Monitor

1. **Queue Depths:**
   - igpUpdates channel utilization
   - bgpUpdates channel utilization
   - prefixUpdates channel utilization

2. **Processing Rates:**
   - Messages/second per worker
   - Update latency (message → DB)

3. **Error Rates:**
   - Failed IGP syncs
   - Failed BGP peer creations
   - Failed prefix classifications

### Debugging Commands

**Check IGP Sync Status:**
```aql
// Compare edge counts
RETURN {
  igpv4_count: LENGTH(igpv4_graph),
  ipv4_count: LENGTH(ipv4_graph),
  igpv6_count: LENGTH(igpv6_graph),
  ipv6_count: LENGTH(ipv6_graph)
}
```

**Check BGP Session Edges:**
```aql
FOR edge IN ipv4_graph
FILTER edge.protocol LIKE "BGP_%"
COLLECT protocol = edge.protocol WITH COUNT INTO count
RETURN { protocol, count }
```

**Check Prefix Distribution:**
```aql
FOR prefix IN bgp_prefix_v4
COLLECT type = prefix.prefix_type WITH COUNT INTO count
RETURN { type, count }
```

## Comparison to Original Implementation

### Original ipv4-graph/ipv6-graph

**Limitations:**
- Only handled peer and unicast prefix updates
- No IGP sync (IGP data static after startup)
- Basic error handling
- No conflict resolution
- Hardcoded collection names

### New ip-graph

**Improvements:**
- ✅ Full IGP sync from igp-graph processor
- ✅ Comprehensive update handling (IGP + BGP)
- ✅ Proper error handling and graceful degradation
- ✅ Conflict resolution (BGP precedence)
- ✅ Configurable collection names
- ✅ Better logging and monitoring
- ✅ Unified IPv4 and IPv6 processing

## Future Enhancements

### Potential Improvements

1. **Batch Processing:**
   - Accumulate multiple updates
   - Process in batches for efficiency
   - Reduce DB round trips

2. **Update Prioritization:**
   - Critical updates first (link failures)
   - Bulk updates for low priority (metrics)
   - QoS for different message types

3. **State Reconciliation:**
   - Periodic full sync with IGP topology
   - Detect and repair inconsistencies
   - Handle processor restarts gracefully

4. **Performance Optimization:**
   - Connection pooling for DB
   - Prepared statements for common queries
   - Caching for frequently accessed data

5. **Enhanced Monitoring:**
   - Prometheus metrics export
   - Grafana dashboards
   - Alert on queue buildup or errors

## Testing Recommendations

### Functional Testing

1. **IGP Sync:**
   - Add link in IGP → verify in full topology
   - Delete link in IGP → verify removed
   - Modify link metrics → verify updated

2. **BGP Peer:**
   - Establish session → verify edges created
   - Tear down session → verify edges removed
   - Session flap → verify stability

3. **BGP Prefix:**
   - Advertise prefix → verify vertex and edges
   - Withdraw prefix → verify removed
   - Update attributes → verify changed

### Stress Testing

1. **High Message Rate:**
   - 1000+ messages/second
   - Monitor queue depths
   - Check for message loss

2. **Large Topology:**
   - 1000+ nodes
   - 10000+ links
   - 100000+ prefixes

3. **Rapid Changes:**
   - Link flaps
   - BGP route churn
   - Verify eventual consistency

---
*Document Created: 2025-11-01*
*Version: 1.0*
*Author: Automated Code Review with Bruce McDougall*

