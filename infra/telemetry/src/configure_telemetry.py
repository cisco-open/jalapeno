"""Configure MDT on IOS-XR host devices."""
import logging
from netmiko import ConnectHandler
from util import get_hosts

def load_telemetry_config(filename='../../infra/telemetry/config_xr'):
    """Load the config specified in the infra folder."""
    config_lines = None
    with open(filename, 'r') as config_fd:
        config_lines = config_fd.readlines()
    return config_lines

def remove_telemetry_config(netmiko_host_dict):
    """Remove telemetry and gRPC configuration.
    This prevents unexpected gRPC config, and telemetry paths.
    """
    config_set = [
        'no telemetry model-driven',
        'no grpc'
    ]
    with ConnectHandler(**netmiko_host_dict) as connection:
        connection.send_config_set(config_set)
        connection.commit()

def apply_telemetry_config(netmiko_host_dict, telemetry_config):
    """Apply the telemetry configuration.
    This function relies on telemetry_config being correct.
    TODO Refactor to validate config intent somehow.
    """
    with ConnectHandler(**netmiko_host_dict) as connection:
        connection.send_config_set(telemetry_config)
        connection.commit()

def main():
    """Load the hosts and configure telemetry."""
    logging.basicConfig(level=logging.INFO)
    telemetry_config = load_telemetry_config()
    hosts = get_hosts()
    for host in hosts:
        network_host = host['netmiko_network']
        logging.info('Configuring telemetry for %s', network_host['ip'])
        remove_telemetry_config(network_host)
        apply_telemetry_config(network_host, telemetry_config)

if __name__ == '__main__':
    main()
