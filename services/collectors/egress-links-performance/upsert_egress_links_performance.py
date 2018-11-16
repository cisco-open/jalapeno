#! /usr/bin/env python
"""This class upserts EgressLinks (any links out of the network) and their performance metrics into the EgressLinks_Performance collection in ArangoDB"""
from configs import arangoconfig
from util import connections

def upsert_egress_link_performance(key, egress_router, egress_router_interface, in_unicast_pkts, out_unicast_pkts,
                                   in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                                   in_discards, out_discards, in_errors, out_errors, in_octets, out_octets):
    """Insert or update EgressLinks with their performance measurements in the EgressLinks_Performance collection."""
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    egress_link_performance_key_exists = check_existing_egress_link_performance(arango_client, key)
    if(egress_link_performance_key_exists):
        update_egress_link_performance(arango_client, key, egress_router, egress_router_interface, in_unicast_pkts,
                                   out_unicast_pkts, in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts,
                                   out_broadcast_pkts, in_discards, out_discards, in_errors, out_errors, in_octets, out_octets)
    else:
        insert_egress_link_performance(arango_client, key, egress_router, egress_router_interface, in_unicast_pkts,
                                   out_unicast_pkts, in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts,
                                   out_broadcast_pkts, in_discards, out_discards, in_errors, out_errors, in_octets, out_octets)

 
def update_egress_link_performance(arango_client, key, egress_router, egress_router_interface, in_unicast_pkts,
                                   out_unicast_pkts, in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, 
                                   out_broadcast_pkts, in_discards, out_discards, in_errors, out_errors, in_octets, out_octets):
    """Update existing EgressLinks with new performance data in the EgressLinks_Performance collection (specified by key)."""
    print("Updating existing EgressLinks_Performance record " + key + " with performance metrics")
    aql = """FOR p in EgressLinks_Performance
        FILTER p._key == @key
        UPDATE p with { Egress_Router: @egress_router, Egress_Router_Interface: @egress_router_interface, 
                        in_unicast_pkts: @in_unicast_pkts, out_unicast_pkts: @out_unicast_pkts,
                        in_multicast_pkts: @in_multicast_pkts, out_multicast_pkts: @out_multicast_pkts,
                        in_broadcast_pkts: @in_broadcast_pkts, out_broadcast_pkts: @out_broadcast_pkts,
                        in_discards: @in_discards, out_discards: @out_discards,
                        in_errors: @in_errors, out_errors: @out_errors, 
                        in_octets: @in_octets, out_octets: @out_octets } in EgressLinks_Performance"""
    bindVars = {'key': key, 'egress_router': egress_router, 'egress_router_interface': egress_router_interface,
                        'in_unicast_pkts': in_unicast_pkts, 'out_unicast_pkts': out_unicast_pkts,
                        'in_multicast_pkts': in_multicast_pkts, 'out_multicast_pkts': out_multicast_pkts,
                        'in_broadcast_pkts': in_broadcast_pkts, 'out_broadcast_pkts': out_broadcast_pkts,
                        'in_discards': in_discards, 'out_discards': out_discards,
                        'in_errors': in_errors, 'out_errors': out_errors,
                        'in_octets': in_octets, 'out_octets': out_octets} 
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)


def insert_egress_link_performance(arango_client, key, egress_router, egress_router_interface, in_unicast_pkts,
                                   out_unicast_pkts, in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts,
                                   out_broadcast_pkts, in_discards, out_discards, in_errors, out_errors, in_octets, out_octets):
    """Insert new EgressLinks with their performance data in the EgressLinks_Performance collection."""
    print("Inserting EgressLinks_Performance record " + key + " with performance metrics")
    aql = """INSERT { _key: @key, Egress_Router: @egress_router, Egress_Router_Interface: @egress_router_interface,
                        in_unicast_pkts: @in_unicast_pkts, out_unicast_pkts: @out_unicast_pkts,
                        in_multicast_pkts: @in_multicast_pkts, out_multicast_pkts: @out_multicast_pkts,
                        in_broadcast_pkts: @in_broadcast_pkts, out_broadcast_pkts: @out_broadcast_pkts,
                        in_discards: @in_discards, out_discards: @out_discards,
                        in_errors: @in_errors, out_errors: @out_errors,
                        in_octets: @in_octets, out_octets: @out_octets } into EgressLinks_Performance RETURN { after: NEW }"""
    bindVars = {'key': key, 'egress_router': egress_router, 'egress_router_interface': egress_router_interface,
                        'in_unicast_pkts': in_unicast_pkts, 'out_unicast_pkts': out_unicast_pkts,
                        'in_multicast_pkts': in_multicast_pkts, 'out_multicast_pkts': out_multicast_pkts,
                        'in_broadcast_pkts': in_broadcast_pkts, 'out_broadcast_pkts': out_broadcast_pkts,
                        'in_discards': in_discards, 'out_discards': out_discards,
                        'in_errors': in_errors, 'out_errors': out_errors,
                        'in_octets': in_octets, 'out_octets': out_octets} 
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)

def check_existing_egress_link_performance(arango_client, key):
    egress_link_performance_key_exists = False
    aql = """FOR e in EgressLinks_Performance
        FILTER e._key == @key
        RETURN { key: e._key }"""
    bindVars = {'key': key}
    result = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(result) > 0):
        egress_link_performance_key_exists = True
    return egress_link_performance_key_exists
