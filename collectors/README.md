# Jalapeno Collectors

Jalapeno Collectors are responsible for collecting topology and performance data from the network. Any Jalapeno infrastructure component that brings data into the Jalapeno cluster from the network is considered a Collector.

## OpenBMP Collector
To collect topology-related data, Jalapeno currently uses OpenBMP. Devices are configured to send OpenBMP data to the cluster via the OpenBMP Collector, which is deployed as a StatefulSet in Kubernetes. Inside the `openbmp` directory, the `openbmp_ss.yaml` file defines the pod to be built using the latest OpenBMP image, and then loads its configuration from the `openbmp_cfg.yaml` file. 

Thus, Jalapeno has an OpenBMP Collector pod running in the cluster that serves as the ingress point for all topology data from the devices. Once the data arrives from the network to the OpenBMP pod, the data is then forwarded to Kafka for the next level of data processing(see [processors](../processors)). 

You can learn more about OpenBMP and the way OpenBMP organizes topology data [here](https://www.snas.io/docs/).

## Telegraf-Ingress Collector
To collect performance-related data, Jalapeno uses Telegraf. Devices are configured to send telemetry data to the cluster via the Telegraf-Ingress Collector, which is deployed as a Deployment in Kubernetes. Inside the `telegraf-ingress` directory, the `telegraf_ingress_dp.yaml` file defines the pod to be built using the latest Telegraf image, and then loads its configuration from the `telegraf_ingress_cfg.yaml` file. 

Thus, Jalapeno has an Telegraf-Ingress Collector pod running in the cluster that serves as the ingress point for all performance data from the devices. As defined in the `telegraf_ingress_cfg.yaml` file, the data is then outputted to Kafka into the `jalapeno.telemetry` topic, which is queries by Jalapeno's performance processors(see [processors](../processors)). 

Note: This is called Telegraf-Ingress specifically as there is a Telegraf-Egress processor that forwards the data from Kafka to InfluxDB further down Jalapeno's data pipeline. 

## Jalapeno Collector Utilities
### Telemetry Configuration on Devices
An example configuration of telemetry is shown below:
```
grpc
 port 57400
!
telemetry model-driven
 destination-group jalapeno
  address-family ipv4 <server_ip> port 32400
   encoding self-describing-gpb
   protocol grpc no-tls
  !
 !
 sensor-group cisco_models
  sensor-path Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface
 !
 sensor-group openconfig_interfaces
  sensor-path openconfig-interfaces:interfaces/interface
 !
 subscription base_metrics
  sensor-group-id cisco_models sample-interval 30000
  sensor-group-id openconfig_interfaces sample-interval 30000
  destination-id jalapeno
  source-interface Loopback0
 !
!
```

To facillitate configuring telemetry on multiple nodes at the same time, Jalapeno comes with the `mdt-util` directory. Fill out the `hosts.json` file to specify which nodes the configuration should be deployed on, and then edit the `<server_ip>` field  in the `mdt_config_xr` file to point to where the Jalapeno cluster is running. Finally, run the `configure_telemetry.py` script. This script uses netmiko to authenticate and deploy the configuration to each device.

### OpenBMP Configuration on Devices
An example configuration of OpenBMP is shown below:
```
bmp server 1
 host <server-ip> port 30555
 description jalapeno OpenBMP
 update-source MgmtEth0/RP0/CPU0/0
 flapping-delay 60
 initial-delay 5
 stats-reporting-period 60
 initial-refresh delay 30 spread 2
!
```
To facillitate configuring OpenBMP on multiple nodes at the same time, Jalapeno comes with the `openbmp-util` directory. Fill out the `hosts.json` file to specify which nodes the configuration should be deployed on, and then edit the `<server_ip>` field in the `openbmp_config_xr` file to point to where the Jalapeno cluster is running. Finally, run the `configure_openbmp.py` script. This script uses netmiko to authenticate and deploy the configuration to each device.
