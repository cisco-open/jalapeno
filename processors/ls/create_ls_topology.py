#! /usr/bin/env python
"""This script creates network ls-topology records in an "LS_Topology" collection in ArangoDB.

Given configuration information (database connection parameters, set in arangoconfig)
the "LS_Topology" collection will be created or joined.

Given configuration information (query parameters, set in queryconfig), all "LS_Topology"
documents will then be created from existing (seemingly unrelated) data from various 
Arango collections. Relevant data will be collected and organized, corresponding 
"LS_Topology" documents will be created, and finally, the "LS_Topology" documents will be 
upserted into the "LS_Topology" collection.
"""

from pyArango.connection import *
from configs import arangoconfig, queryconfig
from util import connections
import logging, time, json, sys
from ls_topology_queries import *

def main():
    setup_logging()
    logging.info('Creating connection to Arango')
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    logging.info('Creating collection in Arango')
    collection = create_collection(database, queryconfig.collection)  # the collection name is set in queryconfig
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
            local_node = get_local_node(database, current_ls_topology_key)
            remote_node = get_remote_node(database, current_ls_topology_key)
            local_prefix_sid = get_prefix_sid(database, str(local_node[0]))
            remote_prefix_sid = get_prefix_sid(database, str(remote_node[0]))
            update_prefix_sid(database, current_ls_topology_key, str(local_prefix_sid[0]), str(remote_prefix_sid[0]))

        time.sleep(10)
 
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
