#! /usr/bin/env python
"""This script collects all EPEPaths.
Given configuration information set in arangoconfig (database connection parameters),
the "EPEPaths" collection will be used to collect EPEPath data.
"""
import logging, os
from configs import arangoconfig, queryconfig
from util import connections

def generate_epe_paths(arango_client):
    """Connect to Arango using parameters in arangoconfig.
    Collect EPEPath information from EPEPaths collection in ArangoDB.
    """
    #print("\nGenerating all EPEPath information")
    epe_paths = generate_epe_paths_query(arango_client)
    return epe_paths

def generate_epe_paths_query(arango_client):
    """AQL Query to collect EPEPath information from the EPEPaths collection in Arango."""
    aql = """FOR e in EPEPaths
        RETURN { Key: e._key, Egress_Peer: e.Egress_Peer, Egress_Interface: e.Egress_Interface, Label_Path: e.Label_Path, Destination: e.Destination}"""
    bindVars = {}
    epe_paths = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return epe_paths

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    setup_logging()
    logging.info('Creating connection to Arango')
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    epe_paths = generate_epe_paths(arango_client)
    for path in epe_paths:
        print(path)
    print("--------------------------------------------------------------------------------")
    exit(0)
