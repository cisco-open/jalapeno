#! /usr/bin/env python
"""MPLS and MPLSPacket Manager classes
Construct MPLS packets and send them over a specific interface.
"""

from scapy.all import *

class MPLS(Packet):
    name = "MPLS"
    fields_desc =  [
        BitField("label", 3, 20),
        BitField("experimental_bits", 0, 3),
        BitField("bottom_of_label_stack", 1, 1),
        ByteField("TTL", 255)
        ]
    def guess_payload_class(self, payload):
        if len(payload) >= 1:
            if not self.bottom_of_label_stack:
               return MPLS
            ip_version = (orb(payload[0]) >> 4) & 0xF
            if ip_version == 4:
                return IP
            elif ip_version == 6:
                return IPv6
        return Padding

class MPLSPacketManager(object):
    def __init__(self):
        self.p0 = ICMP()

    def create_packet(self, src_MAC, dst_MAC, label_stack, src_IP, dst_IP):
        bind_layers(Ether, MPLS, type = 0x8847)
        bind_layers(MPLS, MPLS, bottom_of_label_stack = 0)
        bind_layers(MPLS, IP)
        pkts = MPLS(label = label_stack[0], bottom_of_label_stack=0)
        for i in range(1,len(label_stack)-1):
            pkt = MPLS(label = label_stack[i], bottom_of_label_stack=0)
            pkts = pkts / pkt
        pkts = pkts / MPLS(label = label_stack[len(label_stack)-1], TTL = 255, bottom_of_label_stack=1)
        self.p0 = Ether(src = src_MAC, dst = dst_MAC) / pkts / IP(src = src_IP, dst = dst_IP) / ICMP()

    def send_ICMP_packet(self):
        srp(self.p0, iface="ens4", retry=0, timeout=1.5)

