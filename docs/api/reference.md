# Jalapeno API Reference

Complete reference for the Jalapeno REST API.

## Base URL

```
http://localhost:8000/api/v1
```

---

## Collections

### Get all collections
```bash
curl http://localhost:8000/api/v1/collections
```

### Get only graph collections
```bash
curl http://localhost:8000/api/v1/collections?filter_graphs=true
```

### Get only non-graph collections
```bash
curl http://localhost:8000/api/v1/collections?filter_graphs=false
```

### Get data from any collection
```bash
curl "http://localhost:8000/api/v1/collection/ls_node"
curl "http://localhost:8000/api/v1/collection/ls_link"
curl "http://localhost:8000/api/v1/collection/ls_prefix"
curl "http://localhost:8000/api/v1/collection/ls_srv6_sid"
curl "http://localhost:8000/api/v1/collection/bgp_node"
curl "http://localhost:8000/api/v1/collection/igp_node"
curl "http://localhost:8000/api/v1/collection/bgp_prefix_v4"
curl "http://localhost:8000/api/v1/collection/bgp_prefix_v6"
```

### Get data with limits
```bash
curl "http://localhost:8000/api/v1/collection/bgp_node?limit=10"
curl "http://localhost:8000/api/v1/collection/igp_node?limit=10"
```

### Get data with a specific key
```bash
curl "http://localhost:8000/api/v1/collection/bgp_node?filter_key=some_key"
```

### Get just the keys from a collection
```bash
curl "http://localhost:8000/api/v1/collection/peer/keys"
```

---

## Search

### Search by ASN only
```bash
curl "http://localhost:8000/api/v1/collection/igp_node/search?asn=65001"
```

### Search by protocol only
```bash
curl "http://localhost:8000/api/v1/collection/igp_node/search?protocol=IS-IS%20Level%202"
```

### Search with multiple filters
```bash
curl "http://localhost:8000/api/v1/collection/igp_node/search?asn=65001&srv6_enabled=true"
```

---

## Graphs

### Get specific graph data
```bash
curl http://localhost:8000/api/v1/collections/igpv4_graph
curl http://localhost:8000/api/v1/collections/igpv6_graph
curl http://localhost:8000/api/v1/collections/ipv4_graph
curl http://localhost:8000/api/v1/collections/ipv6_graph
```

### Get graph info
```bash
curl http://localhost:8000/api/v1/collections/igpv4_graph/info
curl http://localhost:8000/api/v1/collections/igpv6_graph/info
```

### Get graph edges
```bash
curl http://localhost:8000/api/v1/graphs/ipv6_graph/edges
curl http://localhost:8000/api/v1/graphs/ipv4_graph/edges
curl http://localhost:8000/api/v1/graphs/igpv6_graph/edges
curl http://localhost:8000/api/v1/graphs/igpv4_graph/edges
```

### Get graph vertices
```bash
curl http://localhost:8000/api/v1/graphs/ipv6_graph/vertices
curl http://localhost:8000/api/v1/graphs/ipv4_graph/vertices
curl http://localhost:8000/api/v1/graphs/igpv6_graph/vertices
curl http://localhost:8000/api/v1/graphs/igpv4_graph/vertices
```

### Get vertex keys
```bash
curl http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/keys
curl http://localhost:8000/api/v1/graphs/ipv4_graph/vertices/keys
```

### Get vertex IDs
```bash
curl http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/ids
curl http://localhost:8000/api/v1/graphs/ipv4_graph/vertices/ids
```

### Get vertices by algorithm (Flex-Algo)
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/algo?algo=128"
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/algo?algo=129"
```

---

## Topology

### Get full topology
```bash
curl http://localhost:8000/api/v1/graphs/ipv6_graph/topology
curl http://localhost:8000/api/v1/graphs/ipv6_graph/topology?limit=50
```

### Get node-to-node connections
```bash
curl http://localhost:8000/api/v1/graphs/ipv6_graph/topology/nodes
curl http://localhost:8000/api/v1/graphs/ipv6_graph/topology/nodes?limit=50
```

### Get topology per algorithm
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/topology/nodes/algo?algo=128"
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/topology/nodes/algo?algo=129"
```

### Get vertex summary
```bash
curl http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/summary
curl http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/summary?limit=25
curl http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/summary?vertex_collection=igp_node
curl http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/summary?vertex_collection=igp_node&limit=10
```

---

## Shortest Path

### Basic shortest path
```bash
# Outbound (default)
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0001.0065&destination=igp_node/2_0_0_0000.0002.0067"

# Inbound
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0001.0065&destination=igp_node/2_0_0_0000.0002.0067&direction=inbound"

# Any direction
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0001.0065&destination=igp_node/2_0_0_0000.0002.0067&direction=any"
```

### Shortest path with Flex-Algo
```bash
# Algo 0 (default SPF)
curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path?source=bgp_prefix_v4/10.10.46.0_24&destination=bgp_prefix_v4/96.1.0.0_24&direction=any&algo=0"

# Algo 128 (low latency)
curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path?source=bgp_prefix_v4/10.10.46.0_24&destination=bgp_prefix_v4/96.1.0.0_24&direction=any&algo=128"

# Algo 129 (high bandwidth)
curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path?source=bgp_prefix_v4/10.10.46.0_24&destination=bgp_prefix_v4/96.1.0.0_24&direction=any&algo=129"
```

### Prefix to prefix path
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=ls_prefix/2_0_2_0_0_fc00:0:701:1::_64_0000.0001.0065&destination=ls_prefix/2_0_2_0_0_fc00:0:701:1003::_64_0000.0002.0067&direction=any"
```

---

## Optimized Paths

### Latency-weighted shortest path
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path/latency?source=gpus/host08-gpu02&destination=gpus/host12-gpu02"
```

