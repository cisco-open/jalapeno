# ArangoDB Query Examples for API

This document contains example AQL queries used by the Jalapeno API.

## Basic Collection Queries

### Query a specific document by key

```aql
FOR doc IN hosts
    FILTER doc._key == "amsterdam"
    RETURN {
        _id: doc._id,
        _key: doc._key,
        cpu_utilization: doc.cpu_utilization,
        gpu_utilization: doc.gpu_utilization,
        memory_utilization: doc.memory_utilization,
        time_to_first_token: doc.time_to_first_token,
        cost_per_million_tokens: doc.cost_per_million_tokens,
        cost_per_hour: doc.cost_per_hour,
        gpu_model: doc.gpu_model,
        language_model: doc.language_model,
        available_capacity: doc.available_capacity,
        response_time: doc.response_time
    }
```

### Find document with lowest metric value

```aql
FOR doc IN hosts
    FILTER doc.gpu_utilization != null
    SORT doc.gpu_utilization ASC
    LIMIT 1
    RETURN {
        _id: doc._id,
        _key: doc._key,
        name: doc.name,
        gpu_utilization: doc.gpu_utilization
    }
```

## Resource Path Optimization Queries

### Find optimal endpoint based on metric

```aql
FOR doc IN @@collection
    FILTER doc.@@metric != null
    SORT doc.@@metric ASC
    LIMIT 1
    RETURN doc
```

### Find endpoints matching specific value

```aql
FOR doc IN @@collection
    FILTER doc.@@metric == @value
    RETURN doc
```

## Graph Traversal Queries

### Shortest path query

```aql
FOR v, e IN OUTBOUND SHORTEST_PATH @source TO @destination @@graph
    RETURN {
        vertices: v,
        edges: e
    }
```

### K-shortest paths query

```aql
FOR p IN OUTBOUND K_SHORTEST_PATHS @source TO @destination @@graph
    OPTIONS {bfs: true}
    LIMIT @limit
    RETURN {
        path: p.vertices[*],
        edges: p.edges[*],
        hopcount: LENGTH(p.vertices) - 1
    }
```

### Flex-Algo aware shortest path

```aql
FOR p IN OUTBOUND K_SHORTEST_PATHS @source TO @destination @@graph
    OPTIONS {bfs: true}
    
    // Filter to ensure all igp_nodes in path participate in the algo
    FILTER (
        FOR v IN p.vertices
            FILTER v._id LIKE 'igp_node/%'
            FILTER @algo IN v.sids[*].algo
            COLLECT WITH COUNT INTO nodeCount
            RETURN nodeCount
    )[0] == LENGTH(
        FOR v IN p.vertices
            FILTER v._id LIKE 'igp_node/%'
            COLLECT WITH COUNT INTO nodeCount
            RETURN nodeCount
    )[0]
    
    LIMIT 1
    RETURN {
        path: p.vertices[*],
        edges: p.edges[*],
        hopcount: LENGTH(p.vertices) - 1
    }
```

### Neighbors query

```aql
FOR v, e IN 1..@depth OUTBOUND @source @@graph
    RETURN DISTINCT v
```

### Traverse with depth limit

```aql
FOR v, e, p IN 1..@max_depth OUTBOUND @source @@graph
    FILTER v._id == @destination
    RETURN {
        path: p.vertices[*],
        edges: p.edges[*]
    }
```

## Topology Queries

### Get all graph edges

```aql
FOR edge IN @@graph
    RETURN edge
```

### Get all vertices from graph

```aql
FOR v IN 1..1 ANY @start_vertex @@graph
    RETURN DISTINCT v
```

### Get vertices by algorithm

```aql
FOR collection_name IN @vertex_collections
    FOR vertex IN @@db[collection_name]
        FILTER @algo IN vertex.sids[*].algo
        RETURN vertex
```

## Path Optimization Queries

### Latency-weighted path

```aql
FOR v, e IN OUTBOUND SHORTEST_PATH @source TO @destination @@graph
    OPTIONS {
        weightAttribute: 'latency',
        defaultWeight: 1
    }
    RETURN {
        vertices: v,
        edges: e
    }
```

### Utilization-weighted path

```aql
FOR v, e IN OUTBOUND SHORTEST_PATH @source TO @destination @@graph
    OPTIONS {
        weightAttribute: 'percent_util_out',
        defaultWeight: 1
    }
    RETURN {
        vertices: v,
        edges: e
    }
```

### Sovereignty-constrained path

```aql
FOR v, e IN OUTBOUND SHORTEST_PATH @source TO @destination @@graph
    PRUNE v.country IN @excluded_countries
    FILTER v.country NOT IN @excluded_countries
    RETURN {
        vertices: v,
        edges: e
    }
```

## Bulk Operations

### Update edge attribute

```aql
FOR edge IN @@graph
    UPDATE edge WITH { @attribute: @value } IN @@graph
    RETURN NEW
```

### Reset all edge loads

```aql
FOR edge IN @@graph
    UPDATE edge WITH { load: 0 } IN @@graph
```

## Search Queries

### Search by multiple criteria

```aql
FOR doc IN @@collection
    FILTER doc.asn == @asn
    FILTER doc.protocol == @protocol
    FILTER doc.srv6_enabled == @srv6_enabled
    RETURN doc
```

### Get collection keys only

```aql
FOR doc IN @@collection
    RETURN doc._key
```

### Get collection IDs only

```aql
FOR doc IN @@collection
    RETURN doc._id
```

## Summary and Statistics

### Vertex summary by collection

```aql
FOR vertex_coll IN @vertex_collections
    LET count = LENGTH(@@db[vertex_coll])
    RETURN {
        collection: vertex_coll,
        count: count
    }
```

### Get vertices with sample data

```aql
FOR vertex_coll IN @vertex_collections
    LET sample = (
        FOR v IN @@db[vertex_coll]
            LIMIT @limit
            RETURN v
    )
    RETURN {
        collection: vertex_coll,
        count: LENGTH(@@db[vertex_coll]),
        sample: sample
    }
```

## Notes

- Bind parameters are prefixed with `@` (e.g., `@source`, `@destination`)
- Collection binds use `@@` (e.g., `@@collection`, `@@graph`)
- All queries should use parameterized inputs to prevent AQL injection
- The API automatically handles parameter binding and escaping

