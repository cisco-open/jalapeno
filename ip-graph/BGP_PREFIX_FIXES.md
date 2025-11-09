# BGP Prefix Processing Fixes - Summary

## Overview
This document summarizes the fixes applied to match the original `ipv4-graph` and `ipv6-graph` behavior for BGP prefix classification, specifically for "Internet" (public ASN) prefixes.

## Issue Reported

**Missing Prefix:** The new code was not capturing this unicast_prefix_v4 entry:

```json
{
  "_key": "6.6.0.0_22_10.6.6.2",
  "prefix": "6.6.0.0",
  "prefix_len": 22,
  "peer_asn": 100006,
  "origin_as": 100006,
  "base_attrs": {
    "as_path": [100006],
    "as_path_count": 1
  }
}
```

**Characteristics:**
- Both `peer_asn` and `origin_as` are **100006** (public ASN)
- Both are **the same value** (direct announcement from origin)
- `as_path_count`: 1
- Prefix length: 22

This is a **direct announcement** from the origin AS, which should be captured as an Internet prefix.

## Root Cause Analysis

### Issue #1: Incorrect Filter Logic

**Location:** `bgp-deduplication-processor.go` line 172 (IPv4) and line 320 (IPv6)

**Previous Code:**
```aql
FILTER u.peer_asn != u.origin_as
```

This filter **excluded** prefixes where `peer_asn == origin_as`, which are direct announcements from the origin AS.

**Original Code (from ipv4-graph and ipv6-graph):**
```aql
filter u.remote_asn != u.origin_as
```

**The Problem:** The field `remote_asn` **does not exist** in the `unicast_prefix` collection! This was a quirk/bug in the original code. When AQL evaluates `NULL != value`, it results in behavior that effectively makes this filter a no-op, so ALL prefixes meeting the other criteria were captured, including direct announcements.

**Fix:** Remove the `FILTER u.peer_asn != u.origin_as` line entirely to match original behavior and capture direct announcements from origin ASNs.

### Issue #2: Wrong Field for Internal ASN List

**Location:** `bgp-deduplication-processor.go` line 167 (IPv4) and line 314 (IPv6)

**Previous Code:**
```aql
LET internal_asns = (FOR l IN igp_node RETURN l.asn)
```

**Original Code:**
```aql
let internal_asns = ( for l in ls_node return l.peer_asn )
```

Note: `ls_node` in original is equivalent to `igp_node` in new code.

**The Problem:** Using `l.asn` instead of `l.peer_asn` to build the internal ASN list. This could cause incorrect filtering of what ASNs are considered "internal" vs "external".

**Fix:** Changed to `l.peer_asn` to match original logic exactly.

## Changes Applied

### File: `/ip-graph/arangodb/bgp-deduplication-processor.go`

#### 1. processInternetIPv4Prefixes() - Lines 160-196

**Changes:**
- Line 170: Changed `l.asn` → `l.peer_asn`
- Line 172: **Removed** `FILTER u.peer_asn != u.origin_as`
- Added comment explaining the quirk in original code

**New Query:**
```aql
FOR u IN unicast_prefix_v4 
LET internal_asns = (FOR l IN igp_node RETURN l.peer_asn)  ← Fixed
FILTER u.peer_asn NOT IN internal_asns 
FILTER u.peer_asn NOT IN 64512..65535 
FILTER u.origin_as NOT IN 64512..65535 
FILTER u.prefix_len < 30 
-- REMOVED: FILTER u.peer_asn != u.origin_as  ← Fixed
INSERT { ... }
```

#### 2. processInternetIPv6Prefixes() - Lines 310-347

**Changes:**
- Line 320: Changed `l.asn` → `l.peer_asn`
- Line 320: **Removed** `FILTER u.peer_asn != u.origin_as`
- Added comment explaining the quirk in original code

**New Query:**
```aql
FOR u IN unicast_prefix_v6 
LET internal_asns = (FOR l IN igp_node RETURN l.peer_asn)  ← Fixed
FILTER u.peer_asn NOT IN internal_asns 
FILTER u.peer_asn NOT IN 64512..65535 
FILTER u.peer_asn NOT IN 4200000000..4294967294 
FILTER u.origin_as NOT IN 64512..65535 
FILTER u.prefix_len < 96 
-- REMOVED: FILTER u.peer_asn != u.origin_as  ← Fixed
INSERT { ... }
```

