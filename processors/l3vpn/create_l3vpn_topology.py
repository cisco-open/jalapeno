#! /usr/bin/env python
"""This script creates network l3vpn-topology records in an "L3VPN_Topology" collection in ArangoDB.

Given configuration information (database connection parameters, set in arangoconfig)
the "L3VPN_Topology" collection will be created or joined.

All "L3VPN_Topology" documents will then be created from existing (seemingly unrelated) data from various 
Arango collections. Relevant data will be collected and organized, corresponding 
"L3VPN_Topology" documents will be created, and finally, the "L3VPN_Topology" documents will be 
upserted into the "L3VPN_Topology" collection.
"""

from pyArango.connection import *
from configs import arangoconfig
from util import connections
import logging, time, json, sys
from l3vpn_topology_queries import *

def main():
    setup_logging()
    logging.info('Creating connection to Arango')
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    logging.info('Creating collection in Arango')
    collection_name = "L3VPN_Topology"
    fib_collection_name = "L3VPN_FIB"
    collection = create_collection(database, collection_name)
    fib_collection = create_collection(database, fib_collection_name)
    while(True):
        print("Creating L3VPN_Topology edges between L3VPNPrefixes and L3VPNNodes")
        create_l3vpnprefix_l3vpnnode_edges(database, collection)
        print("Creating L3VPN_Topology edges between L3VPNNodes and L3VPNNodes")
        create_l3vpnnode_l3vpnnode_edges(database, collection)
        print("Done parsing L3VPN-Topology! Next collection begins in 10 seconds.\n")
        create_l3vpn_fib_edges(database, fib_collection)
        print("Done parsing L3VPN-Topology! Next collection begins in 10 seconds.\n")
        time.sleep(10)

# This function creates the edges from L3VPNPrefixes to L3VNNodes in the L3VPN_Topology collection
def create_l3vpnprefix_l3vpnnode_edges(database, collection):
        all_prefixes = get_prefix_data(database)
        if(len(all_prefixes) == 0):
            print("ALERT: No edge data found -- perhaps the L3VPNPrefix collection is not up or populated?\n")
            return
        for prefix_index in range(len(all_prefixes)):
            current_prefix_document = all_prefixes[prefix_index]
            vpn_prefix = current_prefix_document["Prefix"]
            vpn_prefix_length = current_prefix_document["Length"]
            router_id = current_prefix_document["RouterID"]
            vpn_label = current_prefix_document["VPN_Label"]
            rd = current_prefix_document["RD"]
            rt = current_prefix_document["ExtComm"]
            ipv4 = False
            if(current_prefix_document["IPv4"] == True):
                ipv4 = True
            router_igp_id = get_router_igp_id(database, router_id)
            if(len(router_igp_id) > 0):
                router_igp_id = router_igp_id[0]
            srgb_start = get_srgb_start(database, router_igp_id)
            sid_index = get_sid_index(database, router_igp_id)
            prefixSID = None
            if(len(srgb_start) > 0 and len(sid_index) > 0):
                prefixSID = int(srgb_start[0]) + int(sid_index[0])
            srv6_sid = current_prefix_document["SRv6_SID"]
            print(vpn_prefix, vpn_prefix_length, router_id, prefixSID, vpn_label, rd, rt, ipv4, srv6_sid)
            upsert_l3vpnprefix_l3vpnnode_edge(database, collection, vpn_prefix, vpn_prefix_length, router_id, prefixSID, vpn_label, rd, rt, ipv4, srv6_sid)
            upsert_l3vpnnode_l3vpnprefix_edge(database, collection, vpn_prefix, vpn_prefix_length, router_id, prefixSID, vpn_label, rd, rt, ipv4, srv6_sid)
            print("===========================================================================")

def create_l3vpnnode_l3vpnnode_edges(database, collection):
        # parse existing collections for relevant fields that correlate to a potential L3VPN-Topology Edge
        ## get all RDs that exist in the L3VPNNode collection
        all_rds = get_all_rds(database)
        if(len(all_rds[0]) == 0):
            print("ALERT: No edge data found -- perhaps the L3VPNNode collection is not up or populated?\n")
            return

        l3vpn_rds = all_rds[0]["RDs"]
        #print("we have l3vpn_rds: " + str(l3vpn_rds))

        for l3vpn_rd_index in range(len(l3vpn_rds)):
            l3vpn_rd = l3vpn_rds[l3vpn_rd_index]
            #print("The current l3vpn_rd to be parsed is: " + l3vpn_rd)
            l3vpn_nodes = get_l3vpn_nodes_from_rd(database, l3vpn_rd)
            #print("The l3vpn nodes that are part of this RD are: " + str(l3vpn_nodes))

            """we now have the current RD i.e. 100:100, and a list of l3vpn_nodes, i.e. 
            10.0.0.7, 10.0.0.6, that share the RD and need to be connected to each other"""

            ## Connect L3VPNNodes in list with matching RD to one another
            for l3vpn_node_index in range(len(l3vpn_nodes)):
                l3vpn_node = l3vpn_nodes[l3vpn_node_index]
                #print("The current l3vpn_node to be parsed is: " + str(l3vpn_node))
                for remaining_l3vpn_node_index in range(l3vpn_node_index+1, len(l3vpn_nodes)):
                    l3vpn_destination_node = l3vpn_nodes[remaining_l3vpn_node_index]
                    #print("The current l3vpn_destination_node to be parsed is: " + str(l3vpn_destination_node))
                    # for each record of l3vpn-topology data, create & upsert a corresponding L3VPN-Topology document into the L3VPN-Topology collection
                    upsert_l3vpnnode_l3vpnnode_edge(database, collection, l3vpn_rd, l3vpn_node, l3vpn_destination_node)
                    upsert_l3vpnnode_l3vpnnode_edge(database, collection, l3vpn_rd, l3vpn_destination_node, l3vpn_node)
                    print("===========================================================================")

