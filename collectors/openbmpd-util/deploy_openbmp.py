#!/usr/bin/python3.6
"""Deploy OpenBMP container locally.
Receives OpenBMP data and forwards to Kafka."""
#import configure_openbmp
import os

def main():
    ### Assuming OpenBMP is configured on devices

    # Skipping configuration of OpenBMP on devices, uncomment if necessary
    # print("Configuring OpenBMP on devices")
    # configure_openbmp.main()

    print("Deploying OpenBMP container on host")
    os.system("docker run -d --name=openbmp_collector -e KAFKA_FQDN=10.0.250.2:30902 -v /var/openbmp/config:/config -p 5000:5000 openbmp/collector")

if __name__ == '__main__':
    main()
