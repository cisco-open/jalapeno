# Jalapeno Collectors

Jalapeno Collectors are responsible for collecting topology and performance data from the network. Any Jalapeno infrastructure component that brings data into the Jalapeno cluster from the network is considered a Collector.

## OpenBMP Collector
For topology-related data, Jalapeno currently uses OpenBMP. Devices are configured to send OpenBMP data to the cluster via the OpenBMP Collector, which is deployed as a StatefulSet in Kubernetes. Inside the `openbmp` directory, the `openbmp_ss.yaml` defines the pod to be built using the latest OpenBMP image, and then loads its configuration from `openbmp_cfg.yaml` file. 

Thus, Jalapeno has an OpenBMP-collector pod running in the cluster that serves as the ingress point for all topology data from the devices. Once the data arrives from the network to the OpenBMP pod, the data is then forwarded to Kafka for the next level of data processing(see [processors](../processors)). 

You can learn more about OpenBMP and the way OpenBMP organizes topology data [here](https://www.snas.io/docs/).

## Telegraf-Ingress Collector

## Jalapeno Collector Utilities
### Telemetry Configuration on Devices

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
To facillitate configuring OpenBMP on multiple nodes at the same time, Jalapeno comes with `openbmp-util`. Fill out the `hosts.json` file and run the `configure_openbmp.py` script. This script uses netmiko to authenticate and deploy the configuration described in `openbmp_config_xr` to each device.
