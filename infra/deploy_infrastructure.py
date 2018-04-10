#!/usr/bin/python
"""Deploy Infrastructure including Kafka,
Arango, Influx, OpenBMP, Telemetry, MDT.
"""
from subprocess import call
import pexpect, sys
import kafka.deploy_kafka as KafkaDeployer
import arangodb.deploy_arango as ArangoDeployer
import influxdb.deploy_influx as InfluxDeployer
import openbmpd.deploy_openbmp as OpenbmpDeployer
import telemetry.deploy_telemetry as TelemetryDeployer
import mdt-stack.deploy_mdt as MDTDeployer



### OpenShift Login
#call(["oc", "login", "https://10.200.99.44:8443"])
oclogin = pexpect.spawn('oc login https://10.200.99.44:8443')
oclogin.delaybeforesend = 1
oclogin.expect('Username:')
oclogin.sendline('admin')
oclogin.expect('Password:')
oclogin.sendline('admin')

### Infrastructure Deployment
KafkaDeployer.main()
ArangoDeployer.main()
InfluxDeployer.main()
OpenbmpDeployer.main()
TelemetryDeployer.main()
MDTDeployer.main()

