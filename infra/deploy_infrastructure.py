#!/usr/bin/python
"""Deploy Infrastructure including Kafka, Arango, Influx, Grafana, OpenBMP, Pipeline, and Telemetry."""
import kafka.deploy_kafka as KafkaDeployer
import arangodb.deploy_arango as ArangoDeployer
import influxdb.deploy_influx as InfluxDeployer
import grafana.deploy_grafana as GrafanaDeployer
import openbmpd.deploy_openbmp as OpenbmpDeployer
import pipeline.deploy_pipeline as PipelineDeployer
# import telemetry.deploy_telemetry as TelemetryDeployer

def main():
    """Deploy each component of Voltron's Infrastructure."""
    ### Infrastructure Deployment
    KafkaDeployer.main()
    ArangoDeployer.main()
    InfluxDeployer.main()
    GrafanaDeployer.main()
    OpenbmpDeployer.main()
    PipelineDeployer.main()
    #TelemetryDeployer.main()

if __name__ == '__main__':
    main()
