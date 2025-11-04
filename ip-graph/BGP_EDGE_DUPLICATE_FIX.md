# BGP Session Edge Duplicate Fix - Summary

## Overview
This document summarizes the fix for duplicate BGP session edges being created in the ipv4_graph and ipv6_graph collections.

## Issue Reported

**Problem:** Duplicate edges were being created for BGP peer sessions with incorrect key formats.

### Example of Duplicates Found:

**Good Edge (needed):**
```json
{
  "_key": "10.0.0.43_65042_10.2.1.11",
  "_from": "bgp_node/10.0.0.41_65041",
  "_to": "bgp_node/10.0.0.43_65042",
  "protocol": "BGP_ebgp_private",
  "local_ip": "10.2.1.10",
  "remote_ip": "10.2.1.11"
}
```

**Bad Edges (duplicates, not needed):**
```json
{
  "_key": "10.0.0.41_10.2.1.10_rev",     // Wrong: has _rev suffix
  "_key": "10.0.0.43_10.2.1.11_fwd",     // Wrong: has _fwd suffix  
  "_key": "10.0.0.41_10.2.1.10_fwd",     // Wrong: has _fwd suffix
  "_key": "10.0.0.43_10.2.1.11_rev"      // Wrong: has _rev suffix
}
```

**Key Observations:**
- Good edges use format: `router_id_asn_ip`
- Bad edges use format: `router_id_ip_fwd` or `router_id_ip_rev`
- Multiple edges exist between the same two nodes
- Only ONE edge per direction should exist

## Root Cause Analysis

### Issue: Creating Multiple Edges Per Peer Message

The problem was in the `createBGPSessionEdges` function which was trying to be "smart" about bidirectional edge creation:

**Previous Logic (WRONG):**
```go
// Check if this is an eBGP session between IGP node and external peer
localInIGP, _ := uc.checkIGPPeerASNExists(ctx, localASN)
remoteInIGP, _ := uc.checkIGPPeerASNExists(ctx, remoteASN)

if localInIGP && !remoteInIGP {
    // Create bidirectional edges (2 edges)
    createSessionEdge(..., true)
    createSessionEdge(..., false)
} else {
    // Create single edge (1 edge)
    createSessionEdge(..., true)
}
```

This logic had several problems:
1. It created multiple edges per peer message
2. It used `_fwd` and `_rev` suffixes for edge keys
3. BMP sends peer messages from BOTH routers, so this doubled the edges

### How BMP Provides Bidirectionality

BMP naturally provides bidirectional session information by sending **TWO separate peer messages**:

**Example BGP Session between Router A and Router B:**

