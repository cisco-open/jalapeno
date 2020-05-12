#! /usr/bin/env python
"""This script creates network ls-topology records in an "LS_Topology" collection in ArangoDB.

Given configuration information (database connection parameters, set in arangoconfig)
the "LS_Topology" collection will be created or joined.

The "LS_Topology" documents will then be created from existing (seemingly unrelated) data from various
Arango collections. Relevant data will be collected and organized, corresponding
"LS_Topology" documents will be created, and finally, the "LS_Topology" documents will be
upserted into the "LS_Topology" collection.
"""

from pyArango.connection import *
from configs import arangoconfig
from util import connections
import logging, time, json, sys
from ls_topology_queries import *

def main():
    setup_logging()
    logging.info('Creating connection to Arango')
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    logging.info('Creating collection in Arango')
    collection_name = "LS_Topology"
    collection = create_collection(database, collection_name)
    while(True):
        ls_link_keys = get_ls_link_keys(database)
        for ls_link_index in range(len(ls_link_keys)):
            current_ls_link_key = ls_link_keys[ls_link_index]
            ls_link_exists = check_exists_ls_topology(database, current_ls_link_key)
            if(ls_link_exists):
                updateBaseLSTopologyDocument(database, current_ls_link_key)
            else:
                createBaseLSTopologyDocument(database, current_ls_link_key)

        ls_topology_keys = get_ls_topology_keys(database)
        for ls_topology_index in range(len(ls_topology_keys)):
            current_ls_topology_key = ls_topology_keys[ls_topology_index]
            enhance_ls_topology_document(database, current_ls_topology_key)
            local_node = get_local_igpid(database, current_ls_topology_key)[0]
            remote_node = get_remote_igpid(database, current_ls_topology_key)[0]
            local_igpid, remote_igpid = local_node["LocalIGPID"], remote_node["RemoteIGPID"]
            local_srgb_start = get_srgb_start(database, local_igpid)[0]
            remote_srgb_start = get_srgb_start(database, remote_igpid)[0]
            local_msd = get_max_sid_depth(database, local_igpid)
            remote_msd = get_max_sid_depth(database, remote_igpid)
            local_max_sid_depth = handle_msd(local_msd)
            remote_max_sid_depth = handle_msd(remote_msd)
            local_prefix_info = get_prefix_info(database, local_igpid)
            remote_prefix_info = get_prefix_info(database, remote_igpid)
            local_prefix_sid, local_prefixes = parse_prefix_info(local_prefix_info, local_srgb_start)
            remote_prefix_sid, remote_prefixes = parse_prefix_info(remote_prefix_info, remote_srgb_start)
            update_ls_topology_document(database, current_ls_topology_key, local_prefix_sid, remote_prefix_sid, local_prefixes, remote_prefixes, local_max_sid_depth, remote_max_sid_depth)
        time.sleep(10)

def handle_msd(max_sid_depth):
    msd = ""
    if(len(max_sid_depth) > 0) and max_sid_depth[0] != None:
        max_sid_depth_split = max_sid_depth[0].split(":")
        msd = max_sid_depth_split[1]
    return msd

def parse_prefix_info(prefix_info, srgb_start):
    prefix_info_list = []
    prefix_sid = None
    for index in range(len(prefix_info)):
        sid_index = prefix_info[index]["SIDIndex"][0]
        prefix = prefix_info[index]["Prefix"]
        length = prefix_info[index]["Length"]
        if(prefix_info[index]["SRFlag"] != None and prefix_info[index]["SRFlag"][0] == "n"):
            prefix_sid = srgb_start + sid_index
        sid = srgb_start + sid_index
        prefix_dict = {"Prefix": prefix, "Length": length, "SID": sid}
        prefix_info_list.append(prefix_dict)
    return(prefix_sid, prefix_info_list)

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
