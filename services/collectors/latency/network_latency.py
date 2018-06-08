#! /usr/bin/env python
"""This script calculates the latency for each path in the path set as
calculated and returned by the label_generator script. For each path,
a MPLSPacketManager creates the packet and sends it, while the NetworkSniffer
uses scapy's sniffer tool to monitor request and reply times.
"""

import os, decimal, time
from scapy.all import *
from multiprocessing import Process

from util import get_mac_info
from network_sniff import NetworkSniff
from MPLS_packet_manager import MPLSPacketManager
import label_generator

def calculate_latency(path):
    """Calculate the latency for a given path. Format the MPLS packet
    to be constructed. Begin listening as you send out the MPLS packet.
    """
    mac_info = get_mac_info.get_mac_data()
    src_MAC = mac_info["src_MAC"]
    dst_MAC = mac_info["dst_MAC"]
    key=path["Key"]
    source=path["Source"]
    destination=path["Destination"] # ie 10.11.0.0_24
    print("Calculating latency for " + key)
    # question here :: why does the MPLS packet generation need to destination to be 10.11.0.1 instead of 10.11.0.0_24? bug?
    if(destination == "10.0.254.0_24"):
        packet_destination = '10.0.254.1'
    else:
        packet_destination = 0

    labels = ' '.join(path["Label_Path"].split('_'))  # label stack formatting
    labels = [int(x) for x in labels.split(' ')]

    packet_manager = MPLSPacketManager()
    packet_manager.create_packet(src_MAC, dst_MAC, label_stack=labels, src_IP=source, dst_IP=packet_destination)

    latency_obj = NetworkSniff()
    # start listening and calculation process
    p1 = Process(target=latency_obj.calculate_latency, args=(key,))
    p1.start()
    p1.join(timeout=1)

    # send packet
    p2 = Process(target=packet_manager.send_ICMP_packet)
    p2.start()
    p2.join(timeout=2)

def main():
    while True:
        all_label_stacks = label_generator.generate_labels()
        #for label_stack_set in all_label_stacks:
        for path in all_label_stacks[0]:
            if '_' in path["Label_Path"]:
                calculate_latency(path)
            else:
                print("Single label path -- NodeSID Labels not currently incorporated. Skipping this path.")
                print ("###############################################################################")
            time.sleep(0.5)

    #all_label_stacks = label_generator.generate_labels()
    #for label_stack_set in all_label_stacks:
    #    for path in label_stack_set:
    #	    print path
    #        calculate_latency(path)
    #        time.sleep(0.5)


if __name__ == '__main__':
    main()
    exit(0)
