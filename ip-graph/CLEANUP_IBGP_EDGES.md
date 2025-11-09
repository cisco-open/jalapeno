# Cleanup Script for Stale iBGP Edges

## Issue

After a fresh install, stale iBGP edges with old key formats (containing `_rev`/`_fwd` suffixes) remain in the `ipv4_graph` and `ipv6_graph` collections. These were created by older versions of the code before the iBGP session filter was implemented.

## Example Stale Edge

```json
{
    "_key": "10.0.0.9_10.0.0.9_rev",
    "_id": "ipv4_graph/10.0.0.9_10.0.0.9_rev",
    "_from": "igp_node/2_0_0_0000.0000.0009",
    "_to": "igp_node/2_0_0_0000.0000.0025",
    "protocol": "BGP_ibgp",
    "local_node_asn": 100000,
    "remote_node_asn": 100000
}
```

**Characteristics of stale edges:**
- `protocol` == `"BGP_ibgp"`
- `_key` contains `_rev` or `_fwd` suffix
- `local_node_asn` == `remote_node_asn`

## Cleanup Queries

### Remove All iBGP Edges from IPv4 Graph

```aql
FOR edge IN ipv4_graph
FILTER edge.protocol == "BGP_ibgp"
REMOVE edge IN ipv4_graph
RETURN OLD
```

### Remove All iBGP Edges from IPv6 Graph

```aql
FOR edge IN ipv6_graph
FILTER edge.protocol == "BGP_ibgp"
REMOVE edge IN ipv6_graph
RETURN OLD
```

### Check for Remaining iBGP Edges (Should return 0)

```aql
RETURN {
  ipv4_count: LENGTH(
    FOR edge IN ipv4_graph
    FILTER edge.protocol == "BGP_ibgp"
    RETURN 1
  ),
  ipv6_count: LENGTH(
    FOR edge IN ipv6_graph
    FILTER edge.protocol == "BGP_ibgp"
    RETURN 1
  )
}
```

## Verification

### 1. Verify iBGP Sessions Are Not Creating New Edges

After cleanup, restart `ip-graph` and check that no new iBGP edges are created:

```aql
// Wait 5 minutes after restart, then run:
FOR edge IN ipv4_graph
FILTER edge.protocol == "BGP_ibgp"
RETURN COUNT
```

**Expected result:** `0`

### 2. Verify eBGP Edges Still Exist

```aql
FOR edge IN ipv4_graph
FILTER edge.protocol IN ["BGP_ebgp_private", "BGP_ebgp_public", "BGP_ebgp_hybrid"]
COLLECT protocol = edge.protocol WITH COUNT INTO count
RETURN { protocol, count }
```

**Expected result:** Should show counts for eBGP session types

### 3. Verify IGP Edges Are Intact

```aql
FOR edge IN ipv4_graph
FILTER edge.protocol_id != null
LIMIT 10
RETURN {
  key: edge._key,
  from: edge._from,
  to: edge._to,
  protocol: edge.protocol
}
```

**Expected result:** Should show IGP edges (ISIS, OSPF, etc.)

## Prevention

The current code (post-fix) includes multiple layers of iBGP edge prevention:

### 1. BGP Peer Processor Filter
File: `bgp-peer-processor.go`

```go
// Skip iBGP sessions - they run over IGP infrastructure, no separate edges needed
if sessionType == "ibgp" {
    glog.V(7).Infof("Skipping edge creation for iBGP session %s (uses IGP connectivity)", key)
    return nil
}
```

This filter applies to:
- Initial peer loading (`loadInitialPeers`)
- Real-time peer updates (`processBGPPeerUpdate`)

### 2. IGP Sync Processor Filter
File: `igp-sync-processor.go`

Added safety filters to prevent any BGP edges from being synced from IGP graphs:

```go
// Initial sync
FOR edge IN igpv4_graph
FILTER edge.protocol_id != null OR edge.protocol NOT LIKE "BGP_%"
INSERT UNSET(edge, "_id", "_rev") INTO ipv4_graph
```

```go
// Periodic reconciliation
FOR igp_edge IN igpv4_graph
FILTER igp_edge.protocol_id != null OR igp_edge.protocol NOT LIKE "BGP_%"
LET exists = (...)
FILTER LENGTH(exists) == 0
INSERT UNSET(igp_edge, "_id", "_rev") INTO ipv4_graph
```

**Result:** No new iBGP edges should be created going forward, with multiple safety layers.

## When to Run Cleanup

Run the cleanup queries if:
1. You upgraded from an older version of `ip-graph`
2. You see iBGP edges in the graph collections
3. Graph queries show unexpected `BGP_ibgp` protocol entries
4. Edge keys contain old `_fwd`/`_rev` suffixes

## Full Cleanup Script

```bash
#!/bin/bash
# cleanup_ibgp_edges.sh

# ArangoDB connection details
ARANGO_URL="http://localhost:8529"
DB_NAME="jalapeno"
USER="root"
PASSWORD="jalapeno"

echo "Cleaning up iBGP edges from IP graphs..."

# Remove iBGP edges from IPv4 graph
echo "Removing iBGP edges from ipv4_graph..."
curl -X POST "${ARANGO_URL}/_db/${DB_NAME}/_api/cursor" \
  -u "${USER}:${PASSWORD}" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "FOR edge IN ipv4_graph FILTER edge.protocol == \"BGP_ibgp\" REMOVE edge IN ipv4_graph RETURN OLD"
  }'

# Remove iBGP edges from IPv6 graph
echo "Removing iBGP edges from ipv6_graph..."
curl -X POST "${ARANGO_URL}/_db/${DB_NAME}/_api/cursor" \
  -u "${USER}:${PASSWORD}" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "FOR edge IN ipv6_graph FILTER edge.protocol == \"BGP_ibgp\" REMOVE edge IN ipv6_graph RETURN OLD"
  }'

# Verify cleanup
echo "Verifying cleanup..."
curl -X POST "${ARANGO_URL}/_db/${DB_NAME}/_api/cursor" \
  -u "${USER}:${PASSWORD}" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "RETURN { ipv4_count: LENGTH(FOR edge IN ipv4_graph FILTER edge.protocol == \"BGP_ibgp\" RETURN 1), ipv6_count: LENGTH(FOR edge IN ipv6_graph FILTER edge.protocol == \"BGP_ibgp\" RETURN 1) }"
  }'

echo "Cleanup complete!"
```

## Alternative: Fresh Install

For a truly fresh install, clear all graph collections before starting `ip-graph`:

```aql
// WARNING: This removes ALL edges, including IGP data
FOR edge IN ipv4_graph REMOVE edge IN ipv4_graph
FOR edge IN ipv6_graph REMOVE edge IN ipv6_graph
FOR node IN bgp_node REMOVE node IN bgp_node
FOR prefix IN bgp_prefix_v4 REMOVE prefix IN bgp_prefix_v4
FOR prefix IN bgp_prefix_v6 REMOVE prefix IN bgp_prefix_v6
```

Then restart both `igp-graph` and `ip-graph` to rebuild from scratch.

---
*Document Created: 2025-11-08*
*Issue: Stale iBGP edges from pre-fix code versions*

