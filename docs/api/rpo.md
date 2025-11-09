# Resource Path Optimization (RPO) API

## Overview
The Resource Path Optimization (RPO) API provides intelligent destination selection based on metrics, combined with shortest path calculation and SRv6 USID generation.

## Base URL
```
http://localhost:8000/api/v1/rpo
```

---

## 1. Discovery and Information

### Get RPO capabilities and available graphs
```bash
curl "http://localhost:8000/api/v1/rpo"
```

**Response includes:**
- Supported metrics and optimization strategies
- Available graph collections for path finding
- API description and usage notes

---

## 2. Collection Management

### List all endpoints in a collection
```bash
curl "http://localhost:8000/api/v1/rpo/hosts"
```

### List endpoints with limit
```bash
curl "http://localhost:8000/api/v1/rpo/hosts?limit=5"
```

---

## 3. Optimal Endpoint Selection (from all endpoints)

### Select endpoint with lowest GPU utilization
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=gpu_utilization&graphs=ipv6_graph"
```

### Select endpoint with lowest CPU utilization
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=cpu_utilization&graphs=ipv6_graph"
```

### Select endpoint with lowest time to first token
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=time_to_first_token&graphs=ipv6_graph"
```

### Select endpoint with lowest cost per million tokens
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=cost_per_million_tokens&graphs=ipv6_graph"
```

### Select endpoint with specific GPU model (exact match)
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=gpu_model&value=GB300&graphs=ipv6_graph"
```

### Select endpoint with specific language model (exact match)
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=language_model&value=llama-3-70b&graphs=ipv6_graph"
```

---

## 4. Selection from Specific List

### Select from specific destinations (lowest GPU utilization)
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_utilization&graphs=ipv6_graph"
```

### Select from specific destinations (lowest cost)
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s,hosts/london&metric=cost_per_million_tokens&graphs=ipv6_graph"
```

### Select from specific destinations (lowest response time)
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=response_time&graphs=ipv6_graph"
```

### Select from specific destinations (highest capacity)
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=available_capacity&graphs=ipv6_graph"
```

---

## 5. Different Graph Collections

### Using IPv4 graph
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_utilization&graphs=ipv4_graph"
```

### Using IPv6 graph
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_utilization&graphs=ipv6_graph"
```

### Using IGPv4 graph
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_utilization&graphs=igpv4_graph"
```

### Using fabric graph
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_utilization&graphs=fabric_graph"
```

---

## 6. Different Direction Options

### Outbound direction (default)
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_utilization&graphs=ipv6_graph&direction=outbound"
```

### Inbound direction
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_utilization&graphs=ipv6_graph&direction=inbound"
```

---

## 7. Flex-Algo Support

### Select endpoint with default algo (algo 0)
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=gpu_utilization&graphs=ipv6_graph"
```

### Select endpoint using Flex-Algo 128
```bash
curl "http://localhost:8000/api/v1/rpo/bgp_prefix_v4/select-optimal?source=bgp_prefix_v4/10.17.1.0_24&metric=gpu_utilization&graphs=ipv4_graph&algo=128" | jq
```

### Select endpoint using Flex-Algo 129
```bash
curl "http://localhost:8000/api/v1/rpo/bgp_prefix_v4/select-optimal?source=bgp_prefix_v4/10.17.1.0_24&metric=gpu_utilization&graphs=ipv4_graph&algo=129" | jq
```

### Select from list with Flex-Algo 128
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_utilization&graphs=ipv6_graph&algo=128"
```

### Low latency path with Flex-Algo 128
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=time_to_first_token&graphs=ipv6_graph&algo=128"
```

### Cost optimization with specific algo
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s,hosts/london&metric=cost_per_million_tokens&graphs=ipv6_graph&algo=129"
```

---

## 8. Complex Scenarios

### Multi-destination selection with cost optimization
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s,hosts/london,hosts/paris&metric=cost_per_hour&graphs=ipv6_graph"
```

### Memory utilization optimization
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-optimal?source=hosts/rome&metric=memory_utilization&graphs=ipv6_graph"
```

### Specific model requirement with fallback optimization
```bash
curl "http://localhost:8000/api/v1/rpo/hosts/select-from-list?source=hosts/rome&destinations=hosts/amsterdam,hosts/berlin-k8s&metric=gpu_model&value=V100&graphs=ipv6_graph"
```

---

## Response Format

All endpoints return comprehensive information including:

```json
{
  "collection": "hosts",
  "source": "hosts/rome",
  "selected_endpoint": {
    "_id": "hosts/amsterdam",
    "name": "amsterdam",
    "gpu_utilization": 0.3,
    "cost_per_million_tokens": 4,
    "time_to_first_token": 2
  },
  "optimization_metric": "gpu_utilization",
  "metric_value": 0.3,
  "optimization_strategy": "minimize",
  "total_candidates": 2,
  "valid_endpoints_count": 2,
  "path_result": {
    "found": true,
    "path": [...],
    "hopcount": 7,
    "srv6_data": {
      "srv6_sid_list": ["fc00:0:7777::", "fc00:0:6666::", "fc00:0:2222::", "fc00:0:1111::"],
      "srv6_usid": "fc00:0:7777:6666:2222:1111::"
    }
  },
  "summary": {
    "destination": "hosts/amsterdam",
    "destination_name": "amsterdam",
    "path_found": true,
    "hop_count": 7
  }
}
```

---

## Supported Metrics

| Metric | Type | Optimization Strategy | Description |
|--------|------|----------------------|-------------|
| `cpu_utilization` | numeric | minimize | CPU usage percentage |
| `gpu_utilization` | numeric | minimize | GPU usage percentage |
| `memory_utilization` | numeric | minimize | Memory usage percentage |
| `time_to_first_token` | numeric | minimize | Response time in seconds |
| `cost_per_million_tokens` | numeric | minimize | Cost per million tokens |
| `cost_per_hour` | numeric | minimize | Hourly cost |
| `gpu_model` | string | exact_match | Specific GPU model required |
| `language_model` | string | exact_match | Specific language model required |
| `available_capacity` | numeric | maximize | Available processing capacity |
| `response_time` | numeric | minimize | General response time |

---

## Notes

- **Required Parameters**: `source`, `metric`, `graphs`
- **Optional Parameters**: 
  - `value` (required for exact_match metrics)
  - `direction` (default: outbound)
  - `algo` (Flex-Algo ID, default: 0)
- **Graph Collections**: Use the discovery endpoint to see available graphs
- **Flex-Algo**: Specify `algo` parameter to use specific Flex-Algo for path finding
  - Default is algo 0 (standard SPF)
  - Common values: 128 (low latency), 129 (high bandwidth), etc.
  - Path will only traverse nodes that participate in the specified algo
  - SRv6 SIDs are automatically selected based on the specified algo
- **SRv6 USID**: Generated automatically for path execution based on algo
- **Error Handling**: Comprehensive error messages for invalid parameters or missing data

