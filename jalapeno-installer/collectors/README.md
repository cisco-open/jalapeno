# Jalapeno Collectors

Jalapeno Collectors are responsible for collecting topology and performance data from the network. Any Jalapeno infrastructure component that brings data into the Jalapeno cluster from the network is considered a Collector.

## GoBMP Collector
To collect topology-related data, Jalapeno currently uses the golang implementation of OpenBMP, or "GoBMP". Devices are configured to send GoBMP data to the cluster via the GoBMP Collector, which is deployed as a StatefulSet in Kubernetes. Inside the `gobmp` directory, the `gobmp-collector.yaml` file defines the pod to be built using the latest GoBMP image.

Thus, Jalapeno has a GoBMP Collector pod running in the cluster that serves as the ingress point for all topology data from the devices. Once the data arrives from the network to the GoBMP pod, the data is then forwarded to Kafka for the next level of data processing(see [processors](../processors)). 

You can learn more about GoBMP and the way GoBMP organizes topology data [here](https://github.com/sbezverk/gobmp).

## Telegraf-Ingress Collector
To collect performance-related data, Jalapeno uses Telegraf. Devices are configured to send telemetry data to the cluster via the Telegraf-Ingress Collector, which is deployed as a Deployment in Kubernetes. Inside the `telegraf-ingress` directory, the `telegraf_ingress_dp.yaml` file defines the pod to be built using the latest Telegraf image, and then loads its configuration from the `telegraf_ingress_cfg.yaml` file. 

Thus, Jalapeno has an Telegraf-Ingress Collector pod running in the cluster that serves as the ingress point for all performance data from the devices. As defined in the `telegraf_ingress_cfg.yaml` file, the data is then outputted to Kafka into the `jalapeno.telemetry` topic, which is queries by Jalapeno's performance processors(see [processors](../processors)). 

Note: This is called Telegraf-Ingress specifically because there is a Telegraf-Egress processor that forwards the data from Kafka to InfluxDB further down Jalapeno's data pipeline. 

