#!/usr/bin/python3.6
"""Deploy OpenBMP container locally.
Receives OpenBMP data and forwards to Kafka."""
#import configure_openbmp
import os

def main():
    ### Assuming OpenBMP is configured on devices
    #print("Configuring OpenBMP on devices")
    #configure_openbmp.main()

    print("Deploying OpenBMP container on host")
    os.system("docker run -d --name=openbmp_collector -e KAFKA_FQDN={{ kafka_endpoint }} -v /var/openbmp/config:/config -p {{ openbmp_port }}:{{ openbmp_port }} openbmp/collector")

if __name__ == '__main__':
    main()
