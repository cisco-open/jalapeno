#! /usr/bin/env python
"""This class upserts InternalLinks (any links within the network) and their performance metrics into the InternalLinks_Performance collection in ArangoDB"""
from configs import arangoconfig
from util import connections

def upsert_internal_link_performance(key, internal_router, internal_router_interface, in_unicast_pkts, out_unicast_pkts,
                                     in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                                     in_discards, out_discards, in_errors, out_errors, in_octets, out_octets):
    """Insert or update InternalLinks with their performance measurements in the InternalLinks_Performance collection."""
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    internal_link_performance_key_exists = check_existing_internal_link_performance(arango_client, key)
    if(internal_link_performance_key_exists):
        update_internal_link_performance(arango_client, key, internal_router, internal_router_interface, in_unicast_pkts,
                                         out_unicast_pkts, in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts,
                                         out_broadcast_pkts, in_discards, out_discards, in_errors, out_errors, in_octets, out_octets)
    else:
        insert_internal_link_performance(arango_client, key, internal_router, internal_router_interface, in_unicast_pkts,
                                         out_unicast_pkts, in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts,
                                         out_broadcast_pkts, in_discards, out_discards, in_errors, out_errors, in_octets, out_octets)


def update_internal_link_performance(arango_client, key, internal_router, internal_router_interface, in_unicast_pkts,
                                     out_unicast_pkts, in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts,
                                     out_broadcast_pkts, in_discards, out_discards, in_errors, out_errors, in_octets, out_octets):
    """Update existing InternalLinks with new performance data in the InternalLinks_Performance collection (specified by key)."""
    print("Updating existing InternalLinks_Performance record " + key + " with performance metrics")
    aql = """FOR p in InternalLinks_Performance
        FILTER p._key == @key
        UPDATE p with { Internal_Router: @internal_router, Internal_Router_Interface: @internal_router_interface,
                        in_unicast_pkts: @in_unicast_pkts, out_unicast_pkts: @out_unicast_pkts,
                        in_multicast_pkts: @in_multicast_pkts, out_multicast_pkts: @out_multicast_pkts,
                        in_broadcast_pkts: @in_broadcast_pkts, out_broadcast_pkts: @out_broadcast_pkts,
                        in_discards: @in_discards, out_discards: @out_discards,
                        in_errors: @in_errors, out_errors: @out_errors,
                        in_octets: @in_octets, out_octets: @out_octets } in InternalLinks_Performance"""
    bindVars = {'key': key, 'internal_router': internal_router, 'internal_router_interface': internal_router_interface,
                        'in_unicast_pkts': in_unicast_pkts, 'out_unicast_pkts': out_unicast_pkts,
                        'in_multicast_pkts': in_multicast_pkts, 'out_multicast_pkts': out_multicast_pkts,
                        'in_broadcast_pkts': in_broadcast_pkts, 'out_broadcast_pkts': out_broadcast_pkts,
                        'in_discards': in_discards, 'out_discards': out_discards,
                        'in_errors': in_errors, 'out_errors': out_errors,
                        'in_octets': in_octets, 'out_octets': out_octets}
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)

def insert_internal_link_performance(arango_client, key, internal_router, internal_router_interface, in_unicast_pkts,
                                   out_unicast_pkts, in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts,
                                   out_broadcast_pkts, in_discards, out_discards, in_errors, out_errors, in_octets, out_octets):
    """Insert new InternalLinks with their performance data in the InternalLinks_Performance collection."""
    print("Inserting InternalLinks_Performance record " + key + " with performance metrics")
    aql = """INSERT { _key: @key, Internal_Router: @internal_router, Internal_Router_Interface: @internal_router_interface,
                        in_unicast_pkts: @in_unicast_pkts, out_unicast_pkts: @out_unicast_pkts,
                        in_multicast_pkts: @in_multicast_pkts, out_multicast_pkts: @out_multicast_pkts,
                        in_broadcast_pkts: @in_broadcast_pkts, out_broadcast_pkts: @out_broadcast_pkts,
                        in_discards: @in_discards, out_discards: @out_discards,
                        in_errors: @in_errors, out_errors: @out_errors, 
                        in_octets: @in_octets, out_octets: @out_octets } into InternalLinks_Performance RETURN { after: NEW }"""
    bindVars = {'key': key, 'internal_router': internal_router, 'internal_router_interface': internal_router_interface,
                        'in_unicast_pkts': in_unicast_pkts, 'out_unicast_pkts': out_unicast_pkts,
                        'in_multicast_pkts': in_multicast_pkts, 'out_multicast_pkts': out_multicast_pkts,
                        'in_broadcast_pkts': in_broadcast_pkts, 'out_broadcast_pkts': out_broadcast_pkts,
                        'in_discards': in_discards, 'out_discards': out_discards,
                        'in_errors': in_errors, 'out_errors': out_errors,
                        'in_octets': in_octets, 'out_octets': out_octets}
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)

def check_existing_internal_link_performance(arango_client, key):
    internal_link_performance_key_exists = False
    aql = """FOR e in InternalLinks_Performance
        FILTER e._key == @key
        RETURN { key: e._key }"""
    bindVars = {'key': key}
    result = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(result) > 0):
        internal_link_performance_key_exists = True
    return internal_link_performance_key_exists