### Utilization-optimized path
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path/utilization?source=gpus/host08-gpu02&destination=gpus/host12-gpu02&direction=outbound"
```

### Load-balanced path
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path/load?source=gpus/host01-gpu02&destination=gpus/host12-gpu02&direction=any"
```

### Load-balanced path with Flex-Algo
```bash
curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path/load?source=bgp_prefix_v4/10.10.46.0_24&destination=bgp_prefix_v4/96.1.0.0_24&direction=any&algo=128"
```

### Sovereignty-constrained path
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path/sovereignty?source=hosts/berlin-k8s&destination=hosts/rome&excluded_countries=FRA&direction=outbound"
```

### Sovereignty with Flex-Algo
```bash
curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path/sovereignty?source=bgp_prefix_v4/10.10.46.0_24&destination=bgp_prefix_v4/10.17.1.0_24&excluded_countries=FRA&direction=any&algo=0"
```

---

## K-Shortest Paths

### Best paths (K-shortest)
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path/best-paths?source=hosts/amsterdam&destination=hosts/rome&direction=outbound"
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path/best-paths?source=hosts/amsterdam&destination=hosts/rome&direction=outbound&limit=6"
```

### Best paths with Flex-Algo
```bash
curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path/best-paths?source=bgp_prefix_v4/10.17.1.0_24&destination=bgp_prefix_v4/96.1.0.0_24&limit=5&algo=130"
```

### Next best path
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path/next-best-path?source=hosts/berlin-k8s&destination=hosts/rome&direction=outbound"
```

### Next best path with Flex-Algo
```bash
curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path/next-best-path?source=bgp_prefix_v4/10.17.1.0_24&destination=bgp_prefix_v4/96.1.0.0_24&direction=any&algo=0"

curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path/next-best-path?source=bgp_prefix_v4/10.17.1.0_24&destination=bgp_prefix_v4/96.1.0.0_24&direction=any&same_hop_limit=2&plus_one_limit=5&algo=0"

curl "http://localhost:8000/api/v1/graphs/ipv4_graph/shortest_path/next-best-path?source=bgp_prefix_v4/10.17.1.0_24&destination=bgp_prefix_v4/96.1.0.0_24&direction=any&algo=128"
```

---

## Graph Traversal

### Simple traverse
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/traverse/simple?source=ls_prefix/2_0_2_0_0_fc00:0:701:1::_64_0000.0001.0065&destination=igp_node/2_0_0_0000.0002.0067"

curl "http://localhost:8000/api/v1/graphs/ipv6_graph/traverse/simple?start_node=ls_prefix/2_0_2_0_0_fc00:0:701:1::_64_0000.0001.0065&target_node=ls_prefix/2_0_2_0_0_fc00:0:701:1003::_64_0000.0002.0067&max_depth=6"

curl "http://localhost:8000/api/v1/graphs/ipv6_graph/traverse/simple?source=ls_prefix/2_0_2_0_0_fc00:0:701:1::_64_0000.0001.0065&max_depth=5"
```

### Complex traverse
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/traverse?source=igp_node/2_0_0_0000.0001.0065&max_depth=3"

curl "http://localhost:8000/api/v1/graphs/ipv6_graph/traverse?source=ls_prefix/2_0_2_0_0_fc00:0:701:1::_64_0000.0001.0065&destination=igp_node/2_0_0_0000.0002.0067&max_depth=5&direction=any"
```

---

## Neighbors

### Get immediate neighbors
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/neighbors?node=igp_node/2_0_0_0000.0001.0001"
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/neighbors?source=igp_node/2_0_0_0000.0001.0065"
```

### Get neighbors with specific direction
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/neighbors?source=igp_node/2_0_0_0000.0001.0065&direction=any"
```

### Get neighbors with greater depth
```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/neighbors?source=igp_node/2_0_0_0000.0001.0065&depth=2"
```

---

## Edge Operations

### Reset load on all edges (with AQL)
```aql
FOR edge IN ipv6_graph
  UPDATE edge WITH { load: 0 } IN ipv6_graph
```

### Reset load on all edges (with curl)
```bash
curl -X POST "http://localhost:8000/api/v1/graphs/ipv6_graph/edges" \
     -H "Content-Type: application/json" \
     -d '{"attribute": "load", "value": 0}'
```

---

## Resource Path Optimization (RPO)

For detailed RPO examples, see [RPO API Documentation](rpo.md).

### Basic RPO endpoint selection
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=gpu_utilization&graphs=ipv6_graph"
```

### RPO with Flex-Algo
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=gpu_utilization&graphs=ipv6_graph&algo=128"
```

---

## API Features

### Common Query Parameters

- **direction**: Path direction
  - `outbound` (default): Follow outbound edges
  - `inbound`: Follow inbound edges
  - `any`: Follow edges in any direction

- **algo**: Flex-Algorithm ID (default: 0)
  - `0`: Default SPF
  - `128`: Typically low latency
  - `129`: Typically high bandwidth
  - `130+`: Custom algorithms

- **limit**: Limit number of results returned

### Response Features

- **SRv6 USID Generation**: Automatically generates SRv6 Micro-SID lists for paths
- **Hop Count**: Returns number of hops in the path
- **Full Path Data**: Returns complete vertex and edge information for paths
- **Algo-Aware**: SRv6 SIDs automatically selected based on specified algorithm

---

## Additional Documentation

- [Flex-Algo Implementation](flex-algo.md) - Detailed Flex-Algo support documentation
- [RPO API](rpo.md) - Complete Resource Path Optimization examples
- [ArangoDB Queries](../arango/api-queries.md) - Example AQL queries used by the API

---

## Interactive Documentation

The API provides interactive Swagger/OpenAPI documentation:

```
http://localhost:8000/docs
```

