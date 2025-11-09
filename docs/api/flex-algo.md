# Flex-Algo Implementation Summary

## Overview
This document summarizes the Flex-Algo (Flexible Algorithm) support added to the Jalapeno API. Flex-Algo allows network operators to define multiple routing topologies (algorithms) within a single IGP domain, each optimized for different metrics (e.g., latency, bandwidth, sovereignty).

## What is Flex-Algo?
Flex-Algo is an IGP extension that allows:
- Multiple algorithm IDs (0-255) to coexist in a single IGP domain
- Each algorithm can have different optimization objectives
- Nodes can participate in multiple algorithms
- Each algorithm has its own SRv6 SID space

Common algorithm IDs:
- **Algo 0**: Default SPF (standard shortest path)
- **Algo 128**: Low latency optimization
- **Algo 129**: High bandwidth optimization
- **Algo 130+**: Custom optimization criteria

## Implementation Details

### 1. Data Structure Changes

#### Vertex SID Structure
Each `igp_node` vertex now contains SIDs with algo information:
```json
{
  "_id": "igp_node/2_0_0_0000.0000.0001",
  "name": "xrd01",
  "router_id": "10.0.0.1",
  "sids": [
    {
      "srv6_sid": "fc00:0:1::",
      "algo": 0,
      "endpoint_behavior": 48,
      "flag": 0
    },
    {
      "srv6_sid": "fc00:1:1::",
      "algo": 128,
      "endpoint_behavior": 48,
      "flag": 0
    }
  ]
}
```

### 2. API Endpoints Modified

#### A. New Algo-Aware Endpoints

##### `/graphs/{collection_name}/vertices/algo`
Lists all vertices that participate in a specific Flex-Algo.

**Example:**
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/algo?algo=128"
```

**Response:**
```json
{
  "graph_collection": "ipv6_graph",
  "algo": 128,
  "total_vertices": 12,
  "vertex_collections": ["igp_node"],
  "vertices_by_collection": {
    "igp_node": [
      {
        "_id": "igp_node/2_0_0_0000.0000.0001",
        "name": "xrd01",
        "router_id": "10.0.0.1",
        "sids": [
          {
            "srv6_sid": "fc00:1:1::",
            "algo": 128,
            "endpoint_behavior": 48,
            "flag": 0
          }
        ]
      }
    ]
  }
}
```

##### `/graphs/{collection_name}/topology/algo`
Returns topology visualization filtered by algo participation.

**Example:**
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/topology/algo?algo=128"
```

#### B. Updated Shortest Path Endpoints

All shortest path endpoints now support the `algo` parameter:

1. **Basic Shortest Path**
   - Endpoint: `/graphs/{collection_name}/shortest_path`
   - New parameter: `algo` (optional, default: 0)

2. **Latency-Optimized Path**
   - Endpoint: `/graphs/{collection_name}/shortest_path/latency`
   - New parameter: `algo` (optional, default: 0)

3. **Utilization-Optimized Path**
   - Endpoint: `/graphs/{collection_name}/shortest_path/utilization`
   - New parameter: `algo` (optional, default: 0)

4. **Load-Balanced Path**
   - Endpoint: `/graphs/{collection_name}/shortest_path/load`
   - New parameter: `algo` (optional, default: 0)

5. **Sovereignty-Constrained Path**
   - Endpoint: `/graphs/{collection_name}/shortest_path/sovereignty`
   - New parameter: `algo` (optional, default: 0)

6. **Best Paths (K-Shortest)**
   - Endpoint: `/graphs/{collection_name}/shortest_path/best-paths`
   - New parameter: `algo` (optional, default: 0)

7. **Next Best Paths**
   - Endpoint: `/graphs/{collection_name}/shortest_path/next-best-path`
   - New parameter: `algo` (optional, default: 0)

**Example:**
```bash
# Default algo (0)
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0000.0001&destination=igp_node/2_0_0_0000.0000.0018&direction=outbound"

# With Flex-Algo 128
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0000.0001&destination=igp_node/2_0_0_0000.0000.0018&direction=outbound&algo=128"
```

#### C. RPO (Resource Path Optimization) Endpoints

Both RPO endpoints now support Flex-Algo:

1. **Select Optimal Endpoint**
   - Endpoint: `/rpo/{collection_name}/select-optimal`
   - New parameter: `algo` (optional, default: 0)

2. **Select from List**
   - Endpoint: `/rpo/{collection_name}/select-from-list`
   - New parameter: `algo` (optional, default: 0)

**Example:**
```bash
# Select endpoint with lowest GPU utilization using Flex-Algo 128
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=gpu_utilization&graphs=ipv6_graph&algo=128"
```

### 3. Path Processing Changes

#### `path_processor.py`
Updated to support dynamic USID block detection and algo-aware SID selection:

**Key Changes:**
- Removed hardcoded `usid_block` default
- Auto-detects USID block from first SID in path
- Selects SIDs matching the specified algo from each node's SID array
- Falls back to algo 0 if specified algo not found

