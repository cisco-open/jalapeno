#! /usr/bin/env python
"""This class upserts BorderRouterInterfaces/ExternalLinkEdges (any interfaces/links out of the network) 
and their performance metrics into the BorderRouterInterfaces/ExternalLinkEdges collections in ArangoDB"""
from configs import arangoconfig
from util import connections

def upsert_external_link_performance(collection_name, key, in_unicast_pkts, out_unicast_pkts, in_multicast_pkts,
                                     out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts, in_discards,
                                     out_discards, in_errors, out_errors, in_octets, out_octets):
    """Insert or update performance measurements into the specified collection."""
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    external_link_performance_key_exists = check_existing_external_link_performance(arango_client, collection_name, key)
    if(external_link_performance_key_exists):
        update_external_link_performance(arango_client, collection_name, key, in_unicast_pkts, out_unicast_pkts,
                                         in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                                         in_discards, out_discards, in_errors, out_errors, in_octets, out_octets)
    else:
        insert_external_link_performance(arango_client, collection_name, key, in_unicast_pkts, out_unicast_pkts,
                                         in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                                         in_discards, out_discards, in_errors, out_errors, in_octets, out_octets)

 
def update_external_link_performance(arango_client, collection_name, key, in_unicast_pkts, out_unicast_pkts,
                                     in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                                     in_discards, out_discards, in_errors, out_errors, in_octets, out_octets):
    """Update specified collection document with new performance data (document specified by key)."""
    print("Updating existing " + collection_name + " record " + key + " with performance metrics")
    aql = """FOR p in @@collection
        FILTER p._key == @key
        UPDATE p with { in_unicast_pkts: @in_unicast_pkts, out_unicast_pkts: @out_unicast_pkts,
                        in_multicast_pkts: @in_multicast_pkts, out_multicast_pkts: @out_multicast_pkts,
                        in_broadcast_pkts: @in_broadcast_pkts, out_broadcast_pkts: @out_broadcast_pkts,
                        in_discards: @in_discards, out_discards: @out_discards,
                        in_errors: @in_errors, out_errors: @out_errors, 
                        in_octets: @in_octets, out_octets: @out_octets } in @@collection"""
    bindVars = {'@collection': collection_name, 'key': key, 
                        'in_unicast_pkts': in_unicast_pkts, 'out_unicast_pkts': out_unicast_pkts,
                        'in_multicast_pkts': in_multicast_pkts, 'out_multicast_pkts': out_multicast_pkts,
                        'in_broadcast_pkts': in_broadcast_pkts, 'out_broadcast_pkts': out_broadcast_pkts,
                        'in_discards': in_discards, 'out_discards': out_discards,
                        'in_errors': in_errors, 'out_errors': out_errors,
                        'in_octets': in_octets, 'out_octets': out_octets} 
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)


def insert_external_link_performance(arango_client, collection_name, key, in_unicast_pkts, out_unicast_pkts,
                                     in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                                     in_discards, out_discards, in_errors, out_errors, in_octets, out_octets):
    """Insert performance data into the given collection given a key."""
    print("Inserting " + collection_name + " record " + key + " with performance metrics")
    aql = """INSERT { _key: @key, in_unicast_pkts: @in_unicast_pkts, out_unicast_pkts: @out_unicast_pkts,
                        in_multicast_pkts: @in_multicast_pkts, out_multicast_pkts: @out_multicast_pkts,
                        in_broadcast_pkts: @in_broadcast_pkts, out_broadcast_pkts: @out_broadcast_pkts,
                        in_discards: @in_discards, out_discards: @out_discards,
                        in_errors: @in_errors, out_errors: @out_errors,
                        in_octets: @in_octets, out_octets: @out_octets } into @@collection RETURN { after: NEW }"""
    bindVars = {'@collection': collection_name, 'key': key,
                        'in_unicast_pkts': in_unicast_pkts, 'out_unicast_pkts': out_unicast_pkts,
                        'in_multicast_pkts': in_multicast_pkts, 'out_multicast_pkts': out_multicast_pkts,
                        'in_broadcast_pkts': in_broadcast_pkts, 'out_broadcast_pkts': out_broadcast_pkts,
                        'in_discards': in_discards, 'out_discards': out_discards,
                        'in_errors': in_errors, 'out_errors': out_errors,
                        'in_octets': in_octets, 'out_octets': out_octets} 
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)

def check_existing_external_link_performance(arango_client, collection_name, key):
    external_link_performance_key_exists = False
    aql = """FOR e in @@collection
        FILTER e._key == @key
        RETURN { key: e._key }"""
    bindVars = {'@collection': collection_name, 'key': key}
    result = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(result) > 0):
        external_link_performance_key_exists = True
    return external_link_performance_key_exists
