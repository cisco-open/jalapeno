# Jalapeno Data Processors

Jalapeno's Data Processors are responsible for organizing, parsing, and analysing network topology and performance data.

Any Jalapeno Infrastructure component with data is considered a source for a Processor.

## GoBMP-Arango Processor

BGP speakers send BMP data feeds to GoBMP, which then passes the data to Kafka.  The GoBMP-Arango Processor subscribes to Kafka's BMP topics in order to create topology representations in ArangoDB.

Collections created using this service are considered base-collections. These base-collections have no inference of relationships between network elements, or of any metrics - they are organized collections of individual [GoBMP](https://github.com/sbezverk/gobmp) messages.

For example, the GoBMP-Arango processor creates the LSNode collection and the LSLink collection directly from GoBMP BGP-LS message data.

The configuration for gobmp-arango deployment is in "gobmp-arango_dp.yaml" in the gobmp-arango directory.

## Other Processors

Currently the project is bundled with a limited set of processors. However, other processors can be found in [this repository](https://github.com/orgs/jalapeno/repositories) which may offer additional functionality to Jalapeno.
