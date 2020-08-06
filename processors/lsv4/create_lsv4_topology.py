#! /usr/bin/env python
"""This script creates network lsv4-topology records in an "LSv4_Topology" collection in ArangoDB.

Given configuration information (database connection parameters, set in arangoconfig)
the "LSv4_Topology" collection will be created or joined.

The "LSv4_Topology" documents will then be created from existing (seemingly unrelated) data from various
Arango collections. Relevant data will be collected and organized, corresponding
"LSv4_Topology" documents will be created, and finally, the "LSv4_Topology" documents will be
upserted into the "LSv4_Topology" collection.
"""

from pyArango.connection import *
from configs import arangoconfig
from util import connections
import logging, time, json, sys
from lsv4_topology_queries import *

def main():
    setup_logging()
    logging.info('Creating connection to Arango')
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    logging.info('Creating collection in Arango')
    collection_name = "LSv4_Topology"
    collection = create_collection(database, collection_name)
    while(True):
        lsv4_link_keys = get_lsv4_link_keys(database)
        lsv4_topology_keys = get_lsv4_topology_keys(database)

        for lsv4_topology_index in range(len(lsv4_topology_keys)):
            current_lsv4_topology_key = lsv4_topology_keys[lsv4_topology_index]
            if(check_exists_lsv4_link(database, current_lsv4_topology_key) == False):
                deleteLSv4TopologyDocument(database, current_lsv4_topology_key)

        for lsv4_link_index in range(len(lsv4_link_keys)):
            current_ls_link_key = lsv4_link_keys[lsv4_link_index]
            ls_link_exists = check_exists_lsv4_topology(database, current_ls_link_key)
            if(ls_link_exists):
                updateBaseLSv4TopologyDocument(database, current_ls_link_key)
            else:
                createBaseLSv4TopologyDocument(database, current_ls_link_key)

        lsv4_topology_keys = get_lsv4_topology_keys(database)
        for lsv4_topology_index in range(len(lsv4_topology_keys)):
            current_lsv4_topology_key = lsv4_topology_keys[lsv4_topology_index]
            enhance_lsv4_topology_document(database, current_lsv4_topology_key)
            local_node = get_local_igpid(database, current_lsv4_topology_key)[0]
            remote_node = get_remote_igpid(database, current_lsv4_topology_key)[0]
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
            update_lsv4_topology_document(database, current_lsv4_topology_key, local_prefix_sid, remote_prefix_sid, local_prefixes, remote_prefixes, local_max_sid_depth, remote_max_sid_depth)
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
        sid_index = prefix_info[index]["SIDIndex"]
        prefix = prefix_info[index]["Prefix"]
        length = prefix_info[index]["Length"]
        sr_flag = prefix_info[index]["SRFlag"]
        if(prefix_info[index]["SRFlag"] != None and prefix_info[index]["SRFlag"][0] == "n"):
            prefix_sid = srgb_start + sid_index
        sid = srgb_start + sid_index
        prefix_dict = {"Prefix": prefix, "Length": length, "SID": sid, "SRFlag": sr_flag}
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
