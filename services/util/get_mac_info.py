#! /usr/bin/env python
"""This script generates mac numbers for the current device and the next hop.
This means, for now, this script must be on the source device(in this case, v0-vm0).
The next hop is v0-t0. These calculated mac numbers will be used to force traffic
to v0-t0 when creating MPLS packets for the latency calculation.
"""
import os

def get_mac_data():
    """Generate source and destination mac information."""
    arp_info = os.popen('ip route | grep default').read()
    arp_info = arp_info.split()
    next_hop = arp_info[2]
    outgoing_interface = arp_info[-1]
    src_MAC = os.popen('ifconfig ' + outgoing_interface  + ' | grep HWaddr').read()
    src_MAC = src_MAC.split()[-1]
    dst_MAC = os.popen('arp -a | grep '+ next_hop  + ' | grep ' + outgoing_interface).read()
    dst_MAC = dst_MAC.split()[3]
    return {"src_MAC": src_MAC, "dst_MAC": dst_MAC}

if __name__ == '__main__':
    get_mac_data()
    exit(0)