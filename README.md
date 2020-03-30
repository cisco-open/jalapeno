# Jalapeno
### A cloud-native SDN infrastructure platform

To install Jalapeno and get started, visit the [Getting-Started.md](Getting-Started.md) guide.

### High level architecture 
![jalapeno_architecture](https://github.com/cisco-ie/jalapeno/blob/master/docs/diagrams/jalapeno_architecture.png "jalapeno architecture")

#### Platform Overview: SDN is a database problem
With the statement "SDN is database problem" we are saying all SDN use cases can be executed via database mappings and their associated encapsulations. With this framework in mind, Jalapeno has the theoretical ability to address any kind of virtual topology use case. Therefore, Jalapeno is a generalized SDN platform, which may be used for:

* Internal Traffic Engineering - engineered tunnels traversing a network under common management (BGP-LS use cases - see note below**)
* Egress Peer Engineering - engineered tunnels sending traffic out a specific egress router/interface to an external network
* VPN overlays - engineered tunnels creating point-to-point or multipoint overlay virtual networks
* Network Slicing - see VPN overlays
* Service Chaining - engineered tunnels, potentially a series of them, linked together via or seamlessly traversing midpoint service nodes 

#### Some project principles
* Give applications the ability to choose their service/SLA (path through the network)
* The Host may be the control/encapsulation point (linux, VPP, other)
* Cloud-native microservice architecture from day 1
* Combine network and application performance data
* Emphasize the use of APIs over Protocols - greater agility

#### Jalapeno's key components

Jalapeno is comprised of a series of microservices which can be summarized as:

* Collectors - capture network topology and performance data and feed the data to Kafka.  Eventually we wish to incorporate application and server/host performance data as well.  The collection stack also includes Influx TSDB and Grafana for darta visualization

* Data Processors, Graph Database, and Time-Series Database - Jalapeno has two classes of processors: 
  * Base data processors: parse topology and performance data coming off Kafka and populate the Influx TSDB and base data collections in the Arango graph database.  The Topology and Telegraf pods are base processors.
  * Virtual_Topology processors: mine the graph and TSDB data collections and then populate virtual topology collections in the graph DB.  LS, EPE, and L3VPN, are examples of virtual topology processors.

![jalapeno_processors](https://github.com/cisco-ie/jalapeno/blob/master/docs/diagrams/jalapeno_processors.png "jalapeno processors")

* API-GW - expose Jalapeno's virtual topology data for application consumption

* SR-Apps - mini-applications that mine the graph and time-series databases for the label stack or SRH data needed to execute topology or traffic engineering use cases.  Each SR-App should have its own API to field client requests for Segment Routing network services.  

Jalapeno's kubernetes architecture make it inherently extensible, and we imagine the number of Collectors, graphDB virtual topology use cases, and SR-Apps to expand significantly as our community grows.

Jalapeno's initial example Apps are EPE "Latency" and "Bandwidth": a user or application may call the EPE Latency or Bandwidth API requesting lowest-latency-path to External destination X, or least-utilized (most BW available) to destination Y. The App makes a Jalapeno API-GW call for Latency or Bandwidth which in turn mines the EPE virtual topology database and respond with the appropriate SR label stack or SRH.  

#### ** Note on BGP-LS

The key to developing and supporting virtual topology use cases is the programmatic acquisition of topology data.  Traditional service provider SDN-TE platforms focus on Internal-TE and therefore leverage BGP-LS. With Jalapeno we wish to eventually support all the above categories of use case, and therefore we use BGP Monitoring Protocol (BMP) and leverage the OpenBMP.snas.io collector. BMP provides a superset of topology data, including:

* BGP-LS topology data - which hopefully includes service-chain data in the near future: https://www.ietf.org/id/draft-dawra-idr-bgp-ls-sr-service-segments-03.txt
* iBGP and eBGP IPv4, IPv6, and labeled unicast topology data
* BGP VPNv6, VPNv6, and EVPN topology data





