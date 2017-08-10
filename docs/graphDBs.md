# Graph Databases

Pros and Cons of graph DB offerings. All the Graph Databases leverage a "graph query language" that most have not used before, so we need to account for some learning curve if anyone wants to work on or use these technologies.

## [Neo4j](https://neo4j.com/)

#### Pros
- Established
- Easy to pick up
- Many learning resources available
- Containerized

#### Cons
- Licensing GPLv3 or AGPL
- Supposedly they do not perform as well as many competitors
- Clustering is not supported in the community version
- Golang client is community developed (not officially supported by the neo4j team)


## [OrientDB](http://orientdb.com/)

#### Pros
- Multi-model, NoSQL DB, Key Value DB, and Graph DB all in one place
- Was able to deploy a multi-container cluster with ease (data duplication was on across the DBs)
- Containerized

#### Cons
- Reviews claim it has some performance issues when it gets to scale
- SQL-like query language felt clumsy and forced
- Outdated goclient, must use the python client

## [Cayley](https://github.com/cayleygraph/cayley)

#### Pros
- Separate GraphDB layer and storage backend (portability)
- Containerized / Cloud native

#### Cons
- Random Googler side project
- An "alpha" project. I could not find evidence of it being used in production.
- Memory management issues

## [Arangodb](https://www.arangodb.com/)

#### Pros
- Multi-model, NoSQL DB, Key Value DB, and Graph DB all in one place
- A reasonable SQL-like query language called AQL
- Containerized / cloud native
- [Pluggable](https://www.arangodb.com/why-arangodb/foxx/) : a javascript plugin engine that can manage user access and DB access. Never have to give direct access to the DB, it can all be done via this Foxx engine.
- Official go client

#### Cons
- Mixed reviews on performance, apparently slow in some situations.
- Memory management issues (the beta release fixes these issues by [doing something cute](https://www.arangodb.com/2017/05/rocksdb-integration-arangodb-faqs/) with [RocksDB](http://rocksdb.org/)). We should be able to resolve if deploy in a container

## [JanusGraph](http://janusgraph.org/) ~The DB formally known as Titan~

#### Pros
- Graph frontend with a storage backend (portability)
- Backend support for [Cassandra](http://cassandra.apache.org/), [HBase](http://hbase.apache.org/), [Google Big Table](https://cloud.google.com/bigtable/), AWS? (I believe Titan supported DyanmoDB)
- Backed by the linux foundation

#### Cons
- Complicated to get working (separate parts to deploy), storage backend, optional indexing backend
- Not Containerized
- Need to deploy a Gremlin Server to front JanusGraph, for non-Java clients

## [DGraph](https://dgraph.io/)

#### Pros
- Containerized
- Built to be cloud native / HA / robust
- Golang client
- Decent learning materials (quick to get started)
- Programmy query language
```
{
  find_michael(func: eq(name, "Michael")) {
    _uid_
    name
    age
  }
}
```
- Returns everything in JSON

#### Cons
- UI lacking (crashes on a double click to a node)
- Alpha project (https support just added in March)
  > We recommend using it for internal projects at companies. If you plan to use Dgraph for user-facing production environment, come talk to us. -[Source](https://github.com/dgraph-io/dgraph)
- Release ~monthly. Updates require an export and import
- Possibly [complicated HA deployment](https://docs.dgraph.io/deploy/#multiple-instances), have to manage predicate distribution


## Recommendation

After a few hours with each I'd go with Arangodb. It seems fairly well established, it was easy to pick up its query language and deploy. It offers a lot of resources for learning. In a perfect world I'd want neo4j but its licensing is limiting. My second recommendation would be a tie between JanusGraph and DGraph. They both offer very different benefits, JanusGraph being more portable and established, but seems more complex to install since it requires a storage backend and potentially an indexing backend. DGraph looks appealing as its built to be cloud native, easy to deploy, and has good resources to pick it up. The problem is it is a rapidly changing product, that isn't fully featured. "deployment complexity" vs "hitting the ground running" is at the core of Janus vs DGraph.
