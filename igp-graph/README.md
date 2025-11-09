# IGP Graph Processor

The IGP Graph Processor is a unified microservice that merges the functionality of both `linkstate-edge` and `linkstate-graph` processors. It creates comprehensive IGP topology graphs optimized for large-scale networks (100k+ nodes).

## Features

- **Unified Processing**: Combines edge creation and full IGP graph construction
- **High Performance**: Batch processing and concurrent workers for large networks
- **Dual Graph Support**: Creates both IGPv4 and IGPv6 topology graphs
- **Backward Compatibility**: Maintains `ls_node_edge` collection
- **Fixed Prefix Handling**: 
  - Router IDs/Loopbacks → Node metadata
  - Transit Networks → Separate vertices
  - SRv6 Locators → Node metadata
- **Real-time Updates**: Event-driven incremental graph updates

## Architecture

### Components

- **Main Processor**: `arangodb.go` - Core database client and coordination
- **Batch Processor**: `batch-processor.go` - High-performance batch operations
- **Update Coordinator**: `update-coordinator.go` - Message routing and coordination
- **Kafka Messenger**: `kafkamessenger.go` - Kafka topic subscription

### Collections Created

- `igp_node` - Enhanced node collection with SR/SRv6 metadata
- `igp_domain` - IGP domain information
- `ls_node_edge` - Backward compatibility edge collection
- `igpv4_graph_edge` - IPv4 topology edges
- `igpv6_graph_edge` - IPv6 topology edges

### Graphs Created

- `igpv4_graph` - Complete IPv4 IGP topology
- `igpv6_graph` - Complete IPv6 IGP topology

## Configuration

### Command Line Flags

- `--batch_size`: Batch size for bulk operations (default: 1000)
- `--concurrent_workers`: Number of concurrent workers (default: 2x CPU cores)
- `--igpv4_graph`: IGPv4 graph name (default: "igpv4_graph")
- `--igpv6_graph`: IGPv6 graph name (default: "igpv6_graph")

### Performance Tuning

For large networks (100k+ nodes):
- Increase `batch_size` to 5000-10000
- Set `concurrent_workers` to 2-4x CPU cores
- Ensure adequate memory (8GB+ recommended)
- Use SSD storage for ArangoDB

## Deployment

```bash
# Build
make igp-graph

# Run
./bin/igp-graph \
  --database-server=http://arango:8529 \
  --database-name=jalapeno \
  --message-server=kafka:9092 \
  --batch_size=5000 \
  --concurrent_workers=16
```

## Migration from linkstate-edge/linkstate-graph

This processor replaces both `linkstate-edge` and `linkstate-graph`. It:

1. **Maintains compatibility** with existing `ls_node_edge` collection
2. **Enhances performance** with batch processing and optimized updates
3. **Provides unified management** of all IGP topology constructs
4. **Supports large-scale networks** with improved memory and CPU efficiency

## Development Status

- ✅ Core framework and batch processing
- 🚧 Processing logic implementation (next phase)
- 🚧 Performance optimization and testing
- 🚧 Migration utilities
