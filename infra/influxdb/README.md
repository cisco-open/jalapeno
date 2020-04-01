# InfluxDB

InfluxDB is Jalapeno's time-series database.

Network devices are configured to send performance data to Kafka. There, the data is parsed using Telegraf (a telemetry consumer) and stored in InfluxDB. The data stored in InfluxDB can be queried by Jalapeno Processors.

These queries construct and derive relevant metrics that inform Jalapeno's API Gateway responses to client requests. For example, the [LS-Performance Processor](../../processors/ls-performance/) generates rolling-averages of bytes sent out of a router's interface -- thus simulating link utilization. Those calculated insights are then inserted into ArangoDB and associated with their corresponding edges, providing a holistic view of the current state of the network.

Using InfluxDB as a historical data-store, trends can also be inferred based on historical analysis. Processors and applications can determine whether instantaneous measurements are extreme anomalies, and can enable requests for any number of threshold-based reactions. 

## Deploying InfluxDB
InfluxDB is deployed using `kubectl`, as seen in the [deploy_infrastructure script](../deploy_infrastructure.sh). The configurations for InfluxDB's deployment are in the various YAML files in the [jalapeno/infra/influxdb/](.) directory.

## Accessing InfluxDB
To access InfluxDB via Kubernetes, enter the pod's terminal and run:
```
influx
auth root jalapeno
use mdt_db
show series
```
