# Grafana

Grafana is Jalapeno's visual dashboard and metric-visualization tool.

Jalapeno's data-pipeline populates the time-series database [InfluxDB](../influxdb) with metrics regarding the state of the network. 
Jalapeno's instance of Grafana loads InfluxDB as its data-source. 

Grafana can have any number of graphical representations of the health of the network, including historical bandwidth usage, historical latency metrics, and more. 

## Deploying Grafana
Grafana is deployed using `kubectl`, as seen in the [deploy_infrastructure script](../deploy_infrastructure.sh) script. The configurations for Grafana's deployment are in the various YAML files in the [jalapeno/infra/grafana/](.) directory.  

## Accessing Grafana
To access Grafana's UI, log in at `<server_ip>:30300`, using credentials `root/jalapeno`.

## Creating Grafana Dashboards for Jalapeno Telemetry Data

Dashboards are not automatically loaded. Follow these steps to get some dashboards up and running:

1.	Connect to Grafana:  Http://<jalapeno_ip>:30300
2.	User/pw = root/jalapeno 
3.	Add InfluxDB as a data source

URL: http://influxdb:8086
Database: mdt_db
Basic Auth: root/jalapeno
Http method GET
Save and Test

4.	Hover over Dashboard icon (4 squares on the left)

