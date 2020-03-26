#! /usr/bin/env python
"""This script collects all InternalLinks (defined as "InternalRouterInterfaces" in Arango).
It also collects InternalLinkEdges (given a source router and interface ip).
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

def generate_internal_link_edges(arango_client, router_ip, router_interface_ip):
    """Connect to Arango using parameters in arangoconfig.
    Collect InternalLinkEdges given a Source RouterIP and InterfaceIP.
    """
    #print("\nGenerating all InternalLinkEdges")
    internal_link_edges = generate_internal_link_edges_query(arango_client, router_ip, router_interface_ip)
    return internal_link_edges

def generate_internal_link_edges_query(arango_client, router_ip, router_interface_ip):
    """AQL Query to collect InternalLinkEdges given a Source RouterIP and InterfaceIP from the InternalLinkEdges collection in Arango."""
    aql = """FOR e in InternalLinkEdges
        FILTER e._from == @router_ip AND e.SrcInterfaceIP == @router_interface_ip
        RETURN { key: e._key, source: e._from, source_intf_ip: e.SrcInterfaceIP, destination: e._to, dest_intf_ip: e.DstInterfaceIP }"""
    bindVars = {'router_ip': "Routers/"+router_ip, 'router_interface_ip': router_interface_ip }
    internal_link_edges = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return internal_link_edges

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    setup_logging()
    logging.info('Creating connection to Arango')
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    internal_links = generate_internal_links(arango_client)
    internal_link_edges = generate_internal_link_edges(arango_client, "10.0.0.0", "10.1.1.2") # sample execution
    for link in internal_links:
        print(link)
    for internal_link_edge in internal_link_edges:
        print(internal_link_edge)
    print("--------------------------------------------------------------------------------")
    exit(0)
