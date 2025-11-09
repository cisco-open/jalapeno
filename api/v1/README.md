# Jalapeno API

A FastAPI-based REST API for querying and analyzing network topology data from Jalapeno's ArangoDB graph database.

## Features

- **Graph Operations**: Shortest path, K-shortest paths, graph traversal, neighbors
- **Path Optimization**: Latency, utilization, load balancing, sovereignty constraints
- **Flex-Algo Support**: Multi-topology routing with algorithm-aware path computation
- **Resource Path Optimization (RPO)**: Intelligent destination selection based on metrics
- **SRv6 Integration**: Automatic SRv6 USID generation for computed paths
- **Collection Management**: Query and search across all ArangoDB collections

## Prerequisites

- Python 3.9+
- Access to Jalapeno ArangoDB instance
- Kubernetes cluster (for production deployment)

## Quick Start

### Local Development

1. **Create virtual environment**

```bash
cd api/v1
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

2. **Install dependencies**

```bash
pip install -r requirements.txt
```

3. **Set environment variables**

```bash
export LOCAL_DEV=1
# Optional: Set custom ArangoDB connection
export ARANGO_HOST=localhost
export ARANGO_PORT=8529
export ARANGO_USER=root
export ARANGO_PASSWORD=jalapeno
export ARANGO_DB=jalapeno
```

4. **Run the API**

```bash
uvicorn app.main:app --reload
```

5. **Access the API**

- API Documentation: http://localhost:8000/docs
- Alternative docs: http://localhost:8000/redoc
- API Root: http://localhost:8000/api/v1

## Project Structure

```
api/v1/
├── app/
│   ├── config/           # Configuration and settings
│   ├── routes/           # API endpoint definitions
│   │   ├── collections.py
│   │   ├── graphs.py
│   │   ├── instances.py
│   │   ├── rpo.py
│   │   └── vpns.py
│   ├── utils/            # Helper functions
│   │   ├── load_processor.py
│   │   └── path_processor.py
│   └── main.py           # FastAPI application entry point
├── requirements.txt      # Python dependencies
└── README.md             # This file
```

## Example Usage

### Get all collections

```bash
curl http://localhost:8000/api/v1/collections
```

### Find shortest path

```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0000.0001&destination=igp_node/2_0_0_0000.0000.0018&direction=outbound"
```

### Find shortest path with Flex-Algo 128

```bash
curl "http://localhost:8000/api/v1/graphs/ipv6_graph/shortest_path?source=igp_node/2_0_0_0000.0000.0001&destination=igp_node/2_0_0_0000.0000.0018&direction=outbound&algo=128"
```

### Resource Path Optimization

```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=gpu_utilization&graphs=ipv6_graph"
```

### Get topology summary

```bash
curl http://localhost:8000/api/v1/graphs/ipv6_graph/vertices/summary
```

## Production Deployment

### Build Docker Image

```bash
docker build -t iejalapeno/jalapeno-api:latest -f ../../build/Dockerfile.api .
```

### Deploy to Kubernetes

```bash
kubectl apply -f ../../deployment/api-deployment.yaml
```

## Configuration

The API uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `LOCAL_DEV` | Enable local development mode | `0` |
| `ARANGO_HOST` | ArangoDB host | `arango.jalapeno.svc.cluster.local` |
| `ARANGO_PORT` | ArangoDB port | `8529` |
| `ARANGO_USER` | ArangoDB username | `root` |
| `ARANGO_PASSWORD` | ArangoDB password | `jalapeno` |
| `ARANGO_DB` | ArangoDB database name | `jalapeno` |

## API Documentation

For detailed API documentation, see:

- **[API Reference](../../docs/api/reference.md)** - Complete endpoint reference
- **[Flex-Algo Guide](../../docs/api/flex-algo.md)** - Flex-Algorithm implementation details
- **[RPO Examples](../../docs/api/rpo.md)** - Resource Path Optimization examples
- **Interactive Docs** - http://localhost:8000/docs (when running)

## Key Features

### Flex-Algo Support

The API supports Flexible Algorithm (Flex-Algo) for multi-topology routing:

- Query vertices by algorithm participation
- Compute paths constrained to specific algorithms
- Automatic SRv6 SID selection based on algorithm
- Support for algorithms 0-255

### Resource Path Optimization (RPO)

Intelligent destination selection combining metrics with path computation:

- Minimize/maximize numeric metrics (CPU, GPU, latency, cost)
- Exact match for categorical requirements (GPU model, language model)
- Multi-graph support
- Flex-Algo integration

### SRv6 USID Generation

Automatic generation of SRv6 micro-SID lists:

- Auto-detects USID block from topology
- Algo-aware SID selection
- Compressed USID format output
- Full SID list for validation

## Development

### Running Tests

```bash
# Install test dependencies
pip install pytest pytest-asyncio

# Run tests
pytest
```

### Code Style

The project follows PEP 8 style guidelines. Format code with:

```bash
pip install black
black app/
```

## Troubleshooting

### Cannot connect to ArangoDB

- Verify ArangoDB is running and accessible
- Check environment variables for correct connection details
- In Kubernetes, ensure service DNS resolution is working

### API returns empty results

- Verify ArangoDB contains data
- Check collection names match your topology
- Ensure graph collections are properly configured

### SRv6 USID generation fails

- Verify nodes have SRv6 SIDs configured in the `sids` array
- Check that SIDs include the `algo` field
- Ensure USID block format matches your topology (e.g., `fc00:0:`)

## Contributing

See the main [Jalapeno contributing guide](../../docs/development/contributing.md) for details.

## License

See [LICENSE](../../LICENSE) for details.
