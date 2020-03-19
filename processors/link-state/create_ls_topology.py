#! /usr/bin/env python
"""This script creates network ls-topology records in an "LS_Topology" collection in ArangoDB.

Given configuration information (database connection parameters, set in arangoconfig)
the "LS_Topology" collection will be created or joined.

Given configuration information (query parameters, set in queryconfig), all "LS_Topology"
documents will then be created from existing (seemingly unrelated) data from various 
Arango collections. Relevant data will be collected and organized, corresponding 
"LS_Topology" documents will be created, and finally, the "L3VPN_Topology" documents will be 
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
        print("Gathering LSLinks")
        lsTopologyKeys = getLSTopologyKeys(database)
        uncreatedLSLinkKeys = getDisjointKeys(database, lsTopologyKeys)
        lsLinkKeys = getLSLinkKeys(database)
        for lsLinkKeyIndex in range(len(lsLinkKeys)):
            currentLSLinkKey = lsLinkKeys[lsLinkKeyIndex]
            print("Creating base LS_Topology Documents")
            if currentLSLinkKey in uncreatedLSLinkKeys:
                createBaseLSTopologyDocument(database, currentLSLinkKey)
            else:
                updateBaseLSTopologyDocument(database, currentLSLinkKey)
        print("Done parsing LS-Topology! Next collection begins in 10 seconds.\n")
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