# This function creates the L3VPN_FIB edges from L3VPNNode to Prefix 
def create_l3vpn_fib_edges(database, fib_collection):
        all_prefixes = get_prefix_data(database)
        if(len(all_prefixes) == 0):
            print("ALERT: No edge data found -- perhaps the L3VPNPrefix collection is not up or populated?\n")
            return
        for prefix_index in range(len(all_prefixes)):
            current_prefix_document = all_prefixes[prefix_index]
            vpn_prefix = current_prefix_document["Prefix"]
            vpn_prefix_length = current_prefix_document["Length"]
            router_id = current_prefix_document["RouterID"]
            vpn_label = current_prefix_document["VPN_Label"]
            rd = current_prefix_document["RD"]
            rt_db = current_prefix_document["ExtComm"]
            rt_chars = ([s.replace('rt=', '') for s in rt_db])
            rt_string = ' '.join([str(elem) for elem in rt_chars])
            rt = rt_string.split(",")
            ipv4 = False
            if(current_prefix_document["IPv4"] == True):
                ipv4 = True
            router_igp_id = get_router_igp_id(database, router_id)
            if(len(router_igp_id) > 0):
                router_igp_id = router_igp_id[0]
            srgb_start = get_srgb_start(database, router_igp_id)
            sid_index = get_sid_index(database, router_igp_id)
            prefixSID = None
            if(len(srgb_start) > 0 and len(sid_index) > 0):
                prefixSID = int(srgb_start[0]) + int(sid_index[0])
            srv6_sid = current_prefix_document["SRv6_SID"]
            print(vpn_prefix, vpn_prefix_length, router_id, prefixSID, vpn_label, rd, rt, ipv4, srv6_sid)
            upsert_l3vpn_fib_edge(database, fib_collection, vpn_prefix, vpn_prefix_length, router_id, prefixSID, vpn_label, rd, rt, ipv4, origin_as, srv6_sid)
            print("===========================================================================")

 
def create_collection(db, collection_name):
    """Create new collection in ArangoDB.
    If the collection exists, connect to that collection.
    """
    database = db
    print("Creating " + collection_name + " collection in Arango")
    try:
        collection = database.createCollection(className='Edges', name=collection_name)
    except CreationError:
        print(collection_name + " collection exists: entering collection.\n")
        collection = database[collection_name]
    return collection

def create_fib_collection(db, fib_collection_name):
    """Create new collection in ArangoDB.
    If the collection exists, connect to that collection.
    """
    database = db
    print("Creating " + fib_collection_name + " collection in Arango")
    try:
        fib_collection = database.createCollection(className='Edges', name=fib_collection_name)
    except CreationError:
        print(fib_collection_name + " collection exists: entering collection.\n")
        fib_collection = database[fib_collection_name]
    return fib_collection


def upsert_l3vpnnode_l3vpnnode_edge(db, collection, rd, source_node, destination_node):
    l3vpn_topology_edge_key = source_node + "_" + rd + "_" + destination_node
    existing_l3vpn_topology_edge = get_l3vpn_topology_edge_key(db, l3vpn_topology_edge_key)
    if len(existing_l3vpn_topology_edge) > 0:
        update_node_to_node_topology_edge_query(db, l3vpn_topology_edge_key, rd, source_node, destination_node)
    else:
        create_node_to_node_topology_edge_query(db, l3vpn_topology_edge_key, rd, source_node, destination_node)

def upsert_l3vpnprefix_l3vpnnode_edge(db, collection, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid):
    l3vpn_topology_edge_key = prefix + "_" + rd + "_" + router_id
    existing_l3vpn_topology_edge = get_l3vpn_topology_edge_key(db, l3vpn_topology_edge_key)
    if len(existing_l3vpn_topology_edge) > 0:
        update_prefix_to_node_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid)
    else:
        create_prefix_to_node_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid)

def upsert_l3vpnnode_l3vpnprefix_edge(db, collection, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4,  srv6_sid):
    l3vpn_topology_edge_key = router_id + "_" + rd + "_" + prefix
    existing_l3vpn_topology_edge = get_l3vpn_topology_edge_key(db, l3vpn_topology_edge_key)
    if len(existing_l3vpn_topology_edge) > 0:
        update_node_to_prefix_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid)
    else:
        create_node_to_prefix_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid)

def upsert_l3vpn_fib_edge(db, fib_collection, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, origin_as, srv6_sid):
    l3vpn_fib_edge_key = router_id + "_" + rd + "_" + prefix
    existing_l3vpn_fib_edge = get_l3vpn_fib_edge_key(db, l3vpn_fib_edge_key)
    if len(existing_l3vpn_fib_edge) > 0:
        update_l3vpn_fib_edge_query(db, l3vpn_fib_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, origin_as, srv6_sid)
    else:
        create_l3vpn_fib_edge_query(db, l3vpn_fib_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, origin_as, srv6_sid)


def setup_logging():
    logging.getLogger().setLevel(logging.WARNING)
    logging.getLogger("requests").setLevel(logging.WARNING)
    logging.getLogger("urllib3").setLevel(logging.WARNING)

if __name__ == '__main__':
    main()
    exit(0)
