#!/usr/bin/python3.6
"""Remove Telemetry streaming on host network devices.
Unconfigures devices' currently telemetry configurations.
"""
import logging, os
import configure_telemetry
from util import get_hosts

def main():
    print("Unconfiguring telemetry on devices")
    hosts = get_hosts(os.path.realpath(os.path.join(os.getcwd(), os.path.dirname(__file__))) + '/hosts.json')
    for host in hosts:
        network_host = host['netmiko_network']
        logging.info('Removing telemetry configuration for %s', network_host['ip'])
        configure_telemetry.remove_telemetry_config(network_host)
    exit()

if __name__ == '__main__':
    main()
