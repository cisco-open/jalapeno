# Voltron High-level Software Architecture

## Overview

Voltron provides a framework for:
1. collecting and processing network telemetry from a heterogeneous population of network devices
2. storing a model of the network topology and updating that model based on processed telemetry
3. running services that can query the model and expose an API to network endpoints

At its core Voltron is comprised of two scalable data structures:
1. a message queue for processing network telemetry
2. a graph database for storing the network topology and information about that topology

The message queue is fed by network devices (or forwarders that are actively collecting metrics on behalf of network devices). Tools such as Pipeline and Telegraf are used to forward metrics to the message queue. Metrics should be forwarded based on a configurable interval.

The message queue itself must provide an asynchronous interface and must easily scale to handle many thousands of messages per second. For this purpose, Kafka is chosen as the initial message queue for implementing Voltron.

Between the message queue and the graph database lies a data processing layer. The complexity of this layer very much depends on the Voltron use cases and sophistication of analysis required for achieving those use cases. Initially, a simple custom tool will be written that subscribes to the message queue, parses messages, handles rudimentary analysis, and updates the network model in the graph database. This component must be stateless and must be horizontally scalable. As more types of network devices and more types of telemetry are handled by Voltron, a more extensible framework for processing that data may be necessary. Tools such as Apache Spark Streaming or Apache Storm may provide useful frameworks for consuming telemetry from Kafka, processing it through multiple stages, and updating Neo4j. In the absence of such a framework, simple data processing extensibility can also be achieved through arbitrary programs which consume messages from Kafka and update Neo4j based on those messages.

The graph database itself must offer a graph query interface and must be capable of handling a high rate of updates. For this purpose, Neo4j is chosen as the initial graph database for implementing Voltron. The network topology can be inferred from the network device telemetry and will be built dynamically in Neo4j by the data processing layer.

The services must be flexible and Voltron must offer a framework for running any type of service. Services must have access to the Neo4j graph database in order to make intelligent decisions based on their domain-specific algorithms. Each service may offer its own domain-specific API. For this purpose, Kubernetes is chosen as an application lifecycle manager and clustering tool. Containers and Kubernetes provide a useful abstraction for creating custom services with direct access to a graph database API.

!["Voltron Software Architecture"](voltron-sw-arch.png "Voltron Software Architecture")

## Project Phasing

### Phase 1 - Core Functionality

The core functionality reflects the above description of Voltron. The only point of extensibility is the framework for running services. In the initial implementation, all components are deployed on a single Kubernetes cluster.

### Phase 2 - Service Framework

As the Voltron team gains sufficient experience with writing these services and working with the Neo4j API, patterns may emerge that could be institutionalized within a common framework wrapping the Neo4j API.

Experience with writing Voltron services might also reveal that services should keep a local cache of state, rather than relying on Neo4j for each query. Further, such a local cache may benefit from real-time updates via a message queue.

### Phase 3 - Service Composability

While the number of initial services should be small, a fully-fledged Voltron might expose a large catalog of services with strict interdependencies. Each new service might build on the API and capabilities of existing services. Such dependencies could be defined as part of the service catalog. Voltron could explicitly enforce dependency rules before deploying a service. A service dependency tool such as [Sigma](sigma-dev.github.io/sigma/) could enable this concept of composability.

### Future Phases

In the future, more sophistication can be added to Voltron. A few examples are described in this section with limited detail.

#### Cloud Offload

In addition to running Voltron itself in a cloud environment, individual pieces of Voltron could be offloaded to a public cloud. For example, Kafka could be used to vector messages off to both a local Voltron instance as well as a cloud-hosted Voltron. In a cloud hosted environment, more big data tools may be available for telemetry analysis, the result of which could be fed back to an on-premises Voltron instance.

#### Multi-site & Federation

A multi-site Voltron deployment might map well to a division of responsibilities in the network. For example, an operator might deploy one Voltron instance for their WAN and one for each data center network. An endpoint in one data center wishing to connect with an endpoint in another data center may need to contact multiple Voltron instances to determine a low latency path between them. Some form of Voltron federation could simplify such a query.

#### Time-series Data

In addition to relying on the instantaneous state of the network, Voltron could also store and expose historical information. However, Neo4j may not be the ideal tool for this purpose and the architecture should be reassessed if time-series data becomes a requirement.

## Packaging & Deploy Automation

*TODO: describe how Voltron infra, framework, and services will be packaged, deployed, versioned, and upgraded*
*Author: Matt & Jeff*

