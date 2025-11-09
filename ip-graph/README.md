# IP Graph Processor

The IP Graph processor creates comprehensive IPv4 and IPv6 topology graphs by combining IGP topology data with BGP routing information.

## Overview

The IP Graph processor extends the pure IGP topology graphs created by the IGP Graph processor by adding:

- BGP peer nodes and session edges
- BGP prefix vertices with appropriate classification
- Full end-to-end connectivity across IGP and BGP domains
- Real-time updates for topology changes

## Architecture

### Data Flow

```
IGP Graph Data (igpv4_graph, igpv6_graph)
    ↓ (copy)
IP Graph Data (ipv4_graph, ipv6_graph)
    ↓ (extend)
+ BGP Peers (vertices)
+ BGP Sessions (edges)  
+ BGP Prefixes (vertices with /32,/128 as node metadata)
    ↓
Full IP Topology Graphs
```

### Collections

**Source Collections (Read-Only)**:
- `igpv4_graph` - Pure IGP IPv4 topology
- `igpv6_graph` - Pure IGP IPv6 topology
- `igp_node` - IGP node data
- `igp_domain` - IGP domain data

**Target Collections (Managed by IP Graph)**:
- `ipv4_graph` - Full IPv4 topology (IGP + BGP)
- `ipv6_graph` - Full IPv6 topology (IGP + BGP)
- `bgp_node` - BGP peer nodes
- `bgp_prefix_v4` - BGP IPv4 prefixes
- `bgp_prefix_v6` - BGP IPv6 prefixes

### Processing Strategy

1. **Initial Load**:
   - Copy IGP graph data to IP graph collections
   - Load existing BGP peer and prefix data
   - Build full topology graphs

2. **Real-Time Updates**:
   - Monitor IGP changes and sync to IP graphs
   - Process BGP peer session changes
   - Process BGP prefix advertisements/withdrawals
   - Apply /32 and /128 prefix metadata strategy

## BGP Topology Model

### BGP Peers as Vertices
```json
{
  "_key": "bgp_65001_1.1.1.1",
  "router_id": "1.1.1.1",
  "asn": 65001,
  "node_type": "bgp",
  "tier": "provider"
}
```

### BGP Sessions as Edges
```json
{
  "_key": "session_65001_65002",
  "_from": "bgp_node/bgp_65001_1.1.1.1",
  "_to": "bgp_node/bgp_65002_2.2.2.2",
  "local_asn": 65001,
  "remote_asn": 65002,
  "session_state": "established"
}
```

### BGP Prefixes
- **Host prefixes** (/32, /128): Added as metadata to advertising node
- **Transit prefixes**: Separate vertices with edges to advertising peers
- **Classification**: iBGP, eBGP private, eBGP public, Internet

## Configuration

### Command Line Flags

```bash
--database-server="http://arangodb:8529"
--database-name="jalapeno"
--igpv4-graph="igpv4_graph"          # Source IGP IPv4 graph
--igpv6-graph="igpv6_graph"          # Source IGP IPv6 graph
--ipv4-graph="ipv4_graph"            # Target full IPv4 topology
--ipv6-graph="ipv6_graph"            # Target full IPv6 topology
--bgp-node="bgp_node"                # BGP peer collection
--batch-size=1000                    # Batch processing size
--concurrent-workers=8               # Number of worker threads
```

### Kafka Topics

The processor subscribes to raw BMP topics:

- `gobmp.parsed.ls_node` - IGP node sync
- `gobmp.parsed.ls_link` - IGP link sync  
- `gobmp.parsed.peer` - BGP peer sessions
- `gobmp.parsed.unicast_prefix_v4` - BGP IPv4 prefixes
- `gobmp.parsed.unicast_prefix_v6` - BGP IPv6 prefixes

## Performance

- **Batch Processing**: Configurable batch sizes for high-throughput scenarios
- **Concurrent Workers**: Multi-threaded processing for scalability
- **Incremental Updates**: Only processes topology changes, not full rebuilds
- **Memory Efficient**: Streaming processing without loading full datasets

## Use Cases

1. **End-to-End Path Analysis**: Complete visibility across IGP and BGP domains
2. **Multi-Domain Routing**: Understanding inter-AS connectivity
3. **Traffic Engineering**: Comprehensive topology for path optimization
4. **Network Monitoring**: Real-time topology change detection
5. **Service Mapping**: Mapping services across network boundaries

## Development Status

- ✅ Core architecture and data structures
- ✅ IGP graph copying logic  
- ✅ Kafka message processing framework
- ✅ Batch processing system
- 🚧 BGP peer processing (in development)
- 🚧 BGP prefix processing (in development)
- 🚧 Real-time update coordination (in development)

## Related Components

- **IGP Graph**: Provides pure IGP topology data
- **GoBMP Arango**: Provides raw BMP data from network devices
- **Kafka**: Message bus for real-time updates
- **ArangoDB**: Graph database for topology storage
