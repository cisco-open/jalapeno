# Grafana

Grafana is Jalapeno's visual dashboard and metric-visualization tool.

Jalapeno's data-pipeline populates the time-series database [InfluxDB](../about//infrastructure.md#influxdb) with metrics regarding the state of the network.
Jalapeno's instance of Grafana loads InfluxDB as its data-source.

Grafana can have any number of graphical representations of the health of the network, including historical bandwidth usage, historical latency metrics, and more.

## Deploying Grafana

Grafana is deployed using `kubectl`, as seen in the `/install/infra/deploy_infrastructure.sh` script. The configurations for Grafana's deployment are in the various YAML files in the `/infra/grafana/` directory.  

## Accessing Grafana

To access Grafana's UI, log in at `<server_ip>:30300`, using credentials `root/jalapeno`.

## Creating Grafana Dashboards for Jalapeno Telemetry Data

Jalapeno's installation script does not automatically add Grafana dashboards.

To do so:

1. Connect to Grafana: `http://<jalapeno_ip>:30300`
    - User/pw = root/jalapeno
2. After authenticating, click on **Add data source** & choose InfluxDB
3. Enter InfluxDB parameters:
    - URL: `http://influxdb:8086`
    - Database: `mdt_db`
    - Basic Auth: `root/jalapeno`
    - HTTP method `GET`
4. Click **Save and Test**
5. Once added, hover over the Dashboard icon and click **Manage**
6. Click **Import** on the right side of the screen
7. Import telemetry json files and modify as necessary to fit your topology

A few sample Grafana dashboards can be found here:

- [Egress Traffic](https://github.com/cisco-open/jalapeno/blob/main/install/infra/grafana/egress-mdt.json)
- [Ingress Traffic](https://github.com/cisco-open/jalapeno/blob/main/install/infra/grafana/ingress-mdt.json)
