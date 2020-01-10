# Voltron
### Database-driven, cloud-native SDN

#### Project principles
* Give applications the ability to choose their service/SLA (path thru the network)
* SDN is a database problem
* APIs not Protocols
* The Host is the control/encapsulation point (linux, VPP, other)
* Microservice architecture from day 1
* Combine network and application performance data

#### High level architecture 
![voltron_architecture](https://wwwin-github.cisco.com/spa-ie/voltron/blob/brmcdoug/docs/voltron_architecture.png "voltron architecture")

#### Voltron's key components

Voltron is comprised of a series of microservices we categorize Collectors and Services (SR-Apps).  Collectors capture and parse topology and performance data and feed the data into Voltron's graph database.  Services/SR-Apps are mini-applications that receive user requests for service (TE/QoE, VPN, etc.), and then mine the graph database for the label stack or SRH needed to execute the service request.  Each SR-App's capabilities are exposed via Voltron's API.  

The POC example Apps are "Latency" and "Bandwidth": a user or application may call Voltron's API requesting lowest-latency-path to destination X,  or least-utilized (most BW available) to destination Y.  The API passes the request to the Latency or Bandwidth service which in turn mine the database and respond with the appropriate SR label stack or SRH.


Voltron is a collection of microservices that harness real-time network telemetry data and technologies such as segment routing. Voltron reveals new ways of maximizing business growth using the network.

With the ability to determine which path in a network is the current "low latency path", "high bandwidth path", etc. Voltron gives applications the power to engineer their traffic, effectively enabling segment routing at the host.  

For Service Providers looking to manage their networks using segment routing, Voltron is an answer to the challenge of traffic engineering and network limitation. 



