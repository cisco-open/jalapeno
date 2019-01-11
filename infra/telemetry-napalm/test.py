# Sample script to demonstrate loading a config for a device.
#
# Note: this script is as simple as possible: it assumes that you have
# followed the lab setup in the quickstart tutorial, and so hardcodes
# the device IP and password.  You should also have the
# 'new_good.conf' configuration saved to disk.
from __future__ import print_function

import napalm
import sys
import os


def main():
    # Use the appropriate network driver to connect to the device:
    driver = napalm.get_network_driver('ios')

    # Connect:
    device = driver(hostname='10.0.0.40', username='cisco',
                    password='cisco')

    print('Opening ...')
    device.open()
    interface_dict = device.get_interfaces_ip()
    loopback = list(interface_dict["Loopback0"]["ipv4"].keys())[0]
    print(loopback)
   
    # close the session with the device.
    device.close()
    print('Done.')


if __name__ == '__main__':
    main()
