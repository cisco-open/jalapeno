#! /usr/bin/env python
"""This script gets the lowest latency path of a network
given a source and a destination from ArangoDB. The path is
described using IP addresses and a segment routing label stack.
"""

from pyArango.connection import *
from configs import arangoconfig, queryconfig
from util import connections
import logging, argparse, sys

def main(source, destination):
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    collection = database['Paths']
    aql = """FOR p IN Paths
             FILTER p.Source == @source AND p.Destination == @destination
             SORT p.Latency
             LIMIT 1
             RETURN [p.Path, p.Label_Path, p.Latency]"""
    bindVars = {'source': source, 'destination': destination}
    lowest_latency = database.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    print lowest_latency

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Get the lowest latency path from a source to destination.')
    parser.add_argument('source', help="Source IP (10.1.2.1)")
    parser.add_argument('destination', help="Destination IP (10.11.0.0_24)")
    args = parser.parse_args()

    if(args.source != queryconfig.vmsource):
        print("Error: Invalid source IP")
	exit(0)

    destinations = [dest.rstrip('\n\r') for dest in open('configs/prefixes.txt')]
    if(args.destination not in destinations):
        print("Error: Invalid destination IP")
 	exit(0)

    main(args.source, args.destination)
    exit(0)
