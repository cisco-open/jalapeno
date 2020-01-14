#!/usr/bin/python
"""Configure MDT on IOS-XR / IOS-XE host devices."""
import logging, os, sys, napalm
from util import get_hosts
from jinja2 import *
from configs import pipeline_config

def retrieve_device_info(host_ip, device_type, password):
    logging.info('Retrieving hostname/loopback for %s', host_ip)
    driver = napalm.get_network_driver(device_type)
    device = driver(hostname=host_ip, username='cisco', password=password)
    device.open()
    interface_dict = device.get_interfaces_ip()
    hostname = device.get_facts()['hostname']
    loopback = list(interface_dict["Loopback0"]["ipv4"].keys())[0]
    device.close()
    return hostname, loopback

def render_telemetry_config(device_type, hostname, loopback, pipeline_ip, pipeline_port, template):
    logging.info('Rendering %s telemetry config template for %s', device_type, hostname)
    context = {
        'loopback_ip': loopback,
        'pipeline_ip': pipeline_ip,
        'pipeline_port': pipeline_port,
    }
    outputText = template.render(context)
    dirname = os.path.dirname(os.path.abspath(__file__))
    telemetry_config = os.path.join(dirname, 'configs', device_type, hostname)
    with open(telemetry_config, "w") as file_handler:
        file_handler.write(outputText)

def deploy_telemetry_config(device_type, hostname, host_ip, password):
    logging.info('Deploying %s telemetry config for %s at %s', device_type, hostname, host_ip)
    driver = napalm.get_network_driver(device_type)
    device = driver(hostname=host_ip, username='cisco', password=password)
    device.open()
    dirname = os.path.dirname(os.path.abspath(__file__))
    config_file = dirname + '/configs/' + device_type + '/' + hostname
    device.load_merge_candidate(filename=config_file)
    device.commit_config()
    device.close()

def main():
    """Load the hosts and configure telemetry."""
    logging.basicConfig(level=logging.INFO)
    pipeline_ip, pipeline_port = pipeline_config.pipeline_ip, pipeline_config.pipeline_port
    templateLoader = FileSystemLoader(searchpath=os.path.dirname(os.path.abspath(__file__)))
    templateEnv = Environment(loader=templateLoader)
    xe_template_file, xr_template_file = 'config_xe_template', 'config_xr_template'
    xe_template, xr_template = templateEnv.get_template(xe_template_file), templateEnv.get_template(xr_template_file)

    hosts = get_hosts(os.path.realpath(os.path.join(os.getcwd(), os.path.dirname(__file__))) + '/hosts.json')
    for host in hosts:
        network_host = host['netmiko_network']
        host_ip, device_type, password = network_host['ip'], network_host['device_type'], network_host['password']
        logging.info('Configuring telemetry for %s', host_ip)
        hostname, loopback = retrieve_device_info(host_ip, device_type, password)
        if device_type == 'ios':
            render_telemetry_config(device_type, hostname, loopback, pipeline_ip, pipeline_port, xe_template)
        elif device_type == 'iosxr':
            render_telemetry_config(device_type, hostname, loopback, pipeline_ip, pipeline_port,  xr_template)
        deploy_telemetry_config(device_type, hostname, host_ip, password)
        logging.info('Configured telemetry for %s!', host_ip)
        print()

if __name__ == '__main__':
    main()
