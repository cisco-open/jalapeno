#! /usr/bin/env python
"""This script collects all ExternalLinks (defined physically as BorderRouterInterfaces in ArangoDB). 
It also collects ExternalLinkEdges (given a source router and interface ip). 
The collections are reached given configuration information set in arangoconfig (database connection parameters).
"""
import logging, os
from configs import arangoconfig, queryconfig
from util import connections

def generate_external_links(arango_client):
    """Connect to Arango using parameters in arangoconfig.
    Collect ExternalLinks from BorderRouterInterfaces collection in ArangoDB.
    """
    #print("\nGenerating all ExternalLinks")
    external_links = generate_external_links_query(arango_client)
    return external_links

def generate_external_links_query(arango_client):
    """AQL Query to collect ExternalLink information from the BorderRouterInterfaces collection in Arango."""
    aql = """FOR e in BorderRouterInterfaces
        RETURN { key: e._key, RouterIP: e.RouterIP, InterfaceIP: e.RouterInterfaceIP }"""
    bindVars = {}
    external_links = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return external_links

def generate_external_link_edges(arango_client, router_ip, router_interface_ip):
    """Connect to Arango using parameters in arangoconfig.
    Collect ExternalLinkEdges given a Source RouterIP and InterfaceIP.
    """
    #print("\nGenerating all ExternalLinkEdges")
    external_link_edges = generate_external_link_edges_query(arango_client, router_ip, router_interface_ip)
    return external_link_edges

def generate_external_link_edges_query(arango_client, router_ip, router_interface_ip):
    """AQL Query to collect ExternalLinkEdges given a Source RouterIP and InterfaceIP from the ExternalLinkEdges collection in Arango."""
    aql = """FOR e in ExternalLinkEdges
        FILTER e.Source == @router_ip AND e.SrcInterfaceIP == @router_interface_ip
        RETURN { key: e._key, source: e.Source, source_intf_ip: e.SrcInterfaceIP, destination: e.Destination, dest_intf_ip: e.DstInterfaceIP }"""
    bindVars = {'router_ip': "Routers/"+router_ip, 'router_interface_ip': router_interface_ip }
    external_link_edges = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return external_link_edges
   
def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    setup_logging()
    logging.info('Creating connection to Arango')
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    external_links = generate_external_links(arango_client)
    external_link_edges = generate_external_link_edges(arango_client, "10.0.0.1", "2.2.72.0")
    for link in external_links:
        print(link)
    for external_link_edge in external_link_edges:
        print(external_link_edge)
    print("--------------------------------------------------------------------------------")
    exit(0)
