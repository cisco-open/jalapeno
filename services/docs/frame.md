# Voltron Framework

This framework defines what is a vService (Voltron Service) and describes the infrastructure that supports it.
The purpose of a vService is to provide SR labels to clients and populate the database for the Voltron system. vServices
are pluggable components of the Voltron system that will enable flexibility and customization.

vServices must be implemented as containers that are Kubernetes friendly. Responder services will be exposed
through [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/). vServices must provide
a [liveness and readiness endpoint](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)
so they can be properly managed by Kubernetes, they must also provide a
[prometheus](https://prometheus.io/docs/operating/configuration/#scrape_config)
[metrics endpoint](https://prometheus.io/docs/instrumenting/writing_exporters/).
This will enable vServices information to be available in Prometheus.

There are two classes of vServices Collector vServices, and Responder vServices.

## vService Registration

The framework team will expose an endpoint that allows services to register themselves. Collector and Responder
services will register through separate endpoints. Collector vServices tell the framework what
edge field is being updated, its update frequency and its current health. Responder vServices will be able
to see what fields are available for pathing and they can do any custom business logic on top of that.
The framework will expose what services of each type have registered, what fields are provided by Collectors
and what is provided by Responders. The data just listed will be provided by the vService upon registration.

#### Fields
```
- name
- description
- health
- update interval
```

## Collector vServices (CvService)
Collectors job is to update the database. It takes in information, most likely from Kafka, processes that information and
populates the data model in ArangoDB. An example of a CvService a Latency CvService, this CvService would
ingest information related to latency and define a weight on edges in the Graph. Its sole purpose is to set weights
and compute those values. Another Collector could be Utilization it could be taking information from a different source
but at the end of the day it is modifying edges in the database. Collectors are internal, they do not respond to to
clients looking for Segment Routing information, they provide information that informs a Responder vService.
When a vService registers it provides the edge field name that it is updating, the fieldType it will add that is
numeric and the edgeType its modifying it could be both prefix edge and link edges, or one of them.

#### Collector Only Fields:
```
- edgeType
- fieldName
- fieldType (This must be numeric to properly compute a shortest path)
```

## Responder vServices (RvService)
Responders implement business logic for routing based on the topology and Collector provided fields. Responders
ask the framework about what Collected Fields are available. A RvService can utilize one / many of the Collected Fields,
this means one front end can exist to leverage both Latency and Utilization. Responders provide an API
to Apps / clients / hosts that are looking to make network requests, they provide a label stack to guide packets through the network.

#### Responder Only Fields
```
- Ingress endpoint path or Service IP
```

TLDR: Ingress is basically a K8s Service that fronts multiple containers. `http://<Ingress>/RvS1`
will route traffic to the `latency RvService` and `http://<Ingress>/RvS2` will route traffic to the
`utilization RvService`

#### POST /label
```
Input:
- Start
- Destination
- Constraint (optional)

Output:
- Label Stack
```
*other custom fields could be added here, but this would be the minimum*

An example of a responder service is a Path Responder. A client would POST to the responder API with the inputs listed.
The response will return the shortest path based on what constraint was provided. A RvService could handle, one or
many different constraints such as latency, utilization, security.


## Why not Foxx services
[Foxx](https://www.arangodb.com/why-arangodb/foxx/) services are a JS framework provided by ArangoDB that
simplify DB access. We want to avoid being tied to this framework to make vService development easier.
Each vService should be self sufficient in talking to ArangoDB, since they provide clients
for a variety of programming languages. We should not depend on this Foxx abstraction because it may require each vService to
also develop a Foxx Service. I'm also uncertain how we scale and manage Foxx services. I would rather let Kubernetes do
the work for us in managing / scaling / life cycling these vServices.
