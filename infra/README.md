# Voltron Infrastructure
Voltron has the following components that create its infrastructure. A majority of the components are deployed using the oc OpenShift CLI, and thus reside as part of the OpenShift cluster. Some components are run locally in their own containers, as noted below. All components of the Voltron infrastructure are deployed using the `deploy_infrastructure.py` script. 

In the various YAML configuration files for each of the infrastructure components, there are references to IP addresses and ports representing the OpenShift server, deployment node-ports, etc. In order to update these files to reflect your current network and environment, run the `configure_voltron.py` script in the `voltron/` directory. 

## Kafka
Kafka is Voltron's message bus. Kafka receives data from various sources, from topology data to telemetry data. Voltron services are responsible for reading and restructuring the data, and for inferring relevant metrics from the data. Kafka is deployed using oc, as seen in the `deploy_infrastructure.py` script. The configurations for Kafka's deployment are in the various YAML files in the `voltron/infra/kafka/` directory.

## ArangoDB
ArangoDB is Voltron's graph-database. Voltron services parse through data in Kafka and create various collections in ArangoDB that represent both the network's topology and current state. For example, the Topology vService parses OpenBMP messages that have been streamed to Kafka, and builds out collections such as "routers" and "internal_prefix_edges" in ArangoDB. These collections, in conjunction with ArangoDBs rapid graphical traversals and calculations, make it easy to determine the lowest-latency path, etc. ArangoDB is deployed using oc, as seen in the `deploy_infrastructure.py` script. The configurations for ArangoDB's deployment are in the various YAML files in the `voltron/infra/arangodb/` directory.

## InfluxDB
InfluxDB is Voltron's time-series database. Telemetry data is parsed and passed from Kafka into InfluxDB using a telemetry consumer (Pipeline). InfluxDB thus receives pertinent metrics, such as available bandwidth over a specific interface on router1, over time. Using this historical data-store, Voltron services can infer trends based on historical analysis. They can also determine whether instantaneous measurements are extreme anomalies, and enable requests for any number of threshold-based reactions. InfluxDB is deployed using oc, as seen in the `deploy_infrastructure.py` script. The configurations for InfluxDB's deployment are in the various YAML files in the `voltron/infra/influxdb/` directory.

## Grafana
Grafana is Voltron's visual dashboard. Loaded with InfluxDB as its data-source, Grafana has various graphical representations of the network, including historical bandwidth usage, historical latency metrics, and more. Grafana is deployed using oc, as seen in the `deploy_infrastructure.py` script. The configurations for Grafana's deployment are in the various YAML files in the `voltron/infra/grafana/` directory.

## OpenBMP
OpenBMP is Voltron's topology collector. OpenBMP is run as a local container rather than as part of the larger OpenShift deployment. This containerized consumer receives OpenBMP data from every router configured to send OpenBMP data in the network. It then passes the data onto Kafka, where Voltron services can infer relationships and create representations of the network. The `voltron/infra/openbmpd` directory also comes with a `hosts.json` file that can be modified to reflect the network. Upon running the `configure_openbmp.py` script, each device in the `hosts.json` file would be configured with the included `openbmp_config_xr` config, and thus would send data to the OpenBMP container. 

## Telemetry
Telemetry is much more of a multi-step configuration deployment than a piece of infrastructure deployed in OpenShift. First, the included `hosts.json` file must reflect the various nodes in the network where telemetry needs to be streamed from. Telemetry is then configured on these device using the `configure_telemetry.py` script, as the telemetry configuration in `config_xr` is set. Next, guest shell is enabled on each device using the `enable_guestshell.py` script to give Voltron bash access. Finally, Pipeline, a telemetry consumer/forwarder, is placed on each device and is started through the `manage_pipeline.py` script. Thus, telemetry is streaming from each device to Pipeline on-box, which is sending the data onwards to Kafka. 

## Pipeline
Pipeline is a telemetry consumer and forwarder. While instances of Pipeline sit on-box in the network, a deployment of Pipeline also sits in the larger OpenShift deployment in order to consume telemetry data after the data has gotten to Kafka. The data is then forwarded once more to InfluxDB. This Pipeline deployment is deployed using oc, as seen in the `deploy_infrastructure.py` script. The configurations for Pipeline's deployment are in the various YAML files in the `voltron/infra/pipeline/` directory.