Preliminary thoughts:
 - Top priority is making it easy to stand up a Voltron instance to start creating services and solving use cases in minimal time.  If any troubleshooting or debugging time is spent in this phase of service creation, we have failed.  Documentation of "how to interact with Voltron" will be another critical requirement, but that's for another section.  Bottom line, if it's too hard to solve problems with Voltron, nobody will use and we will have squandered a huge opportunity.
 - If we're using kubernetes, we need to have at minimum at k8s installer that will deploy its [base components and requirements](https://kubernetes.io/docs/concepts/overview/components/) on a system or set of systems.  For simplicity, we should start with an all-in-one deployment model (controller and node on same system) until it becomes clear we need to scale beyond.
 - The base installer should run as a standalone script or Ansible playbook....or a combination thereof (there's always the bootstrapping problem of getting the tools loaded you need to run the tools).  Lots of such examples exist that could be leveraged.
 - To facilitate running on AWS, we should build cloud images that can easily be deployed.  In the process, we're also defining a more generic VM image that will be capable of running for standalone deployments and especially automated testing. (*TODO: investigate corporate AWS account*)
 - We assume that a Voltron Service will comprise several containers working together to achieve a use case, so a service should provide its own set of yaml files associated with running it on a k8s cluster.  However, we should not require that a service creator understands the intricacies of k8s, hence we should provide simpler tools or templates for generation of k8s templates.
 - A Voltron deployment needs to know which services to load; these ultimately translate into k8s templates, but they imply some variety of manifest to define what should load.  The interaction with the services framework is implied, as well, although most of the framework duties are already covered by the kubectl API and don't need to be reinvented.
 - We will start with fixed revisions of depdendent components (k8s, graph db, kafka, etc.) to keep sands from shifting beneath us, although our build system should take into account the potential for upgrade as we rev Voltron in the future.
 - We should not make assumptions about public connectivity to download dependent packages, implying we should stand up our own Docker repository at a Voltron deployment that is pre-loaded with all the needed images.
 - For the foreseeable future, we should assume updates are a wipe and replace (see HA section below).

## Configuration

*TODO: describe important points of configuration*
*Author: Jeff*

Areas of configuration:
 * For build time:
    - Services to include in deployer bundle *this feels like we either need an enhancement to the work stream to build a "custom installer" (which is really just some k8s API commands) to install services on an already-deployed instance or do have a "custom builder" that creates manifests for auto-deployment at system startup time*
    - Target deployment environment (bare metal vs. VM vs. AWS)
 * For runtime:
    - Footprint of Voltron services (how many minions, how many cores, how much RAM, etc.)
    - Volumes location (for the container holding the graph DB) (*needed?*)
    - Debug vs. "normal" mode
 * Individual service configuration:
    - Where to find services registry (for locating frontend, etcd, Kafka, graph DB, etc.)
    - Warmup parameters
    - Topology subset to solve specific use case
    - Min time between graph DB queries
    - Algorithm knobs (something domain-specific to the problem being solved

Non-k8s-specific service configuration should be provided through yaml files that are read upon service startup.

### Global Platform Configuration


## User Interface / Visibility

*TODO: describe the end-user interface*
*Author: Rachael & Jeff*

## Logging & Monitoring

*TODO: describe requirements and mechanism for monitoring voltron*
*Author: Omar & Rachael*

## High Availability

*Author: Jeff*

### Global Platform Considerations
Perhaps some amazing technological advancement will occur and a day will come when Voltron is capable of orchestrating every individual flow through every part of the network. Since such a day is incredibly unlikely, we assert that Voltron will be used primarily for steering traffic that has somehow been deemed "special". We are making unconventional decisions about how to forward certain flows that would otherwise be sent along the same path as everything else.  As this backup forwarding mechanism will persist even if Voltron suffers an outage, high availability requirements for Voltron seem substantially lower priority than other engineering problems.

This is not to say that we will not have multiple instances of certain parts of Voltron's infrastructure.  However, these instances will serve primarily to increase the performance of a Voltron system and not the availability.  When Voltron's services start maintaining state to address use cases in network flow management, we will determine an appropriate path for repliating that state and ensuring access to it remains uninterrupted.  Starting with Kubernetes's built-in [High Availability mechanisms](https://kubernetes.io/docs/admin/high-availability) would be a prudent first step.

### Individual Services
Individual services may define their own schemes for achieving high availability through lower-volume, simpler cache replication schemes, or simply by maintaining multiple replicas of a responder or querier.  In cases where an algorithm performs more complex computations or creates state in the warmup phase, the service may enlist redis or memcached for storing computations or complex query results.  If it becomes evident that multiple services find consistent use of such distributed stores useful, the Votron platform should provide a global instance instead of allowing each service to spin up its own.


## Test Considerations

*TODO: describe the test strategy for voltron*
*Author: Rachael*

## Development Environment

*TODO: describe the development environment, CI pipeline, and any developer tooling*
*Author: Matt & Dan*

## Documentation

*TODO: describe what tools will be used for generating documentation*
*Author: Matt & Dan*

## Infrastructure

*Author: Erez & Omar*

## Open Questions

1. What does telemetry contain? Is it documented?
