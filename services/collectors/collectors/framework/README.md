# Voltron Framework

The framework provides a consistent method for Services to add scores, nodes and
edges to the graph database. The framework is made up of 2 components, the first
is the Framework API. You can find the swagger yaml of the API in
`framework/api/v1/gen/v1.yaml`. This API supports reads and writes to the database.
Long term goals are to have this API provided authorization for different services
so each can have minimum amount of permissions when accessing the database. The
second component of the framework is the topology generator. This reads OpenBMP
messages from a Kafka instance and translates those messages into a high level
network topology. It provides the base graph that Collector Services can build
upon.

## Architecture Diagram

!["Framework Architecture"](../docs/Framework.png "Framework Architecture")

## Services

#### Collector Services
Collectors modify the database by providing `Scores` to edges in the graph. This
will let one discover "shortest paths" through the graph, based on that score.
The french press service was a simple ping based latency calculator. The scores
put on edges should a value where lower numbers are preferred over larger values.


Collectors must register with the framework and heartbeat with the framework. This
lets Responders know what `Scores` are available to query against. If a responder
does not heartbeat, the framework will consider it "Down" and it will not return
labels based on that `Score`.

#### Responder Services
Responders implement business logic around returning SR label stacks. At minimum
a responder must receive a source and a destination and it can return a label stack.
It is possible to write a corresponding responder for every collector, or one
generic responder for all collectors known and active in the framework. The responders
should be optimized and built for future use cases.


Sample OpenBMP generated data is available [Here](./openbmp_parsed_data.txt).
This data is generated from
[this virtual topology](https://wwwin-github.cisco.com/raw/paulduda/voltron-network0/master/doc/voltron-network0.png)
in the BXB lab.

# Graph Data Model
The nodes and edges below describe the French Press data model. More nodes and
edges are in development. These edges/nodes will be added by Collectors, not by the
topology engine. Also make note that Scores do not appear in the data model. Collectors
will add them, but we will not be changing the data model each time a new Score is computed.
Those fields will be added at the top level of the object dynamically by the
framework.

## Nodes
```
Router {
  Key String
  Name String
  BGPID String
  ASN String
  RouterIP String (like an id, BMP uses the loopback addr or highest IP)
  IsLocal Bool
}
```

```
Prefix {
  Key String
  Prefix String
  Length Int
}
```

## Edges Between Routers
```
LinkEdgeV4 {
  Ket String
  Label String
  From (RouterKey) String
  To (RouterKey) String
  FromIP String
  ToIP String
  Netmask String
  V6 Bool
}
```

## Edges Between Router & Prefix
```
PrefixEdge {
  Key String
  From (Prefix/Router Key)
  To (Prefix/Router Key) String
  InterfaceIP String
  Labels []String
  NextHop String
  ASPath []string
  BGPPolicy String
}
```

##### See the Developer Guide at [framework/DEVELOPER.md](DEVELOPER.md)
