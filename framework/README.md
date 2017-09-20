# Voltron Framework

Framework for reading OpenBMP messages off a Kafka queue and creating the topology in the GraphDB.

Sample OpenBMP generated data is available [Here](./openbmp_parsed_data.txt). This data is generated from [this virtual topology](https://wwwin-github.cisco.com/raw/paulduda/voltron-network0/master/doc/voltron-network0.png) in the BXB lab.

## OpenBMP
The OpenBMP directory contains our OpenBMP message bus library. It should probably be submitted to OpenBMP maintainers to open source it. It parses messages according [to this spec](https://github.com/OpenBMP/openbmp/blob/master/docs/MESSAGE_BUS_API.md).

## main.go
main.go currently just connects to our lab machine's Kafka, pulls down messages from the beginning and prints them sorted by message type. Generically wrapping the functionality contained in main.go is the next step.

## WIP ArangoDB interfaces.
