#! /usr/bin/env python
"""This script collects all Egress-Links. These are defined physically as "PeeringRouterInterfaces" in ArangoDB.
The collection is reached given configuration information set in arangoconfig (database connection parameters).
"""
import logging, os
from configs import arangoconfig, queryconfig
from util import connections

def generate_egress_links(arango_client):
    """Connect to Arango using parameters in arangoconfig.
    Collect Egress-Links from PeeringRouterInterfaces collection in ArangoDB.
    """
    #print("\nGenerating all Egress-Links")
    egress_links = generate_egress_links_query(arango_client)
    return egress_links

def generate_egress_links_query(arango_client):
    """AQL Query to collect Egress-Link information from the PeeringRouterInterfaces collection in Arango."""
    aql = """FOR e in PeeringRouterInterfaces
        RETURN { key: e._key, RouterIP: e.RouterIP, InterfaceIP: e.RouterInterfaceIP }"""
    bindVars = {}
    egress_links = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return egress_links

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    setup_logging()
    logging.info('Creating connection to Arango')
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    egress_links = generate_egress_links(arango_client)
    for link in egress_links:
        print(link)
    print("--------------------------------------------------------------------------------")
    exit(0)
