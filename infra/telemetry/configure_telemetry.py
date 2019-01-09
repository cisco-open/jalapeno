#!/usr/bin/python
"""Configure MDT on IOS-XR host devices."""
import logging, os
from netmiko import ConnectHandler
from util import get_hosts

def load_telemetry_config(filename):
    """Load the config specified in the infra folder."""
    config_lines = None
    with open(filename, 'r') as config_fd:
        config_lines = config_fd.readlines()
    return config_lines

def remove_telemetry_config(netmiko_host_dict):
    """Remove telemetry and gRPC configuration.
    This prevents unexpected gRPC config, and telemetry paths.
    """
    xr_config_set = [
        'no telemetry model-driven',
        'no grpc'
    ]
    xe_config_set = [
        'no telemetry ietf subscription 0',
        'no netconf-yang'
    ]
    with ConnectHandler(**netmiko_host_dict) as connection:
        if netmiko_host_dict['device_type'] == 'cisco_xr':
            connection.send_config_set(xr_config_set)
            connection.commit()
        elif netmiko_host_dict['device_type'] == 'cisco_ios':
            connection.send_config_set(xe_config_set)

def apply_telemetry_config(netmiko_host_dict, telemetry_config):
    """Apply the telemetry configuration.
    This function relies on telemetry_config being correct.
    TODO Refactor to validate config intent somehow.
    """
    with ConnectHandler(**netmiko_host_dict) as connection:
        if netmiko_host_dict['device_type'] == 'cisco_xr':
            connection.send_config_set(telemetry_config)
            connection.commit()
        elif netmiko_host_dict['device_type'] == 'cisco_ios':
            connection.send_config_set(telemetry_config)

def main():
    """Load the hosts and configure telemetry."""
    logging.basicConfig(level=logging.INFO)
    xr_telemetry_config = load_telemetry_config(os.path.realpath(os.path.join(os.getcwd(), os.path.dirname(__file__))) + '/config_xr')
    xe_telemetry_config = load_telemetry_config(os.path.realpath(os.path.join(os.getcwd(), os.path.dirname(__file__))) + '/config_xe')
    hosts = get_hosts(os.path.realpath(os.path.join(os.getcwd(), os.path.dirname(__file__))) + '/hosts.json')
    for host in hosts:
        network_host = host['netmiko_network']
        device_type = network_host['device_type']
        logging.info('Configuring telemetry for %s', network_host['ip'])
        remove_telemetry_config(network_host)
        if device_type == 'cisco_xr':
            apply_telemetry_config(network_host, xr_telemetry_config)
        elif device_type == 'cisco_ios':
            apply_telemetry_config(network_host, xe_telemetry_config)

if __name__ == '__main__':
    main()
