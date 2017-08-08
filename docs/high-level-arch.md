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

### Requirements

 - Top priority is making it easy to stand up a Voltron instance to start creating services and solving use cases in minimal time.  If any troubleshooting or debugging time is spent in this phase of service creation, we have failed.  Documentation of "how to interact with Voltron" will be another critical requirement, but that's for another section.  Bottom line, if it's too hard to solve problems with Voltron, nobody will use and we will have squandered a huge opportunity.
 - If we're using kubernetes, we need to have at minimum at k8s installer that will deploy its [base components and requirements](https://kubernetes.io/docs/concepts/overview/components/) on a system or set of systems.  For simplicity, we should start with an all-in-one deployment model (controller and node on same system) until it becomes clear we need to scale beyond.
 - The base installer should run as a standalone script or Ansible playbook....or a combination thereof (there's always the bootstrapping problem of getting the tools loaded you need to run the tools).  Lots of such examples exist that could be leveraged.
 - To facilitate running on AWS, we should build cloud images that can easily be deployed.  In the process, we're also defining a more generic VM image that will be capable of running for standalone deployments and especially automated testing. (*TODO: investigate corporate AWS account*)
 - We assume that a Voltron Service will comprise several containers working together to achieve a use case, so a service should provide its own set of yaml files associated with running it on a k8s cluster.  However, we should not require that a service creator understands the intricacies of k8s, hence we should provide simpler tools or templates for generation of k8s templates.
 - A Voltron deployment needs to know which services to load; these ultimately translate into k8s templates, but they imply some variety of manifest to define what should load.  The interaction with the services framework is implied, as well, although most of the framework duties are already covered by the kubectl API and don't need to be reinvented.
 - We will start with fixed revisions of depdendent components (k8s, graph db, kafka, etc.) to keep sands from shifting beneath us, although our build system should take into account the potential for upgrade as we rev Voltron in the future.
 - We should not make assumptions about public connectivity to download dependent packages, implying we should stand up our own Docker repository at a Voltron deployment that is pre-loaded with all the needed images.
 - For the foreseeable future, we should assume updates are a wipe and replace (see HA section below).

### Voltron Platform Packaging & Deployment

The Voltron Platform and Infrastructure dependencies are packaged as a single tarball - `voltron-0.0.0.tgz` - where `0.0.0` is replaced by a legitimate version number. The package is a self-contained artifact. It contains all the images and automation scripts reqiured for fully deploying the Voltron Platform to an existing cluster of one or more SSH-able Linux servers. The contents of the tarball are structured as an Ansible project with an Ansible role per component. Each component must include its version number as part of the role name. Components include:

1. Kubernetes
2. Docker Registry
3. Kafka
4. Graph Database
5. Data Processing layers
6. Built-in Services that Voltron may provide by default
7. Logging & Monitoring components
8. User Interface components

A top-level Ansible inventory file and playbook tie together the entire deployment. After editing the inventory with environment-specific information, Voltron can be deployed with a single Ansible command. A separate playbook could be provided for upgrade. 

A secondary layer of packaging can be used to further simplify deployment for specific environments. For example, an ISO or kernelrd/initrd can be provided for bare metal installation. A VM image (such as qcow2, ova, vmdk) can be provide for deploying to a public cloud or local virtualization environment. Note that the secondary packaging could be an image of a completely pre-installed Voltron system (i.e. a snapshot of a system on which the ansible playbook has already been run) or simply an image that includes the tarball itself and any tools (i.e. ansible, python) pre-installed. In the latter case, the secondary package could also be a docker image. With such an image available on docker hub, an entire Voltron system could be deployed through a single docker command such as `docker run cisco/voltron <list of machines to install>`.

### Voltron Services Packaging & Deployment

In Voltron's initial phase, packaging and deployment of Voltron services will match the Voltron platform. Each Service is pacakged as a self-contained Ansible role - including all images and automation required for that Service. The Ansible role can be rolled into the larger `voltron-0.0.0.tgz` tarball or stand alone (e.g. `myservice-0.0.0.tgz`). In this phase, Services will be deployed just like any other component of Voltron.

In the future, Voltron Services should be treated more like plugins and may even be presented in a catalog-like format such that an operator can pick and choose which Services to deploy. In this model, a Service might also define dependencies on other Services that may be required. Achieving a catalog-like user experience may be realized either through the same Ansible role tarball as described above or through another form of packaging. For example, Sigma could be used to represent Service dependencies and each Voltron Service could be packaged as a Sigma Plugin container.

## Configuration

*TODO: describe important points of configuration*

*Author: Jeff*

Areas of configuration:
 * For build time:
    - Services to include in deployer bundle _*this feels like we either need an enhancement to the work stream to build a "custom installer" (which is really just some k8s API commands) to install services on an already-deployed instance or do have a "custom builder" that creates manifests for auto-deployment at system startup time*_
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


## User Interface / Visibility

*Author: Rachael & Jeff*

The specifics of a Voltron user interface are highly dependent on its function:
  * Monitoring
      - health of platform
      - currently loaded services and health
      - performance statistics

    [Heapster](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-usage-monitoring/) is an endorsed dashboard for k8s that will cover its basics.

    For the Voltron platform, a dashboard should contain the following basic monitoring elements:
      - List of installed services and health status (taken from health checks specified in k8s yaml)
      - Kafka statistics over time (# producers/consumers, # and list of topics, resource utilization)
      - DB statistics over time (resource utilization, # entries, total queries, queries/sec, etc.)
      - Front-end statistics over time (resource utilization, total queries, queries/sec, etc.)
  * Configuration
      - adjusting platform parameters

    With the use of etcd as a redundant keystore, it should be possible to tweak basic parameters of containers deployed via k8s without disrupting Voltron services.  Providing a means of tweaking configuration knobs declared in the Voltron services should be straightforward by exposing the options on a generic UI container.  This process assumes that the service has a means for updating its config on the fly rather than reloading.

  * Diagnosis/Troubleshooting
      - viewing event log
      - performance statistics

    At minimum, an event log to both retrieve the persistent log and to show events in real-time will aid troubleshooting.  While building a harness for tracing services is probably overkill, intelligent use of the Kafka bus to both publish and subscribe to service events should make this function relatively simple.

Furthermore, an individual service may wish to have other visualizations to accompany it, or the service may only provide visualization as its core function.  It seems desirable (if not obvious) to have any pure visualization service follow the same template as any other service on Voltron.


## Logging & Monitoring

*TODO: describe requirements and mechanism for monitoring voltron*
*Author: Omar & Rachael*

_side comment from byzek that may be obvious: the logging service should be constructed just like any other solves-a-use-case service.  It pulls data from the platform (in this case, probably off of a Kakfa topic), performs persistence/journaling operations, and publishes an API for outside users to access._

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

Developers are responsible for writing and testing their code with unit tests.
Regularly scheduled testing of release candidate code includes functional tests, security tests, performance tests, endurance tests, scale tests, and stress tests.


Components for a test infrastructure:

- **Test Scheduling** - Jenkins
- **Build Chains / Test Chains** - Jenkins Pipeline
- **Source and Test Artifacts** - Github or Nexus
- **Test Provenance** - a relational database (mySQL / Postgres)
- **Test Lifecycle Monitoring** - TBD.  Necessary for Endurance / Scale / and Stress tests.  Voltron's monitoring system should be employed here, with added intellegence about the phase of the test.
- **Virtual Testbeds** - TBD.  Dynamic testbed creation is critical both for development and for longer term testing.  Options include VIRL, Paul Duda's version of VIRL, Mandelbrot OpenNebula Solution.  Bruce has a static testbed taht has been used for previous Voltron testing.
- **Physical Testbeds** - TBD.  Options include RTP lab & SJ Building 9 lab.  Issue: Segment Routing is not enabled in the SJ lab.
- **Workload Generator** - TBD.  SJ lab has Ixia.  Need to check into RTP lab work generator.
- **Dashboard / Visibility into Test status** - Start with Jenkins.  Alternate views will be necessary (performance views / historical views ...).  This ties in with Voltron Dashboard needs.

## Development Environment

*TODO: describe the development environment, CI pipeline, and any developer tooling*

*Author: Matt & Dan*

Voltron uses the following developer tooling:

- **Source Control** - Github (either wwwin-github or github.com)
- **Code Review** - Github Pull Request
- **Defect Tracking** - Github Issues
- **Task Management** - TBD - Rally?
- **Language** - Go (for data processing) and Ansible (for automation)
- **Continuous Integration** - TBD - Jenkins?
- **Packaging** - Docker
- **Test Framework** - TBD - go test?

As all components of Voltron are containerized, a complete local dev environment should be possible without virtualization. However, given the use of Ansible for automation and Kubernetes for cluster management, a VM environment is recommended. To facilitate development, a Vagrantfile is included to launch a suitable virtual machine and install both Voltron and its dependencies. 

Enabling local development may also require the creation of a telemetry simulator (which can generate fake telemetry and forward it to a Kafka queue) as well as one or more mock Services.

## Documentation

*TODO: describe what tools will be used for generating documentation*

*Author: Matt & Dan*

All documentation will be written in [markdown](https://guides.github.com/features/mastering-markdown/). Documentation will be version controlled alongside Voltron source code (i.e. in the same git repository). A feature is not considered done until both the code and documentation are updated (ideally in the same merge commit). The following markdown documents will be maintained over the lifetime of Voltron in `docs/` directory:

1. architecture.md - describes the components of Voltron and how they fit together (this document)
2. deployment.md - describes how to configure, install, and upgrade Voltron
3. operations.md - describes how to use and monitor Voltron on a day-to-day basis
4. api.md - describes the API (should be auto-generated via Swagger)
5. hacking.md - describes how to contribute to Voltron source code

Each Voltron Service may be maintained in a separate repository with its own Service-specific documentation. In general, each Service should provide the same five documents listed here or a meaningful equivalent.

Documentation can be published as static HTML using Jekyll if desired and either hosted using Github Pages or a separate server. Documentation can also be converted into PDF or other formats using tools such as pandoc.

## Infrastructure

*Author: Erez & Omar*

## Open Questions

1. What does telemetry contain? Is it documented?
  * Update, 8/7: Bruce/Drew/Remington helping BXB team access live stream and documenttion
2. What is the lifecycle of a service from installation to removal?
