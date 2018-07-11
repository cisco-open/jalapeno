# topology/database

This database directory contains the interfaces and structure definitions for each of the ArangoDB collections.

The included collections are:
* Routers - router.go
* Prefixes - prefix.go
* External-Prefixes - internal_prefix.go
* Internal-Prefixes - external_prefix.go
* Internal-Transport-Prefixes - internal_transport_prefix.go
* Link-Edges - link_edge.go
* Prefix-Edges - prefix_edge.go

There are also ArangoDB connection and document-creation helper files: collector.go, database.go, helper.go, object.go

#
### Creating / Modifying Collections
To add a new collection:
- modify arango.go
    - extend the NewArango function, ensuring the collection as either a vertex or edge collection
    - extend the UpsertSafe function, adding handling for your new collection

- create the new collection's go file
    - define the structure of a document in the collection
    - include functions GetKey, SetKey, MakeKey, and GetType
