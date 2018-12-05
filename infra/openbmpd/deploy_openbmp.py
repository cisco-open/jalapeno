"""Deploy the openbmp_collector container locally"""
import os

os.system("docker run -d --name=openbmp_collector -e KAFKA_FQDN=10.0.250.2:30902 -v /var/openbmp/config:/config -p 5000:5000 openbmp/collector")