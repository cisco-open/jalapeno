# Voltron Framework

Framework for reading OpenBMP messages off a Kafka queue and creating the topology in the GraphDB.

Sample OpenBMP generated data is available [Here](./openbmp_parsed_data.txt). This data is generated from [this virtual topology](https://wwwin-github.cisco.com/raw/paulduda/voltron-network0/master/doc/voltron-network0.png) in the BXB lab.

## openbmp/
The OpenBMP directory contains our OpenBMP message bus library. It should probably be submitted to OpenBMP maintainers to open source it. It parses messages according [to this spec](https://github.com/OpenBMP/openbmp/blob/master/docs/MESSAGE_BUS_API.md).

## kafka/
Kafka Consumer implementation. Reads off the message bus and hands off to a handler.

## kafka/handler
Handlers do _something_ with OpenBMP messages. This contains the interface for handlers (as well as a default handler implementation for debugging.)

## arango/
Contains our arango implementation and arangodb handler. **arango/handler.go** does most of the hard work of translating the openBMP messages --> arangodb.

# Getting Started
```
git clone https://wwwin-github.cisco.com/spa-ie/voltron-redux.git
cd framework
make
bin/framework --config sample.yaml
```

Alternatively, framework can be run in debug mode to print current messages/stats. Assuming a kafka broker exists at 10.86.204.8:9092:
```
bin/framework --debug --kafka-brokers 10.86.204.8:9092
```
