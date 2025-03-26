# Installing Jalapeno with Helm

## Prerequisites

- Kubernetes cluster
- Helm 3.x installed

## Installation

1. Clone the Jalapeno repository:
   ```bash
   git clone https://github.com/cisco-open/jalapeno.git
   cd jalapeno
   ```

2. Install the dependencies:
   ```bash
   helm dependency update ./charts/jalapeno
   ```

3. Install Jalapeno with default settings:
   ```bash
   helm install jalapeno ./charts/jalapeno
   ```

3. Install with custom values:
   ```bash
   helm install jalapeno ./charts/jalapeno --values custom-values.yaml
   ```

4. Install with minimal values (see example below):
   ```bash
   helm install jalapeno ./charts/jalapeno --values minimal-values.yaml
   ```

## Configuration

The following table lists the configurable parameters of the Jalapeno chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `arangodb.dbName` | Name of the ArangoDB database | `"jalapeno"` |
| `arangodb.service.nodePort` | NodePort for ArangoDB service | `30529` |
| `collectors.gobmp.service.nodePort` | NodePort for GoBMP collector | `30500` |
| ... | ... | ... |

For a complete list of parameters, see the [values.yaml](../../charts/jalapeno/values.yaml) file.

## Examples

### Minimal Installation

```yaml
minimal-values.yaml
collectors:
telegraf:
enabled: false
processors:
linkstateEdge:
enabled: false
grafana:
enabled: false


