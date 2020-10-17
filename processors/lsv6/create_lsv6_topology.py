#! /usr/bin/env python
"""This script creates network lsv6-topology records in an "LSv6_Topology" collection in ArangoDB.

Given configuration information (database connection parameters, set in arangoconfig)
the "LSv6_Topology" collection will be created or joined.

The "LSv6_Topology" documents will then be created from existing (seemingly unrelated) data from various
Arango collections. Relevant data will be collected and organized, corresponding
"LSv6_Topology" documents will be created, and finally, the "LSv6_Topology" documents will be
upserted into the "LSv6_Topology" collection.
"""

from pyArango.connection import *
from configs import arangoconfig
from util import connections
import logging, time, json, sys
from lsv6_topology_queries import *

def main():
    setup_logging()
    logging.info('Creating connection to Arango')
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    logging.info('Creating collection in Arango')
    collection_name = "LSv6_Topology"
    collection = create_collection(database, collection_name)
    while(True):
        lsv6_link_keys = get_lsv6_link_keys(database)
        lsv6_topology_keys = get_lsv6_topology_keys(database)

        for lsv6_topology_index in range(len(lsv6_topology_keys)):
            current_lsv6_topology_key = lsv6_topology_keys[lsv6_topology_index]
            if(check_exists_lsv6_link(database, current_lsv6_topology_key) == False):
                deleteLSv6TopologyDocument(database, current_lsv6_topology_key)

        for lsv6_link_index in range(len(lsv6_link_keys)):
            current_lsv6_link_key = lsv6_link_keys[lsv6_link_index]
            lsv6_link_exists = check_exists_lsv6_topology(database, current_lsv6_link_key)
            if(lsv6_link_exists):
                updateBaseLSv6TopologyDocument(database, current_lsv6_link_key)
            else:
                createBaseLSv6TopologyDocument(database, current_lsv6_link_key)

        lsv6_topology_keys = get_lsv6_topology_keys(database)
        for lsv6_topology_index in range(len(lsv6_topology_keys)):
            current_lsv6_topology_key = lsv6_topology_keys[lsv6_topology_index]
            enhance_lsv6_topology_document(database, current_lsv6_topology_key)
            local_node = get_local_igpid(database, current_lsv6_topology_key)[0]
            remote_node = get_remote_igpid(database, current_lsv6_topology_key)[0]
            local_igp_id, remote_igp_id = local_node["local_igp_id"], remote_node["remote_igp_id"]
            #local_srgb_start = get_srgb_start(database, local_igp_id)[0]
            #remote_srgb_start = get_srgb_start(database, remote_igp_id)[0]
            #local_msd = get_max_sid_depth(database, local_igp_id)
            #remote_msd = get_max_sid_depth(database, remote_igp_id)
            #local_max_sid_depth = handle_msd(local_msd)
            #remote_max_sid_depth = handle_msd(remote_msd)
            #local_prefix_info = get_prefix_info(database, local_igpid)
            #remote_prefix_info = get_prefix_info(database, remote_igpid)
            #local_prefix_sid, local_prefixes = parse_prefix_info(local_prefix_info, local_srgb_start)
            #remote_prefix_sid, remote_prefixes = parse_prefix_info(remote_prefix_info, remote_srgb_start)
            local_srv6_info = get_srv6_info(database, local_igp_id)[0]
            remote_srv6_info = get_srv6_info(database, remote_igp_id)[0]
            if(len(local_srv6_info) == 0):
                local_srv6_info = handle_empty_srv6()
            if(len(remote_srv6_info) == 0):
                remote_srv6_info = handle_empty_srv6()
            update_lsv6_topology_document(database, current_lsv6_topology_key, local_srv6_info, remote_srv6_info)
        time.sleep(10)

def handle_empty_srv6():
    srv6_info = {"protocol": "", "mt_id": "", "srv6_isd": "", "srv6_endpoint_behavior": "", "srv6_sid_structure": ""}
    return srv6_info

def create_collection(db, collection_name):
    """Create new collection in ArangoDB.
    If the collection exists, connect to that collection.
    """
    database = db
    print("Creating " + collection_name + " collection in Arango")
    try:
        collection = database.createCollection(className='Edges', name=collection_name)
    except CreationError:
        print(collection_name + " collection exists: entering collection.")
        collection = database[collection_name]
    return collection

def setup_logging():
    logging.getLogger().setLevel(logging.WARNING)
    logging.getLogger("requests").setLevel(logging.WARNING)
    logging.getLogger("urllib3").setLevel(logging.WARNING)

if __name__ == '__main__':
    main()
    exit(0)
