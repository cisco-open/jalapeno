from scapy.all import *
import decimal

class NetworkSniff(object):
    def calculate_latency(self):
        print('Start sniffing')
        pkts = sniff(iface="ens4", filter="icmp or mpls", timeout=10)
        print('End sniffing')
        print(pkts)
        request_time = pkts[0].time
        reply_time = pkts[1].time
        latency = decimal.Decimal(reply_time-request_time)
        latency = round(latency, 2)
        print(latency)
