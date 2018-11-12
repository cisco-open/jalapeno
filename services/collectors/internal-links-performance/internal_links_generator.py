#! /usr/bin/env python
"""This script collects all Internal-Links. These are defined physically as "InternalRouterInterfaces" in ArangoDB.
The collection is reached given configuration information set in arangoconfig (database connection parameters).
"""
import logging, os
from configs import arangoconfig, queryconfig
from util import connections

def generate_internal_links(arango_client):
    """Connect to Arango using parameters in arangoconfig.
    Collect Internal-Links from InternalRouterInterfaces collection in ArangoDB.
    """
    #print("\nGenerating all Internal-Links")
    internal_links = generate_internal_links_query(arango_client)
    return internal_links

def generate_internal_links_query(arango_client):
    """AQL Query to collect Internal-Link information from the InternalRouterInterfaces collection in Arango."""
    aql = """FOR e in InternalRouterInterfaces
        RETURN { key: e._key, RouterIP: e.RouterIP, InterfaceIP: e.RouterInterfaceIP }"""
    bindVars = {}
    internal_links = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return internal_links

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    setup_logging()
    logging.info('Creating connection to Arango')
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    internal_links = generate_internal_links(arango_client)
    for link in internal_links:
        print(link)
    print("--------------------------------------------------------------------------------")
    exit(0)
