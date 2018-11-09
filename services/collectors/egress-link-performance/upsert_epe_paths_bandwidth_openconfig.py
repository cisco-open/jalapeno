#! /usr/bin/env python
"""This class upserts EPEPaths and their bandwidths based on OpenConfig telemetry into the EPEPaths_Bandwidth_OpenConfig collection in ArangoDB"""
from configs import arangoconfig
from util import connections

def upsert_epe_path_bandwidth_openconfig(key, egress_peer, egress_interface, labels, destination, bandwidth):
    """Insert or update EPEPaths and their bandwidths in the EPEPaths_Bandwidth_OpenConfig collection."""
    arango_connection = connections.ArangoConn()
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    epe_path_bandwidth_openconfig_key_exists = check_existing_epe_path_bandwidth_openconfig(arango_client, key)
    if(epe_path_bandwidth_openconfig_key_exists):
        update_epe_path_bandwidth_openconfig(arango_client, key, egress_peer, egress_interface, labels, destination, bandwidth)
    else:
        insert_epe_path_bandwidth_openconfig(arango_client, key, egress_peer, egress_interface, labels, destination, bandwidth)
 
def update_epe_path_bandwidth_openconfig(arango_client, key, egress_peer, egress_interface, labels, destination, bandwidth):
    """Update existing EPEPath with bandwidth based on OpenConfig telemetry in EPEPaths_Bandwidth_OpenConfig collection (specified by key)."""
    #print("Updating existing EPEPaths_Bandwidth_OpenConfig record " + key + " with bandwidth " + str(bandwidth))
    aql = """FOR p in EPEPaths_Bandwidth_OpenConfig
        FILTER p._key == @key
        UPDATE p with { Egress_Peer: @egress_peer, Egress_Interface: @egress_interface, 
        Label_Path: @labels, Destination: @destination, Bandwidth: @bandwidth } in EPEPaths_Bandwidth_OpenConfig"""
    bindVars = {'key': key, 'egress_peer': egress_peer, 'egress_interface': egress_interface, 
             'labels': labels, 'destination': destination, 'bandwidth': str(bandwidth)}
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)

def insert_epe_path_bandwidth_openconfig(arango_client, key, egress_peer, egress_interface, labels, destination, bandwidth):
    """Insert new EPEPath with bandwidth based on OpenConfig telemetry in EPEPaths_Bandwidth_OpenConfig collection."""
    #print("Inserting EPEPaths_Bandwidth_OpenConfig record " + key + " with bandwidth " + str(bandwidth))
    aql = """INSERT { _key: @key, Egress_Peer: @egress_peer, Egress_Interface: @egress_interface,
        Label_Path: @labels, Destination: @destination, Bandwidth: @bandwidth } into EPEPaths_Bandwidth_OpenConfig RETURN { after: NEW }"""
    bindVars = {'key': key, 'egress_peer': egress_peer, 'egress_interface': egress_interface, 
             'labels': labels, 'destination': destination, 'bandwidth': str(bandwidth)}
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)

def check_existing_epe_path_bandwidth_openconfig(arango_client, key):
    epe_path_bandwidth_openconfig_key_exists = False
    aql = """FOR e in EPEPaths_Bandwidth_OpenConfig
        FILTER e._key == @key
        RETURN { key: e._key }"""
    bindVars = {'key': key}
    result = arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(result) > 0):
        epe_path_bandwidth_openconfig_key_exists = True
    return epe_path_bandwidth_openconfig_key_exists
