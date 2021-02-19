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

Jalapeno's installation script does not automatically add Grafana dashboards.
To do so:
```
Connect to Grafana: Http://<jalapeno_ip>:30300
User/pw = root/jalapeno
After authenticating click on add data source, choose InfluxDB
Enter InfluxDB parameters:
URL: http://influxdb:8086
Database: mdt_db
Basic Auth: root/jalapeno
Http method GET
Click 'save and test'
```
Once added, hover over the Dashboard icon (4 squares on the left), click 'Manage'

Click import on the right side of the screen

Import telemetry json files and modify as necessary to fit your topology

Sample json to import:

https://github.com/jalapeno/jalapeno-lab/blob/master/grafana/egress-mdt.json
https://github.com/jalapeno/jalapeno-lab/blob/master/grafana/ingress-mdt.json


