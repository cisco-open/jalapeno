#!/usr/bin/python3.6
"""Remove Telemetry streaming on host network devices.
Unconfigures devices' currently telemetry configurations.
"""
import logging, os, napalm
from util import get_hosts

def remove_telemetry_config(device_type, host_ip):
    driver = napalm.get_network_driver(device_type)
    device = driver(hostname=host_ip, username='cisco', password='cisco')
    device.open()
    if device_type == 'ios':
        device.load_merge_candidate(config='no netconf-yang\n no telemetry ietf subscription 0')
    elif device_type == 'iosxr':
        device.load_merge_candidate(config='no grpc\n no telemetry model-driven')
    device.commit_config()
    device.close()

def main():
    print('Unconfiguring telemetry on ALL devices in hosts.json')
    confirmation = input("Would you like to continue? [yN]: ")
    if confirmation != 'y':
        print('Aborting.')
        exit()
    hosts = get_hosts(os.path.dirname(os.path.abspath(__file__)) + '/hosts.json')
    for host in hosts:
        network_host = host['netmiko_network']
        host_ip, device_type = network_host['ip'], network_host['device_type']
        print('Removing telemetry configuration for', host_ip)
        remove_telemetry_config(device_type, host_ip)
        print('Unconfigured telemetry on', host_ip)
        print()

if __name__ == '__main__':
    main()
