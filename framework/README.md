# Voltron Framework

Framework for reading OpenBMP messages off a Kafka queue and creating the topology in the GraphDB.

Sample OpenBMP generated data is available [Here](./openbmp_parsed_data.txt). This data is generated from [this virtual topology](https://wwwin-github.cisco.com/raw/paulduda/voltron-network0/master/doc/voltron-network0.png) in the BXB lab.

# Getting Started
Arango

```
git clone https://wwwin-github.cisco.com/spa-ie/voltron-redux.git
cd framework
make
sh arango/deploy.sh
```

We support two run modes for this binary, the first is to populate arango with a topology. This mode will read from kafka and parse BMP messages into a network topology. Running this you would run. (this example also includes a configuration)

`./bin/voltron topology --config sample.yaml`


The other run mode is to simply run the framework. This provides an API that lets CvServices register and enables CRUD operations on those objects being stored in the arangoDB.

`./bin/voltron framework --config sample.yaml`

Need help?
`./bin/voltron -h` OR `./bin/voltron topology -h` OR `./bin/voltron framework -h`


# Running tests
1. Deploy the database `./framework/arango/deploy.sh` (give it ~10 seconds to start up)
2. `make test`
3. Stop the database `./framework/arango/stop.sh`

# Graph Data Model

## Nodes
Two types of nodes:
```
Router {
  BGPID String
  ASN Int
  RouterIP String (like an id, BMP uses the loopback addr or highest IP)
  IsLocal Bool
}
```

```
Prefix {
  Prefix String
  Length Int
}
```

## Edges Between Routers
```
LinkEdge {
  Labels []int
  Latency
  Utilization
  From (RouterKey)
  To (RouterKey)
  FromIP
  ToIP
  Netmask
}
```

## Edges Between Router & Prefix
```
PrefixEdge {
  InterfaceIP
  Labels []int
  NextHop
  ASPath []string
  BGPPolicy ???
  Latency ??? (do we want this here... or is it sufficient on other edges?)
}
```

## [Accessing Arango via HTTP](https://docs.arangodb.com/3.2/HTTP/SimpleQuery/)
This can be done without modification to the arango db or working with the [Foxx](https://docs.arangodb.com/3.2/HTTP/Foxx/) services. In the future once we have specific queries that are common we can expose a Foxx service that would be a conveneince for vServices. Foxx could also support more advanced business logic once we know that that looks like, but for alpha I'd recommend we don't write these javascript services until we know what queries are common / business logic we require.

### Basic HTTP Things
To get a specific document the query looks like:
```
curl http://<user>:<pass>@<arango-end-point>:<arango-port>/_db/<db_name>/_api/<collection_name>/<document_key>
```

In this example we are running arango locally from the [voltron-redux/framework/database/deploy.sh](https://wwwin-github.cisco.com/spa-ie/voltron-redux/blob/master/framework/arango/deploy.sh). We’ve parsed BMP messages from Kafka and the result is:
```
http://root:voltron@127.0.0.1:8529/_db/test/_api/document/Routers/10.1.1.4_100000
```

Where:
- `db-name="test"`
- `collection-name="Routers"`
- `document-key="10.1.1.4_100000"`


The API also supports queries, you can do simple queries where you write out the AQL:
```
curl -X POST --data-binary @body.json --dump - http://root:vojltorb@127.0.0.1:8529/_db/test/_api/explain
```

**body.json**
```
{
  "query" : "FOR p IN Routers RETURN p"
}
```

This example query will return all routers in arango.

You can also query that limits by fields
```
curl -X PUT --data-binary @body.json --dump - http://root:vojltorb@localhost:8529/_db/test/_api/simple/by-example
```

**body.json**
```
{
  "collection" : "Routers",
  "example" : {
    "ASN" : 8000
  }
}
```

CRUD operations are available. To see those examples go to https://docs.arangodb.com/3.2/HTTP/SimpleQuery/

## Query Microservice
Arango allows hosted javascript microservices to be mounted as a sub URL.
You can write common queries in the arango/queries/index.js file. This microservice can be added to arango and queried directly. To try this out yourself:
- go to [the services tab](http://127.0.0.1:8529/_db/voltron/_admin/aardvark/index.html#services) of arango
- click "Add Service"
- Enter /queries in the "Mount" field
- click the "zip" tab and select `arango/queries.zip`
- [example]: Go to http://127.0.0.1:8529/_db/voltron/latency/edges/10.1.1.3_100000/ips and you should see all the ip address of the interfaces on the router with `_key=10.1.1.3_100000`

## Latency (LinkEdge) Query
I added an endpoint to our Arango in the BXB lab for editing latency. You can add it yourself by following the instructions above.

The endpoint is `/queries/linkedges/:from/:to/:latency` where :from = FromIP, :to=ToIP, :latency=latency in ms
To add/update the latency (to 20ms) between 10.1.1.1 and 10.1.1.2, hit the following url (will update the values on the lab arango):

`http://10.86.204.8:8529/_db/voltron/queries/linkedges/10.1.1.1/10.1.1.2/20`

(Note this most likely change to a PUT request in the future.)
You can check the latency by going to `http://10.86.204.8:8529/_db/voltron/queries/linkedges/10.1.1.1/10.1.1.2` (will return 0 if not set).

## Latency (PrefixEdge) Query
The French Press demo will add latency on PrefixEdges. The end point is similarly `/queries/prefixedges/:interfaceIP/:prefix/:latency`
If router p3 knows of a prefix `10.11.170.0/23` that is available through it's peer `2.2.2.15` with a latency of 20ms, the following request is used:
`http://10.86.204.8:8529/_db/voltron/queries/prefixedges/2.2.2.15/10.11.170.0_23/20`.

# Directory Structure
## openbmp/
The OpenBMP directory contains our OpenBMP message bus library. It should probably be submitted to OpenBMP maintainers to open source it. It parses messages according [to this spec](https://github.com/OpenBMP/openbmp/blob/master/docs/MESSAGE_BUS_API.md).

## kafka/
Kafka Consumer implementation. Reads off the message bus and hands off to a handler.

## kafka/handler
Handlers do _something_ with OpenBMP messages. This contains the interface for handlers (as well as a default handler implementation for debugging.)

## arango/
Contains our arango implementation and arangodb handler. **arango/handler.go** does most of the hard work of translating the openBMP messages --> arangodb.
