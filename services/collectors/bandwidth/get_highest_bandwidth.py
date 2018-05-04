#! /usr/bin/env python
"""This script gets the highest available bandwidth path of a network
given a source and a destination. The bandwidth is calculated as a rolling
average bandwidth. The path is described using IP addresses and a
segment routing label stack.
"""

from pyArango.connection import *
from configs import arangoconfig, queryconfig
from util import connections
import logging, argparse, sys

def main(source, upstream_source, destination):
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    collection = database['LinkEdgesV4']
    upstream_source = "Routers/" + upstream_source
    destination = "Prefixes/" + destination
    aql = """FOR v,e IN OUTBOUND SHORTEST_PATH @upstream_source TO @destination
             GRAPH "topology"
             OPTIONS {weightAttribute: 'Bandwidth', defaultWeight: 200000000000000}
             FILTER e.Label != null
             RETURN e.Label"""
    bindVars = {'upstream_source': upstream_source, 'destination': destination}
    highest_available_bandwidth = database.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    print highest_available_bandwidth

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Get the highest available bandwidth path from a source to destination.')
    parser.add_argument('source', help="Source IP (10.1.2.1)")
    parser.add_argument('upstream_source', help="Upstream Source IP (10.1.1.0)")
    parser.add_argument('destination', help="Destination IP (10.11.0.0_24)")
    args = parser.parse_args()

    if(args.source != queryconfig.vmsource):
        print("Error: Invalid source IP")
	exit(0)
    if(args.upstream_source != queryconfig.upstream_source):
        print("Error: Invalid upstream source IP")
	exit(0)
    destinations = [dest.rstrip('\n\r') for dest in open('configs/prefixes.txt')]
    if(args.destination not in destinations):
        print("Error: Invalid destination IP")
 	exit(0)

    main(args.source, args.upstream_source, args.destination)
    exit(0)
