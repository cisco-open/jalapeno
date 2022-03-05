## Jalapeno Data Processors
Jalapeno's Data Processors are responsible for organizing, parsing, and analysing network topology and performance data. Any Jalapeno Infrastructure component with data is considered a source for a Processor. 

### Topology Processor
BGP speakers send BMP data feeds to GoBMP, which then passes the data to Kafka.  The Topology Processor subscribes to Kafka's BMP topics in order to create topology representations in ArangoDB.
Collections created using this service are considered base-collections. These base-collections have no inference of relationships between network elements, or of any metrics -- they are organized collections of individual GoBMP messages.
For example, the Topology processor creates the LSNode collection and the LSLink collection directly from GoBMP BGP-LS message data.
 
The configuration for Topology's deployment is in "topology_dp.yaml" in the topology directory.




