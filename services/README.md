# Jalapeno Services

## Collector vServices
Jalapeno Collector vServices are responsible for organizing, parsing, and analysing network data. Any Jalapeno Infrastructure component with data is considered a source for a vService. 

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

### InternalLinks Performance vService
The InternalLinks Performance vService calculates and correlates performance metrics to InternalLinkEdges and InternalRouterInterfaces.
Each document will derive link utiliation metrics from telemetry data in InfluxDB. This collection can then be used by the ArangoDB Jalapeno API to inform the client about various metric optimizations.

The InternalLinks Performance vService is deployed using oc, as seen in the `deploy_collectors.sh` script in the collectors directory. 
The configuration for InternalLinks Performance's deployment is in "internal_links_performance_collector_dp.yaml" in internal-links-performance directory.

### ExternalLinks Performance vService
The ExternalLinks Performance vService calculates and correlates performance metrics to ExternalLinkEdges and BorderRouterInterfaces.
Each document will derive link utiliation metrics from telemetry data in InfluxDB. This collection can then be used by the ArangoDB Jalapeno API to inform the client about various metric optimizations.

The ExternalLinks Performance vService is deployed using oc, as seen in the `deploy_collectors.sh` script in the collectors directory. 
The configuration for ExternalLinks Performance's deployment is in "external_links_performance_collector_dp.yaml" in external-links-performance directory.

## API
The API is deployed using Swagger and enables the client to make a variety of requests for optomized paths through the network. 
