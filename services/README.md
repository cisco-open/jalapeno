# Voltron Services

## Collector vServices
Voltron Collector vServices are responsible for organizing, parsing, and analysing network data. Any Voltron Infrastructure component with data is considered a source for a vService. 

### Topology vService
The Topology vService interacts with OpenBMP data in Kafka in order to create topology representations in ArangoDB.
Collections created using this service are considered base-collections. These base-collections have no inference of relationships between network elements, or of any metrics -- they are organized collections of individual OpenBMP messages.
For example, the Topology vService creates the InternalRouter collection, the BorderRouter collection, and the InternalLinkEdge collection directly from OpenBMP message data.
However, the inference that an InternalRouter can reach a BorderRouter through a path of three InternalLinkEdges is made using other Collector vServices.

The Topology vService is deployed using oc, as seen in the `deploy_collectors.sh` script in the collectors directory. 
The configuration for Topology's deployment is in "topology_dp.yaml" in the topology directory.

### EPEEdges vService
The EPEEdges vService uses base-collections in ArangoDB to craft the EPEEdges collection. EPEEdges are edges exiting the internal network out to the destination.
The source of an EPEEdge is a PeeringRouter, while the destination is an ExternalPrefix. Additional information is included such as ASN, InterfaceIP, SRNodeSID, and EPELabel.

The EPEEdges vService is deployed using oc, as seen in the `deploy_collectors.sh` script in the collectors directory. 
The configuration for EPEEdges's deployment is in "epe_edges_collector_dp.yaml" in epe-edges directory.

### EPEPaths vService
The EPEPaths vService uses the EPEEdges collection to generate the EPEPaths collection. The EPEPaths collection is similar to the EPEEdges collection.
The same source and destination seen in an EPEEdge will be seen in its corresponding EPEPath -- however, an EPEPath also derives a "label path" that can be used to reach the PeeringRouter and go out a specific interface.
A EPEPath label path is comprised of the SRNodeSID to reach the PeeringRouter and the EPELabel for a given interface.

The EPEPaths vService is deployed using oc, as seen in the `deploy_collectors.sh` script in the collectors directory. 
The configuration for EPEPaths's deployment is in "paths_collector_dp.yaml" in epe-paths directory.

### EgressLinks_Performance vService
The EgressLinks_Performance vService uses the PeeringRouterInterfaces collection to create the metric-oriented EgressLinks_Performance collection. 
Each document will derive link utiliation metrics from telemetry data in InfluxDB. This collection can then be used by the ArangoDB Voltron API to inform the client about various metric optimizations.

The EgressLinks_Performance vService is deployed using oc, as seen in the `deploy_collectors.sh` script in the collectors directory. 
The configuration for EgressLinks_Performance's deployment is in "egress_links_performance_collector_dp.yaml" in egress-links-performance directory.

### InternalLinks_Performance vService
The InternalLinks_Performance vService uses the InternalRouterInterfaces collection to create the metric-oriented InternalLinks_Performance collection. 
Each document will derive link utiliation metrics from telemetry data in InfluxDB. This collection can then be used by the ArangoDB Voltron API to inform the client about various metric optimizations.

The InternalLinks_Performance vService is deployed using oc, as seen in the `deploy_collectors.sh` script in the collectors directory. 
The configuration for InternalLinks_Performance's deployment is in "internal_links_performance_collector_dp.yaml" in internal-links-performance directory.

## API
The ArangoDB Voltron API has two core capabilities.
First, a client can query the API for the highest-available-bandwidth (or lowest link-utilization) EPEPath as described above. 
This query will return a label stack representing which PeeringRouter and interface the client should use to send data over.

Second, a client can also query the API for the lowest-latency EPEPath. 
An EPEPaths_Latency vService runs directly on the client. 
The EPEPaths_Latency vService creates the metric-oriented EPEPaths_Latency collection and has EPEPaths, but with latency scores and label paths associated with them.
This query will return a label stack representing which PeeringRouter and interface the client should use to send data over.
