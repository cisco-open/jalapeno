# InfluxDB

InfluxDB is Jalapeno's time-series database.

Network devices are configured to send performance data to Kafka. Then data is parsed using Telegraf (a telemetry consumer) and stored in InfluxDB. The data stored in InfluxDB can be queried by [Jalapeno Processors](../about/processors.md).

These queries construct and derive relevant metrics that inform Jalapeno's API Gateway responses to client requests. For example, a processor could generate rolling-averages of bytes sent out of a router's interface, simulating link utilization. Those calculated insights could then be inserted into ArangoDB and associated with their corresponding edges, providing a holistic view of the current state of the network.

Using InfluxDB as a historical data-store, trends can also be inferred based on historical analysis. Processors and applications can determine whether instantaneous measurements are extreme anomalies, and can enable requests for any number of threshold-based reactions.

## Deploying InfluxDB

InfluxDB is deployed using `kubectl`, as seen in the `/install/infra/deploy_infrastructure.sh`. The configurations for InfluxDB's deployment are in the various YAML files in the `/infra/influxdb/` directory.

## Accessing InfluxDB

To access InfluxDB via Kubernetes, enter the pod's terminal:
```
kubectl exec -it -n jalapeno influxdb-0 -- sh
```

and run:

```bash
influx
auth root jalapeno
use mdt_db
show series #(1)!
```

1. Provides a list of all time-series in the mdt_db

## Querying InfluxDB

Sample queries have been provided in the [Resources](../resources/influxdb.md) section.
