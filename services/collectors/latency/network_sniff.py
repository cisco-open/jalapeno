#! /usr/bin/env python
"""This class has the core functions to calculate and upsert latency.
Using the sniffer tool from scapy, request and reply time are recorded
while sending a packet using the network_latency script. The latency
is then updated in the Paths collection.
"""

from scapy.all import *
from pyArango.connection import *
from configs import arangoconfig, queryconfig
import logging, decimal
from util import connections

class NetworkSniff(object):
    def send_latency(self, key, latency):
        """Update latency in Paths collection for the given document specified by key."""
        connection = connections.ArangoConn()
        database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
        collection = database['Paths']
        aql = """FOR p in Paths
            FILTER p._key == @key
            UPDATE p with { Latency : @latency } in Paths """
        bindVars = {'key': key,'latency': latency}
        database.AQLQuery(aql, rawResults=True, bindVars=bindVars)

    def calculate_latency(self, key):
        """Calculate the latency for a specific path using the sniffer tool."""
        pkts = []
        pkts = sniff(iface='ens4', filter='icmp or mpls', timeout=3)
        print("We have pkts " + str(pkts))
        request_time = pkts[0].time
        # print("We got request time " + str(request_time))
        reply_time = pkts[1].time
        # print("We got reply time " + str(reply_time))
        latency = decimal.Decimal(reply_time - request_time)
        latency = round(latency, 2)
        print ('Latency:', latency)
        print ("###############################################################################")
        latency = float(latency)
        self.send_latency(key, latency)