**Example:**
```python
# Automatically detects fc00:0: or fc00:1: or any other block
srv6_data = process_path_data(
    path_data=results[0]['path'],
    source=source,
    destination=destination,
    algo=128  # Will select SIDs with algo=128
)
```

### 4. AQL Query Logic

#### Algo Filtering Strategy
Since ArangoDB's `SHORTEST_PATH` doesn't support inline filtering, we use `K_SHORTEST_PATHS` with post-filtering:

```aql
FOR p IN OUTBOUND K_SHORTEST_PATHS '{source}' TO '{destination}' {collection_name}
    OPTIONS {{bfs: true}}
    
    // Filter to ensure all igp_nodes in path participate in the algo
    FILTER (
        FOR v IN p.vertices
            FILTER v._id LIKE 'igp_node/%'
            FILTER {algo} IN v.sids[*].algo
            COLLECT WITH COUNT INTO nodeCount
            RETURN nodeCount
    )[0] == LENGTH(
        FOR v IN p.vertices
            FILTER v._id LIKE 'igp_node/%'
            COLLECT WITH COUNT INTO nodeCount
            RETURN nodeCount
    )[0]
    
    LIMIT 1
    RETURN {{
        path: p.vertices[*],
        edges: p.edges[*],
        hopcount: LENGTH(p.vertices) - 1
    }}
```

**Logic Explanation:**
1. Find multiple shortest paths using `K_SHORTEST_PATHS`
2. For each path, verify that all `igp_node` vertices have the specified algo in their SID array
3. Return the first path that meets the criteria
4. If no path found with specified algo, return error

### 5. Response Format

All algo-aware endpoints include the algo in their response:

```json
{
  "found": true,
  "algo": 128,
  "path": [...],
  "hopcount": 7,
  "srv6_data": {
    "srv6_sid_list": [
      "fc00:1:1::",
      "fc00:1:3::",
      "fc00:1:7::",
      "fc00:1:18::"
    ],
    "srv6_usid": "fc00:1:1:3:7:18::"
  }
}
```

Note: The USID block automatically adapts based on the algo (e.g., `fc00:1:` for algo 128).

## Testing Scenarios

### Scenario 1: Verify Algo Participation
```bash
# List all nodes in algo 128
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/algo?algo=128"

# List all nodes in algo 0 (default)
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/algo?algo=0"
```

### Scenario 2: Compare Paths Across Algos
```bash
# Path using default algo
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0000.0001&destination=igp_node/2_0_0_0000.0000.0018&direction=outbound"

# Path using algo 128 (may take different route)
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0000.0001&destination=igp_node/2_0_0_0000.0000.0018&direction=outbound&algo=128"
```

### Scenario 3: Verify Alternate Path When Node Removed
```bash
# Remove node from algo 128 in the database
# Then verify path finds alternate route:
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0000.0001&destination=igp_node/2_0_0_0000.0000.0018&direction=outbound&algo=128"
```

### Scenario 4: RPO with Flex-Algo
```bash
# Select optimal endpoint using low-latency algo
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=gpu_utilization&graphs=ipv6_graph&algo=128"
```

## Error Handling

### No Path Found for Specified Algo
If no path exists using the specified algo (e.g., source or destination doesn't participate in that algo):

```json
{
  "detail": "No path found using algo 128 between specified nodes. Nodes may not participate in this algo."
}
```

### Invalid Algo Parameter
If algo parameter is not a valid integer:

```json
{
  "detail": "Invalid algo parameter. Must be an integer between 0 and 255."
}
```

## Benefits

1. **Flexibility**: Support multiple routing topologies in a single network
2. **Optimization**: Different paths for different traffic types (latency-sensitive, bandwidth-intensive, etc.)
3. **Sovereignty**: Combine with sovereignty constraints for geo-aware routing
4. **Automation**: SRv6 USID generation automatically adapts to the selected algo
5. **Compatibility**: Default behavior (algo 0) maintains backward compatibility

## Future Enhancements

Potential future improvements:
1. Support for algo preferences (try algo X, fallback to algo Y)
2. Algo-specific metrics (e.g., latency only for algo 128 paths)
3. Multi-algo path comparison in a single API call
4. Algo validation endpoint (verify node participation before path calculation)
5. Dynamic algo discovery from IGP data

## Code Files Modified

1. **`app/routes/graphs.py`**
   - Added `algo` parameter to all shortest path endpoints
   - Added `/vertices/algo` endpoint
   - Added `/topology/algo` endpoint
   - Updated AQL queries to filter by algo participation

2. **`app/routes/rpo.py`**
   - Added `algo` parameter to both RPO endpoints
   - Pass algo to `get_shortest_path` calls

3. **`app/utils/path_processor.py`**
   - Removed hardcoded USID block
   - Added auto-detection of USID block
   - Added algo-aware SID selection
   - Updated to handle multiple SIDs per node

## Conclusion

The Flex-Algo implementation provides comprehensive support for multi-topology routing in SRv6 networks. All shortest path and RPO endpoints now support algo-aware path computation, with automatic SRv6 USID generation based on the selected algorithm. The implementation maintains backward compatibility while enabling advanced traffic engineering capabilities.

