#!/usr/bin/python
"""Deploy Infrastructure including Kafka,
Arango, Influx, OpenBMP, Telemetry, MDT.
"""
import kafka.deploy_kafka as KafkaDeployer
import arangodb.deploy_arango as ArangoDeployer
import influxdb.deploy_influx as InfluxDeployer
import openbmpd.deploy_openbmp as OpenbmpDeployer
import telemetry.deploy_telemetry as TelemetryDeployer
import mdt-stack.deploy_mdt as MDTDeployer

def main():
    """Deploy each component of Voltron's Infrastructure."""
    ### OC Login

    ### Infrastructure Deployment
    KafkaDeployer.main()
    ArangoDeployer.main()
    InfluxDeployer.main()
    OpenbmpDeployer.main()
    TelemetryDeployer.main()
    MDTDeployer.main()

if __name__ == '__main__':
    main()
