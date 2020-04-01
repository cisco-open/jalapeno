# Grafana

Grafana is Jalapeno's visual dashboard and metric-visualization tool.

Jalapeno's data-pipeline populates the time-series database [InfluxDB](../influxdb) with metrics regarding the state of the network. 
Jalapeno's instance of Grafana loads InfluxDB as its data-source. 

Grafana can have any number of graphical representations of the health of the network, including historical bandwidth usage, historical latency metrics, and more. 

## Deploying Grafana
Grafana is deployed using `kubectl`, as seen in the [deploy_infrastructure script](../deploy_infrastructure.sh) script. The configurations for Grafana's deployment are in the various YAML files in the [jalapeno/infra/grafana/](.) directory.  

## Accessing Grafana
To access Grafana's UI, log in at `<server_ip>:30300`, using credentials `root/jalapeno`.

