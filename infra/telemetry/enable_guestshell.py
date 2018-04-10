#!/usr/bin/python
"""Enable bash access on IOS-XR devices."""
import logging
import socket
from contextlib import closing
from netmiko import ConnectHandler
from util import get_hosts

def check_port_open(hostname, port):
    """Check if a specified port is open on a host.
    https://stackoverflow.com/a/35370008
    """
    is_open = False
    open_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    with closing(open_socket) as sock:
        if sock.connect_ex((hostname, port)) == 0:
            is_open = True
    return is_open

def check_valid_login(netmiko_linux_dict):
    """Check if login credentials are valid."""
    is_valid = False
    try:
        with ConnectHandler(**netmiko_linux_dict) as connection:
            is_valid = True
    except Exception:
        pass
    return is_valid

def exec_enable_guestshell(netmiko_network_dict, linux_username, linux_password):
    """Execute the bash commands to enable guestshell.
    Adds a non-root user, specifies password,
    and enables SSH/shell.
    """
    with ConnectHandler(**netmiko_network_dict) as connection:
        connection.send_command('bash -c useradd -m {username}'.format(
                username=linux_username
            )
        )
        connection.send_command('bash -c echo {username}:{password} | chpasswd'.format(
                username=linux_username, password=linux_password
            )
        )
        connection.send_command('bash -c chkconfig --add sshd_operns')
        connection.send_command('bash -c service sshd_operns start')

def main():
    """Load the hosts and enable guestshell."""
    logging.basicConfig(level=logging.INFO)
    logging.getLogger('paramiko.transport').setLevel(logging.WARNING)
    hosts = get_hosts()
    for host in hosts:
        network_host = host['netmiko_network']
        guestshell_host = host['netmiko_linux']
        logging.info(
            'Processing network host %s with guestshell %s ...',
            network_host['ip'], guestshell_host['ip']
        )
        if not check_port_open(guestshell_host['ip'], guestshell_host['port']):
            logging.info('Enabling guestshell for %s', network_host['ip'])
            exec_enable_guestshell(
                network_host,
                guestshell_host['username'],
                guestshell_host['password']
            )
        elif check_valid_login(guestshell_host):
            logging.warning(
                'Guestshell appears to already be enabled at %s!',
                guestshell_host['ip']
            )
        else:
            logging.error(
                'Guestshell credentials invalid for %s!',
                guestshell_host['ip']
            )

if __name__ == '__main__':
    main()
