from MPLS_packet_generator import MPLSPacketGenerator
from network_sniff import NetworkSniff
from multiprocessing import Process
import os

arp_info = os.popen('ip route | grep default').read()
print("ARP Info: " + arp_info)
arp_info = arp_info.split()
next_hop = arp_info[2]
print("Next hop: " + next_hop)
outgoing_interface = arp_info[-1]
print("Outgoing interface: " + outgoing_interface)

src_MAC = os.popen('ifconfig ' + outgoing_interface  + ' | grep HWaddr').read()
src_MAC = src_MAC.split()[-1]
print("SRC_MAC of " + outgoing_interface + " is: " + src_MAC)

dst_MAC = os.popen('arp -a | grep '+ next_hop  + ' | grep ' + outgoing_interface).read()
dst_MAC = dst_MAC.split()[3]
print("DST_MAC is " + dst_MAC)

packet_obj = MPLSPacketGenerator()
packet_obj.create_packet(src_MAC, dst_MAC, label_stack=[24004, 24006, 24005], src_IP='10.1.2.1', dst_IP='10.11.0.1')

latency_obj = NetworkSniff()

p1 = Process(target=latency_obj.calculate_latency)
p1.start()

p1.join(timeout=1)

p2 = Process(target=packet_obj.send_ICMP_packet)
p2.start()

p2.join(timeout=5)
p1.join(timeout=5)
