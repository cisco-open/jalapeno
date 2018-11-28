# Voltron Infrastructure
Voltron has the following components that create its infrastructure. A majority of the components are deployed using the oc OpenShift CLI, and thus reside as part of the OpenShift cluster. Some components are run locally in their own containers, as noted below. All components of the Voltron infrastructure are deployed using the `deploy_infrastructure.py` script. 
The components deployed in OpenShift (Kafka, ArangoDB, InfluxDB, Grafana, Pipeline) are deployed using YAML files. These files allow for the configuration of services, deployments, config-maps, node-ports and other OpenShift components.
The two components not deployed in OpenShift are OpenBMP and Telemetry.
OpenBMP is deployed in a container directly on the CentosKVM (enabling accessibility for OpenBMP instances on routers), while Telemetry is configured directly on the routers.

## Kafka
Kafka is Voltron's message bus and core data handler. Voltron's Kafka instance handles two main types of data: OpenBMP data (topology information) and telemetry data (network metrics). Voltron vServices are responsible for reading and restructuring the data, and for inferring relevant metrics from the data.
OpenBMP data is organized into Kakfa topics such as openbmp.parsed.peer and openbmp.parsed.ls_node. These topics are further parsed to create representations of the network topology using the Topology Collector vService.
Telemetry data is collected in the openbmp.telemetry topic. Data in this topic is pushed into Pipeline (a telemetry consumer), and onwards into InfluxDB.
Kafka is deployed using oc, as seen in the `deploy_infrastructure.sh` script. The configurations for Kafka's deployment are in the various YAML files in the `voltron/infra/kafka/` directory.

## ArangoDB
ArangoDB is Voltron's graph database. Voltron vServices parse through data in Kafka and create various collections in ArangoDB that represent both the network's topology and current state. For example, the Topology vService parses OpenBMP messages that have been streamed to Kafka and builds out collections such as "Routers" and "ExternalPrefixes" in Voltron's ArangoDB instance. These collections, in conjunction with ArangoDBs rapid graphical traversals and calculations, make it easy to determine the lowest-latency path, etc. 
ArangoDB also houses the most interactive aspects of Voltron -- vServices for Bandwidth and Latency upsert their scores here, and clients query the ArangoDB Voltron API to generate label stacks for their desired network optimization attribute.
ArangoDB is deployed using oc, as seen in the `deploy_infrastructure.sh` script. The configurations for ArangoDB's deployment are in the various YAML files in the `voltron/infra/arangodb/` directory.  
To access ArangoDB's UI, log in at <server_ip>:30852, using credentials root/voltron. In the list of DBs, select voltron.

## InfluxDB
InfluxDB is Voltron's time-series database. Telemetry data is parsed and passed from Kafka into InfluxDB using a telemetry consumer (Pipeline). 
The data stored in InfluxDB can be queried by Collector vServices and by the ArangoDB Voltron API. These queries construct and derive relevant metrics. For example, the Bandwidth vService generates rolling-averages of bytes sent out of a router's interface -- thus simulating link utilization.
Using InfluxDB as a historical data-store, Voltron vServices can also infer trends based on historical analysis. vServices can determine whether instantaneous measurements are extreme anomalies, and can enable requests for any number of threshold-based reactions. 
InfluxDB is deployed using oc, as seen in the `deploy_infrastructure.sh` script. The configurations for InfluxDB's deployment are in the various YAML files in the `voltron/infra/influxdb/` directory.  
To access InfluxDB via OpenShift, enter the pod's terminal and run:
```bash
influx
auth voltron voltron
use mdt_db
show series
```

## Grafana
Grafana is Voltron's visual dashboard and metric visualization tool. Loaded with InfluxDB as its data-source, Grafana has various graphical representations of the network, including historical bandwidth usage, historical latency metrics, and more. 
Grafana is deployed using oc, as seen in the `deploy_infrastructure.sh` script. The configurations for Grafana's deployment are in the various YAML files in the `voltron/infra/grafana/` directory.  
To access Grafana's UI, log in at <server_ip>:30300, using credentials voltron/voltron.

## OpenBMP
OpenBMP is Voltron's topology collector. OpenBMP is run as a local container rather than as part of the larger OpenShift deployment. This containerized consumer receives OpenBMP data from every router configured to send OpenBMP data in the network. It then passes the data onto Kafka, where Voltron vServices can infer relationships and create representations of the network. The `voltron/infra/openbmpd` directory also comes with a `hosts.json` file that can be modified to reflect the network. Upon running the `configure_openbmp.py` script, each device in the `hosts.json` file would be configured with the included `openbmp_config_xr` config, and thus would send data to the OpenBMP container. 

## Telemetry
Telemetry is much more of a multi-step configuration deployment than a component of infrastructure deployed in OpenShift. First, the included `hosts.json` file must reflect the various nodes in the network where telemetry needs to be streamed from. Telemetry is then configured on these device using the `configure_telemetry.py` script, as the telemetry configuration in `config_xr` is set. Next, guest shell is enabled on each device using the `enable_guestshell.py` script to give Voltron bash access. Finally, Pipeline, a telemetry consumer/forwarder, is placed on each device and is started through the `manage_pipeline.py` script. 
Thus, telemetry is streaming from each device to Pipeline on-box, which is sending the data onwards to Kafka. 

## Pipeline
Pipeline is a telemetry consumer and forwarder. While instances of Pipeline sit on-box in the network, a deployment of Pipeline also sits in the larger OpenShift deployment in order to consume telemetry data after the data has gotten to Kafka. The data is then forwarded once more to InfluxDB. 
This Pipeline deployment is deployed using oc, as seen in the `deploy_infrastructure.sh` script. The configurations for Pipeline's deployment are in the various YAML files in the `voltron/infra/pipeline/` directory.

