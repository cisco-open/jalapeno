#!/usr/bin/python
"""Configure GoBMP on IOS-XR / IOS-XE host devices."""
import logging, os, sys, napalm
from util import get_hosts
from jinja2 import *

def deploy_gobmp_config(device_type, host_ip):
    logging.info('Deploying %s telemetry config at %s', device_type, host_ip)
    driver = napalm.get_network_driver(device_type)
    device = driver(hostname=host_ip, username='cisco', password='cisco')
    device.open()
    dirname = os.path.dirname(os.path.abspath(__file__))
    if device_type == 'ios':
        config_file = dirname + '/gobmp_config_xe'
    elif device_type == 'iosxr':
        config_file = dirname + '/gobmp_config_xr'
    device.load_merge_candidate(filename=config_file)
    device.commit_config()
    device.close()

def main():
    """Load the hosts and configure telemetry."""
    logging.basicConfig(level=logging.INFO)
    hosts = get_hosts(os.path.realpath(os.path.join(os.getcwd(), os.path.dirname(__file__))) + '/hosts.json')
    for host in hosts:
        network_host = host['netmiko_network']
        host_ip, device_type = network_host['ip'], network_host['device_type']
        logging.info('Configuring GoBMP on %s', host_ip)
        deploy_openbmp_config(device_type, host_ip)
        logging.info('Configured GoBMP on %s!', host_ip)
        print()

if __name__ == '__main__':
    main()
