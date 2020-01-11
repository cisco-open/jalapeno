# Voltron
### Database-driven, cloud-native SDN

#### SDN is a database problem
With the statement "SDN is database problem" we are saying all SDN use cases can be executed via database mappings and their associated encapsulations.  With this framework in mind, Voltron has the theoretical ability to address any kind of virtual topology use case.  Therefore, Voltron is a generalized SDN platform, which may be used for:

* Internal Traffic Engineering - engineered tunnels traversing a network under common management
* Egress Peer Engineering - engineered tunnels sending traffic out a specific egress router/interface to an external network
* VPN overlays - engineered tunnels creating point-to-point or multipoint overlay virtual networks
* Network Slicing - see VPN overlays
* Service Chaining - engineered tunnels, potentially a series of them, linked together via or seamlessly traversing midpoint service nodes 

#### Some project principles
* Give applications the ability to choose their service/SLA (path thru the network)
* The Host is the control/encapsulation point (linux, VPP, other)
* Cloud-native microservice architecture from day 1
* Combine network and application performance data
* Emphasize the use of APIs over Protocols - greater agility

#### High level architecture 
![voltron_architecture](https://wwwin-github.cisco.com/spa-ie/voltron/blob/brmcdoug/docs/voltron_architecture.png "voltron architecture")

#### Voltron's key components

Voltron is comprised of a series of microservices which can be summarized as:

* Collectors - capture network topology and performance data and feed the data to Kafka.  Eventually we wish to incorporate application and server/host performance data as well.  

* Data Handlers - parse data coming off Kafka and populated virtual topology data collections in the Arango graph database

* Databases - an Influx time-series database, and Arango graph database

* Services (SR-Apps).  are mini-applications that receive user requests for service (TE/QoE, VPN, etc.), and then mine the graph database for the label stack or SRH needed to execute the service request.  Each SR-App's capabilities are exposed via Voltron's API.  

Voltron's kubernetes/microservice architecture make it inherently extensible, and we imagine the number of Collectors, Services (SR-Apps), and graphDB virtual topology use cases to expand significantly as our community grows.

Voltron's initial POC example Apps are "Latency" and "Bandwidth": a user or application may call Voltron's API-GW requesting lowest-latency-path to destination X,  or least-utilized (most BW available) to destination Y.  The API-GW passes the request to the Latency or Bandwidth service which in turn mine the database and respond with the appropriate SR label stack or SRH.  
 



