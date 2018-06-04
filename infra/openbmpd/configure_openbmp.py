#!/usr/bin/python
"""Configure OpenBMP on IOS-XR host devices."""
import logging, os
from netmiko import ConnectHandler
from util import get_hosts

def load_openbmp_config(filename):
    """Load the config specified in the infra folder."""
    config_lines = None
    with open(filename, 'r') as config_fd:
        config_lines = config_fd.readlines()
    return config_lines

def apply_openbmp_config(netmiko_host_dict, openbmp_config):
    """Apply the openbmp configuration.
    This function relies on openbmp_config being correct.
    TODO Refactor to validate config intent somehow.
    """
    with ConnectHandler(**netmiko_host_dict) as connection:
        connection.send_config_set(openbmp_config)
        connection.commit()

def main():
    """Load the hosts and configure openbmp."""
    logging.basicConfig(level=logging.INFO)
    openbmp_config = load_openbmp_config(os.path.realpath(os.path.join(os.getcwd(), os.path.dirname(__file__))) + '/openbmp_config_xr')
    hosts = get_hosts(os.path.realpath(os.path.join(os.getcwd(), os.path.dirname(__file__))) + '/hosts.json')
    for host in hosts:
        network_host = host['netmiko_network']
        logging.info('Configuring openbmp for %s', network_host['ip'])
        apply_openbmp_config(network_host, openbmp_config)

if __name__ == '__main__':
    main()
