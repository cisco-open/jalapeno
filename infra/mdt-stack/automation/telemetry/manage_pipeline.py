"""Manage Pipeline on host network devices.
Assumes that Pipeline will run in a "guestshell" which
is configured in the hosts.json config file.
"""
import os
import logging
import argparse
from contextlib import closing
from netmiko import ConnectHandler, SCPConn
from util import get_hosts

def provision_pipeline(netmiko_linux_dict, base_path='../../infra/telemetry/pipeline/'):
    """Provision Pipeline files to guestshell via SCP.
    Does not start Pipeline.
    """
    logging.info('Provisioning Pipeline on %s.', netmiko_linux_dict['ip'])
    filenames = ['pipeline', 'pipeline_rsa', 'pipeline.conf']
    file_paths = [os.path.join(base_path, filename) for filename in filenames]
    with ConnectHandler(**netmiko_linux_dict) as connection:
        with closing(SCPConn(connection)) as scp_connection:
            for file_path in file_paths:
                dest_path = os.path.basename(file_path)
                logging.debug('Provisioning %s', dest_path)
                scp_connection.scp_transfer_file(
                    file_path,
                    dest_path
                )

def remove_pipeline(netmiko_linux_dict):
    """Remove all Pipeline files from guestshell."""
    logging.info('Removing Pipeline on %s.', netmiko_linux_dict['ip'])
    remove_pipeline_cmd = 'rm pipeline*'
    with ConnectHandler(**netmiko_linux_dict) as connection:
        connection.send_command(remove_pipeline_cmd)

def start_pipeline(netmiko_linux_dict):
    """Start Pipeline and log its PID to pipeline.pid."""
    logging.info('Starting Pipeline on %s.', netmiko_linux_dict['ip'])
    start_pipeline_cmd = './pipeline \
        -log=pipeline.log -config=pipeline.conf -pem=pipeline_rsa \
        &> pipeline_console.log & echo $! > pipeline.pid'
    with ConnectHandler(**netmiko_linux_dict) as connection:
        connection.send_command(start_pipeline_cmd)

def stop_pipeline(netmiko_linux_dict):
    """Stop Pipeline based upon PID file."""
    logging.info('Stopping Pipeline on %s.', netmiko_linux_dict['ip'])
    stop_pipeline_cmd = 'kill $(tail -n 1 pipeline.pid) && rm pipeline.pid'
    with ConnectHandler(**netmiko_linux_dict) as connection:
        connection.send_command(stop_pipeline_cmd)

def restart_pipeline(netmiko_linux_dict):
    """Stop and Start Pipeline."""
    logging.info('Restarting Pipeline on %s.', netmiko_linux_dict['ip'])
    stop_pipeline(netmiko_linux_dict)
    start_pipeline(netmiko_linux_dict)

def reset_pipeline(netmiko_linux_dict):
    """Remove and Provision Pipeline."""
    logging.info('Resetting Pipeline on %s.', netmiko_linux_dict['ip'])
    remove_pipeline(netmiko_linux_dict)
    provision_pipeline(netmiko_linux_dict)

def setup_args():
    """Formats the arguments expected to run the manager.
    Allows for certain commands and specification of specific hosts.
    """
    parser = argparse.ArgumentParser(
        description="Provision Pipeline to network devices."
    )
    parser.add_argument('action',
        help='(provision | remove | reset | start | stop | restart)'
    )
    parser.add_argument('--hostnames',
        nargs='+',
        help='hostnames to operate against',
    )
    return parser.parse_args()

def main():
    """Load the hosts and manage Pipeline."""
    logging.basicConfig(level=logging.INFO)
    args = setup_args()
    action_map = {
        'provision': provision_pipeline,
        'remove': remove_pipeline,
        'reset': reset_pipeline,
        'start': start_pipeline,
        'stop': stop_pipeline,
        'restart': restart_pipeline
    }
    if args.action not in action_map.keys():
        logging.error('Action must be in action_map!')
        exit(1)
    hosts = get_hosts()
    guestshell_hosts = set([host['netmiko_linux']['ip'] for host in hosts])
    if args.hostnames:
        if not set(args.hostnames).issubset(guestshell_hosts):
            logging.error('Some hostnames do not exist!')
            exit(2)
        else:
            hosts = [host for host in hosts if host['netmiko_linux']['ip'] in args.hostnames]
    for host in hosts:
        action_map[args.action](host['netmiko_linux'])

if __name__ == '__main__':
    main()
