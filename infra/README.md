# Jalapeno Infrastructure
Jalapeno has the following components that create its infrastructure. All infrastructure components reside in the Jalapeno Kubernetes cluster and are deployed using the `deploy_infrastructure.py` script. 

The various Jalapeno infrastructure components define their deployments in their respective YAML files. These files allow for the configuration of services, deployments, config-maps, node-ports and other Kubernetes components.

## Kafka
Kafka is Jalapeno's message bus and core data handler. Jalapeno's Kafka instance handles two main types of data: OpenBMP data (topology information) and telemetry data (network metrics). Jalapeno Processors are responsible for reading and restructuring the data, and for inferring relevant metrics from the data.
OpenBMP data is organized into Kakfa topics such as openbmp.parsed.peer and openbmp.parsed.ls_node. These topics are further parsed to create representations of the network topology using the Topology Processor.
Telemetry data is collected in the openbmp.telemetry topic. Data in this topic is pushed into Telegraf (a telemetry consumer), and onwards into InfluxDB.
Kafka is deployed using `kubectl`, as seen in the `deploy_infrastructure.sh` script. The configurations for Kafka's deployment are in the various YAML files in the `jalapeno/infra/kafka/` directory.

## ArangoDB
ArangoDB is Jalapeno's graph database. Jalapeno Processors parse through data in Kafka and create various collections in ArangoDB that represent both the network's topology and current state. For example, the Topology Processor parses OpenBMP messages that have been streamed to Kafka and builds out collections such as "LSNode" and "L3VPNPrefix" in Jalapeno's ArangoDB instance. These collections, in conjunction with ArangoDBs rapid graphical traversals and calculations, make it easy to determine the lowest-latency path, etc. 
ArangoDB also houses the most interactive aspects of Jalapeno -- Processors for Bandwidth and Latency upsert their scores here, and clients query the ArangoDB Jalapeno API to generate label stacks for their desired network optimization attribute.
ArangoDB is deployed using `kubectl`, as seen in the `deploy_infrastructure.sh` script. The configurations for ArangoDB's deployment are in the various YAML files in the `jalapeno/infra/arangodb/` directory.  
To access ArangoDB's UI, log in at <server_ip>:30852, using credentials root/jalapeno. In the list of DBs, select jalapeno.

## InfluxDB
InfluxDB is Jalapeno's time-series database. Telemetry data is parsed and passed from Kafka into InfluxDB using a telemetry consumer (Telegraf). 
The data stored in InfluxDB can be queried by Processors and by the ArangoDB Jalapeno API. These queries construct and derive relevant metrics. For example, the LS-Performance Processor generates rolling-averages of bytes sent out of a router's interface -- thus simulating link utilization.
Using InfluxDB as a historical data-store, Jalapeno Processors can also infer trends based on historical analysis. Processors and even applications can determine whether instantaneous measurements are extreme anomalies, and can enable requests for any number of threshold-based reactions. 
InfluxDB is deployed using `kubectl`, as seen in the `deploy_infrastructure.sh` script. The configurations for InfluxDB's deployment are in the various YAML files in the `jalapeno/infra/influxdb/` directory.  
To access InfluxDB via Kubernetes, enter the pod's terminal and run:
```bash
influx
auth root jalapeno
use mdt_db
show series
```

## Grafana
Grafana is Jalapeno's visual dashboard and metric visualization tool. Loaded with InfluxDB as its data-source, Grafana has various graphical representations of the network, including historical bandwidth usage, historical latency metrics, and more. 
Grafana is deployed using `kubectl`, as seen in the `deploy_infrastructure.sh` script. The configurations for Grafana's deployment are in the various YAML files in the `jalapeno/infra/grafana/` directory.  
To access Grafana's UI, log in at <server_ip>:30300, using credentials root/jalapeno.

## Telegraf-Egress
Telegraf is a telemetry consumer and forwarder. In Infra, the Telegraf-Egress deployment of Telegraf consumes data from Kafka and forwards this data to InfluxDB.
Telegraf-Egress is deployed using `kubectl`, as seen in the `deploy_infrastructure.sh` script. The configurations for the Pipeline deployments are in the various YAML files in the `jalapeno/infra/telegraf-egress/` directory.
Note: there exists another Telegraf (Telegraf-Ingress) that is part of Jalapeno's Collectors and consumes data from devices directly before forwarding the data to Kafka. 
