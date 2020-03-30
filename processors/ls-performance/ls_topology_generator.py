#! /usr/bin/env python
"""This script collects all edges in the LS_Topology collection.
The collections are reached given configuration information set in arangoconfig (database connection parameters).
"""
import logging, os
from configs import arangoconfig
from util import connections

def generate_ls_topology(arango_client):
    """Connect to Arango using parameters in arangoconfig.
    Collect InternalLinks from InternalRouterInterfaces collection in ArangoDB.
    """
    #print("\nGenerating all InternalLinks")
    ls_topology = generate_ls_topology_query(arango_client)
    return ls_topology

def generate_ls_topology_query(arango_client):
    """AQL Query to collect LS_Topology information from the LS_Topology collection in Arango."""
    aql = """FOR e in LS_Topology
        RETURN { key: e._key, RouterID: e.LocalRouterID, InterfaceIP: e.FromInterfaceIP }"""
    bindVars = {}
    ls_topology = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return ls_topology

def get_node_hostname(arango_client, router_ip):
    aql = """FOR n in LSNode
        FILTER n._key == @ls_node_key
        RETURN n.Name"""
    bindVars = {"ls_node_key": router_ip}
    node_hostname = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return node_hostname[0]

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    setup_logging()
    logging.info('Creating connection to Arango')
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    ls_topology = generate_ls_topology(arango_client)
    for link in range(len(ls_topology)):
        print(ls_topology[link])
    print("--------------------------------------------------------------------------------")
    exit(0)
