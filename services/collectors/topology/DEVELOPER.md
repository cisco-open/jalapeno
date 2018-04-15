# Developer Guide
Arango

```
git clone https://wwwin-github.cisco.com/spa-ie/voltron-redux.git
cd framework
make
sh database/deploy.sh
```

We support two run modes for this binary, the first is to populate arango with a
topology. This mode will read from kafka and parse BMP messages into a network topology.
Running this you would run. (this example also includes a configuration)

`./bin/voltron topology --config sample.yaml`


The other run mode is to simply run the framework. This provides an API that lets
CvServices register and enables CRUD operations on those objects being stored in the arangoDB.

`./bin/voltron framework --config sample.yaml`

Need help?
`./bin/voltron -h` OR `./bin/voltron topology -h` OR `./bin/voltron framework -h`


# Running tests
1. Deploy the database `./framework/database/deploy.sh` (give it ~10 seconds to start up)
2. `make test`
3. Stop the database `./framework/database/stop.sh`


## [Accessing Arango via HTTP](https://docs.arangodb.com/3.2/HTTP/SimpleQuery/)
If the Arango DB was deployed with the script in `database/`, the credentials are
`root/voltron`

# Directory Structure

## api/
Contains all the REST APIs framework exposes for Responders and Collectors. We
leverage swagger to generate docs and clients.

## database/
Contains our arango wrapper that lets the framework interact with the database.

## handler/
This directory does most of the heavy lifting in translating OpenBMP messages
into the graph data model.

## kafka/
The kafka directory handles pulling messages out of kafka and passing that to a
handler, in our initial implementation the handler is arango_handler.

## manager/
The manager watches Collectors and tracks their state. If a heartbeat was received
before timeout, it is kept in the "Running" state, if it does not it is places in
the "Down" state.

## openbmp/
The OpenBMP directory contains our OpenBMP message bus library. It should probably
be submitted to OpenBMP maintainers to open source it. It parses messages according [to this spec](https://github.com/OpenBMP/openbmp/blob/master/docs/MESSAGE_BUS_API.md).
