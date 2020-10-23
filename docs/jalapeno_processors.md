## Jalapeno Data Processors
Jalapeno's Data Processors are responsible for organizing, parsing, and analysing network topology and performance data. Any Jalapeno Infrastructure component with data is considered a source for a Processor. 

### Topology Processor
BGP speakers send BMP data feeds to GoBMP, which then passes the data to Kafka.  The Topology Processor subscribes to Kafka's BMP topics in order to create topology representations in ArangoDB.
Collections created using this service are considered base-collections. These base-collections have no inference of relationships between network elements, or of any metrics -- they are organized collections of individual OpenBMP messages.
For example, the Topology vService creates the LSNode collection and the LSLink collection directly from OpenBMP BGP-LS message data.
However, the inference that an LSNode can reach another LSNode via some set of LSLinks is made using the separate link-state topology processor.
 
The configuration for Topology's deployment is in "topology_dp.yaml" in the topology directory.
## Demo Processors

https://github.com/jalapeno/demo-processors

### EPE_Topology - Egress Peer Engineering for Internal to External Traffic Engineering
The EPE_Topology processor uses EPENode, EPELink, and EPEPrefix base-collections in ArangoDB to creat the EPE_Topology edge collection. This edge collection is a virtual topology representation of egress paths from an internal network to external (Internet) prefixes.
The source of an EPEEdge is a PeeringRouter, while the destination is an ExternalPrefix. Additional information is included such as ASN, InterfaceIP, SRNodeSID, and EPELabel.

The configuration for EPE_Topology deployment is in "epe_topology_dp.yaml" in the epe-topology directory.

### LSLink Performance Processor
The LSLink Performance Processor calculates and correlates performance metrics to link-state interfaces and populates metric data in the LS_Topology edge collection.
Each document will derive link utiliation metrics from telemetry data in InfluxDB. 

The configuration for LSLink Performance's deployment is in "lslink_performance_dp.yaml" in the lslinks-performance directory.

### EPELink Performance Processor
The EPELink Performance Processor calculates and correlates performance metrics on EPELinks and populates metric data in the EPE_Topology edge collection.
Each document will derive link utiliation metrics from telemetry data in InfluxDB. 

The configuration for EPELink Performance's deployment is in "epeink_performance_dp.yaml" in the epelink-performance directory.

## API
The API is deployed using Swagger and enables the client to make a variety of requests for optomized paths through the network. 