## Impact Analysis

### What Prefixes Are Now Captured

With these fixes, the following prefix types will now be properly captured as `ebgp_public`:

1. **Direct Announcements from Origin AS:**
   - Example: peer_asn = 100006, origin_as = 100006
   - AS path length = 1
   - Previously: **EXCLUDED** ❌
   - Now: **CAPTURED** ✅

2. **Transited Public Prefixes:**
   - Example: peer_asn = 65042, origin_as = 100096
   - AS path length > 1
   - Previously: **CAPTURED** ✅
   - Now: **CAPTURED** ✅

### Filter Criteria (after fix)

A prefix will be classified as `ebgp_public` if it meets **ALL** of these criteria:

1. ✅ `peer_asn` is **NOT** in the IGP domain (by `peer_asn`)
2. ✅ `peer_asn` is **NOT** a 2-byte private ASN (64512-65535)
3. ✅ `peer_asn` is **NOT** a 4-byte private ASN (4200000000-4294967294) [IPv6 only]
4. ✅ `origin_as` is **NOT** a 2-byte private ASN (64512-65535)
5. ✅ `prefix_len` < 30 (IPv4) or < 96 (IPv6)
6. ✅ **NO RESTRICTION** on whether `peer_asn == origin_as` (captures direct announcements)

## Verification

### Test Case 1: Direct Announcement (Previously Missing)
```json
{
  "prefix": "6.6.0.0/22",
  "peer_asn": 100006,
  "origin_as": 100006
}
```
- ✅ peer_asn NOT in IGP domain
- ✅ peer_asn NOT private (100006 is public)
- ✅ origin_as NOT private (100006 is public)
- ✅ prefix_len (22) < 30
- **Result: CAPTURED** ✅

### Test Case 2: Transited Public Prefix
```json
{
  "prefix": "96.1.0.0/24",
  "peer_asn": 65042,
  "origin_as": 100096
}
```
- ✅ peer_asn NOT in IGP domain
- ✅ peer_asn is private (65042) but origin_as is public
- ❌ peer_asn IS private → **Will be captured by ebgp_private filter instead**
- **Result: CAPTURED as ebgp_private** ✅

### Test Case 3: Internal IGP Prefix
```json
{
  "prefix": "10.1.1.0/24",
  "peer_asn": 100001,
  "origin_as": 100001
}
```
- If 100001 is in IGP domain (by peer_asn)
- ❌ peer_asn IS in IGP domain
- **Result: NOT CAPTURED by ebgp_public (correct)** ✅

## Known Quirks Preserved

### 1. Original "remote_asn" Bug
The original code referenced a non-existent field `u.remote_asn` in the filter. This was either:
- A typo/bug in the original code
- An intentional use of a NULL value to make the filter a no-op

We preserved this behavior by removing the filter entirely, which has the same effect.

### 2. Prefix Length Filter Anomaly
The original IPv4 code had `filter u.prefix_len < 96` which seems wrong for IPv4 (should be < 30 or < 32). The new code uses the correct value (< 30), which is an **intentional improvement** over the original.

## Testing Recommendations

1. **Verify Direct Announcements:**
   - Query for prefixes where `peer_asn == origin_as` and both are public
   - Confirm they appear in `bgp_prefix_v4` with `prefix_type: "ebgp_public"`

2. **Verify Transited Prefixes:**
   - Query for prefixes where `peer_asn != origin_as` 
   - Confirm they are classified correctly (private vs public)

3. **Verify Internal Exclusion:**
   - Query for prefixes where `peer_asn` is in IGP domain
   - Confirm they do NOT appear in Internet prefix collection

4. **Count Comparison:**
   - Compare prefix counts between original and new processors
   - Should now match exactly (or very close)

## Related Files

- `/ip-graph/arangodb/bgp-deduplication-processor.go` - Main file changed
- `/ip-graph/BGP_NODE_FIXES.md` - Related BGP node fixes

## Next Steps

After verifying these prefix fixes work correctly:

1. Review eBGP private prefix processing
2. Review iBGP prefix processing  
3. Review prefix-to-node attachment logic
4. Review edge creation for prefix advertisements
5. Review routing precedence application

---
*Document Created: 2025-11-01*
*Issue Reported By: Bruce McDougall*
*Fixed By: Automated Code Review*

