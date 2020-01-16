# Jalapeno
### Database-driven, cloud-native SDN

#### High level architecture 
![jalapeno_architecture](https://wwwin-github.cisco.com/spa-ie/jalapeno/blob/master/docs/jalapeno_architecture.png "jalapeno architecture")

#### SDN is a database problem
With the statement "SDN is database problem" we are saying all SDN use cases can be executed via database mappings and their associated encapsulations. With this framework in mind, Jalapeno has the theoretical ability to address any kind of virtual topology use case. Therefore, Jalapeno is a generalized SDN platform, which may be used for:

* Internal Traffic Engineering - engineered tunnels traversing a network under common management (BGP-LS use cases - see note below**)
* Egress Peer Engineering - engineered tunnels sending traffic out a specific egress router/interface to an external network
* VPN overlays - engineered tunnels creating point-to-point or multipoint overlay virtual networks
* Network Slicing - see VPN overlays
* Service Chaining - engineered tunnels, potentially a series of them, linked together via or seamlessly traversing midpoint service nodes 

#### Some project principles
* Give applications the ability to choose their service/SLA (path through the network)
* The Host is the control/encapsulation point (linux, VPP, other)
* Cloud-native microservice architecture from day 1
* Combine network and application performance data
* Emphasize the use of APIs over Protocols - greater agility

#### Jalapeno's key components

Jalapeno is comprised of a series of microservices which can be summarized as:

* Collectors - capture network topology and performance data and feed the data to Kafka.  Eventually we wish to incorporate application and server/host performance data as well.   

* Data Handlers - parse data coming off Kafka and populate virtual topology data collections in the Arango graph database.

* Databases - an Influx time-series database, and an Arango graph database.

* Services (SR-Apps).  are mini-applications that receive user requests for service (TE/QoE, VPN, etc.), and then mine the graph database for the label stack or SRH needed to execute the service request.  Each SR-App's capabilities are exposed via Jalapeno's API.  

Jalapeno's kubernetes/microservice architecture make it inherently extensible, and we imagine the number of Collectors, Services (SR-Apps), and graphDB virtual topology use cases to expand significantly as our community grows.

Jalapeno's initial POC example Apps are "Latency" and "Bandwidth": a user or application may call Jalapeno's API-GW requesting lowest-latency-path to destination X, or least-utilized (most BW available) to destination Y. The API-GW passes the request to the Latency or Bandwidth service which in turn mine the database and respond with the appropriate SR label stack or SRH.  
 

#### ** Note on BGP-LS

The key to developing and supporting virtual topology use cases is the programmatic acquisition of topology data.  Traditional service provider SDN-TE platforms focus on Internal-TE and therefore leverage BGP-LS. With Jalapeno we wish to eventually support all the above categories of use case, and therefore we use BGP Monitoring Protocol (BMP) and leverage the OpenBMP.snas.io collector. BMP provides a superset of topology data, including:

* BGP-LS topology data - which hopefully includes service-chain data in the near future: https://www.ietf.org/id/draft-dawra-idr-bgp-ls-sr-service-segments-03.txt
* iBGP and eBGP IPv4, IPv6, and labeled unicast topology data
* BGP VPNv6, VPNv6, and EVPN topology data

## Getting Started
To get started, visit the [Quick Start](Quick-Start.md) guide.




