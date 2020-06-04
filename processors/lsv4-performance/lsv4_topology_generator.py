#! /usr/bin/env python
"""This script collects all edges in the LSv4_Topology collection.
The collections are reached given configuration information set in arangoconfig (database connection parameters).
"""
import logging, os
from configs import arangoconfig
from util import connections

def generate_lsv4_topology(arango_client):
    """Connect to Arango using parameters in arangoconfig.
    Collect InternalLinks from InternalRouterInterfaces collection in ArangoDB.
    """
    #print("\nGenerating all InternalLinks")
    lsv4_topology = generate_lsv4_topology_query(arango_client)
    return lsv4_topology

def generate_lsv4_topology_query(arango_client):
    """AQL Query to collect LSv4_Topology information from the LSv4_Topology collection in Arango."""
    aql = """FOR e in LSv4_Topology
        RETURN { key: e._key, LocalIGPID: e.LocalIGPID, InterfaceIP: e.FromInterfaceIP }"""
    bindVars = {}
    lsv4_topology = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return lsv4_topology

def get_node_hostname(arango_client, router_igp_id):
    aql = """FOR n in LSNode
        FILTER n._key == @ls_node_key
        RETURN n.Name"""
    bindVars = {"ls_node_key": router_igp_id}
    node_hostname = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return node_hostname[0]

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    setup_logging()
    logging.info('Creating connection to Arango')
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    lsv4_topology = generate_lsv4_topology(arango_client)
    for link in range(len(lsv4_topology)):
        print(lsv4_topology[link])
    print("--------------------------------------------------------------------------------")
    exit(0)
