# Jalapeno Quick-Start Guide
## Requirements

Currently, Jalapeno is deployed in a CentosKVM on a server. This CentosKVM houses:
* The Jalapeno testbed. To set up the Jalapeno testbed, follow the instructions [here](docs/testbed_installation.md).
* The OpenShift cluster -- almost all of Jalapeno's infrastructure and services are deployed as containers in this cluster.

To set up the CentosKVM, follow the instructions [here](docs/centos_vm.md). This also brings up an OpenShift cluster for Jalapeno to use.

#### Server Requirements
* Ubuntu 18.04, minimum 16 vCPU, 96GB memory, 270GB (200GB for Testbed + 80GB for Jalapeno)

## Deploying Jalapeno

Once the OpenShift cluster is up and running in the CentosKVM, Jalapeno can be deployed. 
```
ssh centos@10.0.250.2
git clone https://wwwin-github.cisco.com/spa-ie/jalapeno.git
cd jalapeno
```
If you have your own infrastructure components (i.e. your own ArangoDB instance), you can configure IP addresses etc. by running the `configure_jalapeno.py` script. Otherwise: 
```
./deploy_jalapeno.sh
```

Once Jalapeno is deployed, you should have:
* Infrastructure deployed: Kafka, Zookeeper, Pipeline, ArangoDB, InfluxDB, and Grafana deployed in OpenShift
* Telemetry configured: Any devices listed in [hosts.json](/infra/telemetry/hosts.json) will be streaming data
* Services deployed: Topology and Performance collector services deployed in OpenShift
* API deployed: API ready to receive requests from any Jalapeno clients

## Using Jalapeno

Topology and performance data are now both being collected and married together into the virtual topology collections in ArangoDB. You can access the containers and their UIs from your browser. In this setup, Jalapeno is deployed on the "Bruce-Dev" server (10.200.99.7) so adjust the URL as necessary.

OpenShift UI: https://10.200.99.7:8443/ (username: admin, default password: admin)\
ArangoDB UI: http://10.200.99.7:30852/ (username: root, default password: jalapeno)\
Grafana UI: http://10.200.99.7:30300/ (username: root, default password: jalapeno)

The best way to see Jalapeno in action is to utilize the [Jalapeno Client](https://wwwin-github.cisco.com/spa-ie/jalapeno-client). This client formulates the requests desired (i.e. lowest-latency) and sends the request to Jalapeno's API gateway, which responds with the corresponding label-stack. The client then programs this label stack into the header of packets heading towards their final destination using simple "ip route" commands. 


### Sample API Requests:
Coming soon!


