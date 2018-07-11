# Topology Collector vService

This service creates real-time representations of the network topology in Voltron's ArangoDB instance.
This is used to represent the physical and metric states of the network. Thus, Voltron can create paths
from a given source to destination (using another collector, the Paths Collector vService), and can
attribute latency and bandwidth metrics to both individual links and whole-paths.

The flow of the topology service is as follows:
* Create a collection type in ArangoDB
* Read and parse messages from Kafka's various OpenBMP topics
* Pass each message into a handler, which maps the message-type to its specific function
* Filter and do calculations on the current message's fields in said function
* Create the document for the pertinent collection with the fields set
* Upsert the document into the collection in ArangoDB

#
### Directory Structure

### database/
The database directory contains the interfaces and structure definitions for Voltron's topology collections in ArangoDB.
To modify existing collections or create new collections, see database/README.md.

### handler/
The handler directory contains handling for each type of OpenBMP message, i.e. unicast-prefix messages.
To modify how data is handled or what topology documents are upserted into Voltron's collections in ArangoDB, see handler/README.md

### kafka/
The kafka directory handles the transfer of OpenBMP messages in Kafka to the Voltron handler listed above.
To see how Kafka breaks down messages for handling, see kafka/README.md

### openbmp/
The OpenBMP directory contains the OpenBMP message bus library.
Currently, messages are parsed according to the spec [here](https://github.com/OpenBMP/openbmp/blob/master/docs/MESSAGE_BUS_API.md).
