# MDT Stack in Kubernetes

[Grafana](https://grafana.com/), [InfluxDB](https://www.influxdata.com/time-series-platform/influxdb/), and [Pipeline](https://github.com/cisco/bigmuddy-network-telemetry-pipeline/) all packaged in to one [Kubernetes](https://kubernetes.io/) pod for [MDT](https://www.cisco.com/c/en/us/solutions/service-provider/cloud-scale-networking-solutions/model-driven-telemetry.html) (Telemetry) visualization and exploration.

Grafana serves as the visualization, InfluxDB as the TSDB (Time Series Database), and Pipeline as the MDT consumption/transformation utility. Only Grafana will be exposed on port 30300 on the pod host.

* ```mdt-stack.yaml``` specifies the Kubernetes configuration.
* ```/configmaps/``` contains subfolders which correspond to [configmaps](https://kubernetes.io/docs/tasks/configure-pod-container/configmap/), the folders contain the config files.

Verified to work on OpenStack.

## [Grafana](https://wwwin-github.cisco.com/spa-ie/voltron-redux/tree/mdt-stack/infra/mdt-stack/configmaps/mdt-grafana-config)

* Credentials: ```voltron/voltron```
* No other special configuration.

## InfluxDB

* User Credentials: ```voltron/voltron```
* Admin Credentials: ```admin/gsplab```
* No other special configuration. Lots of room for optimization here.
* Not recommended as a container under high volume, production workloads.

## [Pipeline](https://wwwin-github.cisco.com/spa-ie/voltron-redux/tree/mdt-stack/infra/mdt-stack/configmaps/mdt-pipeline-config)

* Assumes consumption from Kafka.
    * JSON formatted data.
    * ```voltron.telemetry``` topic.
* Assumes export to InfluxDB as ```voltron``` user.
* Filter data based upon ```metrics.json```
* **pipeline_rsa should not be used in production!**

### Update Docker image

```bash
git clone https://github.com/cisco-ie/bigmuddy-network-telemetry-pipeline
cd docker
docker build -f ./Kuberfile -t pipeline:<version> .
docker login -u gspie-deployer
# Development
docker tag pipeline:<version> dockerhub.cisco.com/gspie-dev-docker/pipeline:<version>
docker push dockerhub.cisco.com/gspie-dev-docker/pipeline:<version>
# Production
# docker tag pipeline:<version> dockerhub.cisco.com/gspie-docker/pipeline:<version>
# docker push dockerhub.cisco.com/gspie-docker/pipeline:<version>
```
