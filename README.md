# Voltron
### A database-driven, cloud-native SDN solution

#### Project principles
* Give applications the ability to choose their service/SLA (path thru the network)
* SDN is a database problem
* APIs not Protocols
* The Host is the control/encapsulation point (linux, VPP, other)
* Microservice architecture from day 1
* Combine network and application performance data

#### High level architecture 
![voltron_architecture](https://wwwin-github.cisco.com/spa-ie/voltron/blob/brmcdoug/docs/voltron_architecture.png "voltron architecture")

Voltron's key components

At the heart of Voltron a set of "collectors" and "services" is a set of microservices that collect and parse network topology and performance data (collectors), and then map that data into a graph database.  With the topology + performance mappings we then add 

Voltron is a collection of microservices that harness real-time network telemetry data and technologies such as segment routing. Voltron reveals new ways of maximizing business growth using the network.

With the ability to determine which path in a network is the current "low latency path", "high bandwidth path", etc. Voltron gives applications the power to engineer their traffic, effectively enabling segment routing at the host.  

For Service Providers looking to manage their networks using segment routing, Voltron is an answer to the challenge of traffic engineering and network limitation. 

To make “source routing” work, Voltron uses APIs over protocols, and takes advantage of new technologies that enable data collection, scalability and more (such as BMP, gRPC, etc.).


