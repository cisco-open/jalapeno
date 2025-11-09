# iBGP Session Edge Filtering Fix

## Issue

The processor was creating graph edges for **iBGP sessions** (sessions between routers in the same AS), which duplicated existing IGP connectivity in the topology graph.

### Why This Was Wrong

**iBGP sessions run OVER the IGP infrastructure:**
- iBGP sessions (e.g., to route reflectors) use the existing IGP paths
- Creating separate graph edges for iBGP duplicates the physical/logical connectivity
- The IGP edges already represent all intra-AS connectivity

**Example:**
```
Router A (AS 100) --[IGP link]--> Router B (AS 100)
Router A (AS 100) --[iBGP session]--> Router B (AS 100, Route Reflector)
```

In this case:
- The **IGP link** edge is needed ✅ (represents physical connectivity)
- The **iBGP session** edge is NOT needed ❌ (runs over the IGP link)

## Fix Applied

### Change Location
File: `/ip-graph/arangodb/bgp-peer-processor.go`
Function: `processPeerAddUpdate()`

### Code Change

**Added iBGP session filter:**

```go
// Determine session type
sessionType := uc.classifyBGPSession(localASN, remoteASN)

glog.V(8).Infof("Processing %s session: %s (AS%d) ↔ %s (AS%d)",
    sessionType, localBGPID, localASN, remoteBGPID, remoteASN)

// Skip iBGP sessions - they run over IGP infrastructure, no separate edges needed
if sessionType == "ibgp" {
    glog.V(7).Infof("Skipping edge creation for iBGP session %s (uses IGP connectivity)", key)
    return nil
}

// Only process eBGP sessions (continue with node and edge creation)
```

## Session Type Classification

The `classifyBGPSession()` function identifies session types:

### iBGP Session (SKIPPED)
- `localASN == remoteASN`
- Examples:
  - AS 100 ↔ AS 100 (route reflector)
  - AS 65000 ↔ AS 65000 (full mesh iBGP)
- **Action:** Skip edge creation (uses IGP)

### eBGP Private Session (PROCESSED)
- `localASN != remoteASN`
- Both ASNs are private (64512-65535 or 4200000000-4294967294)
- Examples:
  - AS 65001 ↔ AS 65002
  - AS 4200000001 ↔ AS 4200000002
- **Action:** Create edges and bgp_nodes

### eBGP Public Session (PROCESSED)
- `localASN != remoteASN`
- Both ASNs are public
- Examples:
  - AS 100 ↔ AS 200
  - AS 15169 (Google) ↔ AS 7018 (AT&T)
- **Action:** Create edges and bgp_nodes

### eBGP Hybrid Session (PROCESSED)
- `localASN != remoteASN`
- One ASN private, one public
- Examples:
  - AS 65000 ↔ AS 100
  - Edge/transit connection
- **Action:** Create edges and bgp_nodes

## Impact

### Before Fix
```aql
// Example topology with iBGP sessions creating duplicate edges
FOR edge IN ipv4_graph
FILTER edge.protocol == "BGP_ibgp"
RETURN {
  from: edge._from,
  to: edge._to,
  local_asn: edge.local_node_asn,
  remote_asn: edge.remote_node_asn
}
```
Would return many results - iBGP edges cluttering the graph ❌

### After Fix
```aql
// Same query after fix
FOR edge IN ipv4_graph
FILTER edge.protocol == "BGP_ibgp"
RETURN COUNT
```
Returns 0 - no iBGP edges in graph ✅

### What's Now in the Graph

**Only eBGP session edges are created:**
```aql
FOR edge IN ipv4_graph
FILTER edge.protocol LIKE "BGP_%"
COLLECT protocol = edge.protocol WITH COUNT INTO count
RETURN { protocol, count }
```

Expected results:
- `BGP_ebgp_private`: X edges (private AS interconnections)
- `BGP_ebgp_public`: Y edges (Internet peering)
- `BGP_ebgp_hybrid`: Z edges (edge connections)
- `BGP_ibgp`: 0 edges ✅

## Verification

### Check for Remaining iBGP Edges

```aql
// Should return 0
RETURN LENGTH(
  FOR edge IN ipv4_graph
  FILTER edge.protocol == "BGP_ibgp"
  RETURN edge
)
```

### Check iBGP Sessions in Peer Collection

```aql
// iBGP sessions still exist in peer collection (for BGP prefix processing)
FOR p IN peer
FILTER p.local_asn == p.remote_asn
RETURN {
  key: p._key,
  local: p.local_bgp_id,
  remote: p.remote_bgp_id,
  asn: p.local_asn
}
```
This should return results - iBGP sessions are **tracked** but **not graphed** ✅

### Verify eBGP Edges Exist

```aql
// Should have edges for eBGP sessions only
FOR edge IN ipv4_graph
FILTER edge.protocol IN ["BGP_ebgp_private", "BGP_ebgp_public", "BGP_ebgp_hybrid"]
FILTER edge.local_node_asn != edge.remote_node_asn
RETURN {
  key: edge._key,
  from: edge._from,
  to: edge._to,
  local_asn: edge.local_node_asn,
  remote_asn: edge.remote_node_asn
}
```

## Topology Interpretation

### How to Read the Full Topology Graph

**For Intra-AS Connectivity (within same AS):**
- Use **IGP edges** (ISIS, OSPF)
- Protocol field: `"isis"`, `"ospf"`, etc.
- Source: `igpv4_graph` → `ipv4_graph`

**For Inter-AS Connectivity (between different ASes):**
- Use **eBGP edges**
- Protocol field: `"BGP_ebgp_private"`, `"BGP_ebgp_public"`, `"BGP_ebgp_hybrid"`
- Source: BGP peer messages

**For BGP Route Distribution:**
- Query `bgp_prefix_v4` and `bgp_prefix_v6` collections
- These show prefix advertisements/reachability
- Not dependent on session edges

## Why iBGP Sessions Still Matter

Even though we don't create edges for iBGP sessions, they're still important:

1. **Prefix Distribution:**
   - iBGP carries BGP prefixes within the AS
   - Used for prefix classification (iBGP vs eBGP prefixes)
   - Tracked in `bgp_prefix_v4`/`bgp_prefix_v6`

2. **Topology Understanding:**
   - Route reflector relationships
   - BGP policy points
   - Available in `peer` collection for querying

3. **Network Analysis:**
   - Can correlate IGP paths with iBGP sessions
   - Understand route distribution patterns
   - Troubleshoot control plane issues

## Related Documents

- `/ip-graph/BGP_NODE_FIXES.md` - BGP node processing fixes
- `/ip-graph/BGP_PREFIX_FIXES.md` - BGP prefix classification fixes  
- `/ip-graph/BGP_EDGE_DUPLICATE_FIX.md` - BGP edge duplicate fix
- `/ip-graph/REALTIME_UPDATE_IMPLEMENTATION.md` - Real-time update handling

---
*Document Created: 2025-11-01*
*Issue Reported By: Bruce McDougall*
*Fixed By: Automated Code Review*

