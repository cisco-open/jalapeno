# ArangoDB

ArangoDB is Jalapeno's graph database that serves as the central data-store for all network topology and performance data. It acts as the single source of truth regarding network state for Jalapeno Processors and the Jalapeno API Gateway.

Nework devices are configured to send topology and performance data to Kafka. There, Jalapeno Processors parse through this data to create multiple "virtual topologies" that are stored in ArangoDB to represent the current state of the network.

For example, the [Topology Processor](../about/processors.md#topology-processor) parses GoBMP messages and builds out collections such as `LSNode` and `L3VPNPrefix` in Jalapeno's ArangoDB instance.

These collections, in conjunction with ArangoDBs rapid graphical traversals and calculations, make it easy to determine what path is optimal through the network given a specific SLA. For example, a user can request the `lowest-latency path` from point A to point B.

## Types of ArangoDB Interactions

- Parsing data from Kafka and storing in ArangoDB (i.e. [Topology Processor](../about/processors.md#topology-processor))

- Parsing data from ArangoDB and returning insights to client (i.e. API Gateway)

## Deploying ArangoDB

ArangoDB is deployed using `kubectl`, as seen in the `/install/infra/deploy_infrastructure.sh` script. The configurations for ArangoDB's deployment are in the YAML files in the `/infra/arangodb/` directory.  

## Accessing ArangoDB

To access ArangoDB's UI, log in at `<server_ip>:30852`, using credentials `root/jalapeno`. In the list of DBs, select `jalapeno`.

## Querying ArangoDB

Sample queries have been provided in the [Resources](../resources/arangodb.md) section.