**Message 1 (from Router A's perspective):**
- `local_bgp_id`: A's router ID
- `remote_bgp_id`: B's router ID
- `local_asn`: A's ASN
- `remote_asn`: B's ASN
- `local_ip`: A's IP
- `remote_ip`: B's IP

**Message 2 (from Router B's perspective):**
- `local_bgp_id`: B's router ID
- `remote_bgp_id`: A's router ID
- `local_asn`: B's ASN
- `remote_asn`: A's ASN
- `local_ip`: B's IP
- `remote_ip`: A's IP

### Original Code Logic (CORRECT)

The original ipv4-graph and ipv6-graph code was simpler and correct:

```go
// Each peer message creates ONE edge
// Edge key: remote_bgp_id + remote_asn + remote_ip
// From: local node
// To: remote node
```

**Result:**
- Message 1 creates edge: A → B with key based on B's info
- Message 2 creates edge: B → A with key based on A's info
- Total: 2 edges (one per direction) ✅

## Fix Applied

### 1. Simplified createBGPSessionEdges()

**New Logic:**
```go
func (uc *UpdateCoordinator) createBGPSessionEdges(...) error {
    // Get node IDs
    localNodeID := getBGPNodeID(localBGPID, localASN, localIP)
    remoteNodeID := getBGPNodeID(remoteBGPID, remoteASN, remoteIP)
    
    // Create ONE edge per peer message
    // BMP sends peer messages from both routers, providing natural bidirectionality
    createSessionEdge(sessionKey, localNodeID, remoteNodeID, peerData, sessionType)
    
    return nil
}
```

**Key Changes:**
- ✅ Removed conditional logic for IGP vs non-IGP
- ✅ Create exactly ONE edge per peer message
- ✅ Rely on BMP to provide bidirectionality naturally

### 2. Simplified createSessionEdge()

**Changed Function Signature:**
```go
// Before:
func createSessionEdge(..., isForward bool) error

// After:  
func createSessionEdge(...) error  // Removed isForward parameter
```

**Key Format:**
```go
// Edge key uses REMOTE peer's information
edgeKey := fmt.Sprintf("%s_%d_%s", remoteBGPID, remoteASN, remoteIP)
```

This ensures:
- Each peer message creates a unique edge key
- Keys are based on the remote (destination) peer
- No `_fwd` or `_rev` suffixes

### 3. Fixed removeBGPSessionEdges()

**Previous (WRONG):**
```go
edgeKeys := []string{
    fmt.Sprintf("%s_fwd", sessionKey),
    fmt.Sprintf("%s_rev", sessionKey),
}
// Try to remove these specific keys
```

**New (CORRECT):**
```go
// Query for edges related to this session and remove them
query := `
    FOR edge IN collection
    FILTER edge.protocol LIKE "BGP_%"
    FILTER CONTAINS(edge._key, @sessionKey)
    REMOVE edge IN collection
`
```

This is more robust because:
- Doesn't assume specific key format
- Finds all related edges dynamically
- Handles edge cases better

## Files Modified

### `/ip-graph/arangodb/bgp-peer-processor.go`

**Changes:**
1. **Lines 172-199:** Simplified `createBGPSessionEdges()`
   - Removed conditional logic for bidirectional edges
   - Create exactly ONE edge per peer message

2. **Lines 201-249:** Simplified `createSessionEdge()`
   - Removed `isForward` parameter
   - Single, consistent edge creation logic

3. **Lines 319-350:** Fixed `removeBGPSessionEdges()`
   - Changed from key-based removal to query-based removal
   - More robust edge cleanup

## Expected Results

### Before Fix:
- 6 edges created for a single BGP session between two routers
  - 2 good edges ✅
  - 4 duplicate edges ❌

### After Fix:
- 2 edges created for a single BGP session between two routers
  - 1 edge from A → B (key based on B's info) ✅
  - 1 edge from B → A (key based on A's info) ✅
  - 0 duplicate edges ✅

### Edge Key Format:
- Format: `router_id_asn_ip`
- Example: `10.0.0.43_65042_10.2.1.11`
- Based on REMOTE peer's information
- Unique per peer message

## Verification Steps

1. **Check Edge Count:**
   ```aql
   FOR edge IN ipv4_graph
   FILTER edge.protocol LIKE "BGP_%"
   COLLECT from = edge._from, to = edge._to WITH COUNT INTO count
   FILTER count > 1
   RETURN { from, to, count }
   ```
   Should return 0 results (no duplicates)

2. **Check Edge Key Format:**
   ```aql
   FOR edge IN ipv4_graph
   FILTER edge.protocol LIKE "BGP_%"
   FILTER CONTAINS(edge._key, "_fwd") OR CONTAINS(edge._key, "_rev")
   RETURN edge
   ```
   Should return 0 results (no _fwd/_rev keys)

3. **Check Bidirectionality:**
   ```aql
   FOR edge IN ipv4_graph
   FILTER edge.protocol LIKE "BGP_%"
   FILTER edge._from == "bgp_node/A"
   FILTER edge._to == "bgp_node/B"
   RETURN edge
   ```
   Should return exactly 1 edge

   ```aql
   FOR edge IN ipv4_graph
   FILTER edge.protocol LIKE "BGP_%"
   FILTER edge._from == "bgp_node/B"
   FILTER edge._to == "bgp_node/A"
   RETURN edge
   ```
   Should return exactly 1 edge (reverse direction)

## Related Documents

- `/ip-graph/BGP_NODE_FIXES.md` - BGP node processing fixes
- `/ip-graph/BGP_PREFIX_FIXES.md` - BGP prefix classification fixes

---
*Document Created: 2025-11-01*
*Issue Reported By: Bruce McDougall*
*Fixed By: Automated Code Review*

