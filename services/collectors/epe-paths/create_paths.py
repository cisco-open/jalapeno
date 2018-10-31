#! /usr/bin/env python
"""This script creates network path records in a "Paths" collection in ArangoDB.

Given configuration information set in arangoconfig (database connection parameters)
the "Paths" collection will be created or joined.

Given configuration information set in queryconfig (query parameters), paths will then
be calculated from Arango edge data, and inserted as records into the "Paths" collection.

In the future, it may make more sense to create these paths and the Path collection directly
from OpenBMP data using the Topology collector service (during the parsing from data in Kafka
to data in Arango).

This script should be running constantly. However, it should not re-create any
existing paths in the Path collection. It should only upsert.
"""

from pyArango.connection import *
from configs import arangoconfig, queryconfig
from util import utilities, connections
import logging, time

def main():
    setup_logging()
    logging.info('Creating connection to Arango')
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    logging.info('Creating collection in Arango')
    collection = create_collection(database)
    while(True):
        generate_paths(database, collection)


def create_collection(db):
    """Create new collection in ArangoDB.
    If the collection exists, connect to that collection.
    """
    database = db
    collection_name = queryconfig.collection  # the collection name is set in queryconfig
    print("Creating " + collection_name + " collection in Arango")
    try:
        collection = database.createCollection(name=collection_name)
    except CreationError:
        print(collection_name + " collection exists: entering collection.")
        collection = database[collection_name]
    return collection


def generate_paths(db, collection):
    """Generate paths in a network using Arango data.
    Insert generated paths into the specified collection.
    """
    database = db
    destination_list = open("configs/prefixes.txt").readlines()  # all prefixes should be listed in file
    for dest in destination_list:
        destination = dest.rstrip("\n\r")
        #print("\n#############################################################################################################")
        print("Generating all paths to " + destination)
        #print("#############################################################################################################")
        paths = generate_paths_query(database, destination)
        for path in paths:
            create_path_record(collection, path, destination)  # insert path into collection
        clean_paths_collection(database, paths, destination)
    time.sleep(30)


def generate_paths_query(db, destination):
    """AQL Query to generate paths from Arango data."""
    aql = """FOR e in EPEEdges
        FILTER e.Destination == @destination
        RETURN {Source: e.Source, Destination: e.Destination, SRNodeSID: e.SourceSRNodeSID, Interface: e.SourceInterfaceIP, EPELabel: e.EPELabel} """
    bindVars = {'destination': destination}
    paths = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return paths


def clean_paths_collection(db, paths, destination):
    """Remove any paths in the Paths collection that do not exist in reality."""
    print("Removing stale or broken paths to " + destination)
    print("#############################################################################################################")
    real_paths = []
    for path in paths:
        egress_peer = path["Source"]
        epe_label = path["EPELabel"]
        path_destination = path["Destination"]
        key = "EPEPath:" + egress_peer + "_" + epe_label + "_" + path_destination
        real_paths.append(key)

    aql = """FOR p in EPEPaths
        FILTER p.Destination == @destination
        RETURN p._key """
    bindVars = {'destination': destination}
    existing_path_collection = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    for current_path in existing_path_collection:
        if current_path not in real_paths:
            #print("EPEPath " + str(current_path) + " does not exist anymore. Removing from EPEPaths collection.")
            aql = """REMOVE @key IN EPEPaths """
            bindVars = {'key': str(current_path)}
            db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
        #else:
            #print(current_path + " exists, no need to remove")
    #print("#############################################################################################################\n")

def create_path_record(collection, path, destination):
    """Create new path record and insert into collection.
    A sample path record has the following structure:
        key: Path:10.0.0.1_24014_71.71.8.0
        "Egress_Peer": "10.0.0.1", Label_Path": "100003_24013"
        "Interface": "2.2.71.0", "Destination": "71.71.8.0"
    """
    egress_peer = path["Source"]
    sr_node_sid = path["SRNodeSID"]
    egress_interface = path["Interface"]
    epe_label = path["EPELabel"]
    path_destination = path["Destination"]
    label_stack = sr_node_sid + "_" + epe_label
    key = "EPEPath:" + egress_peer + "_" + epe_label + "_" + path_destination
    #print("Creating path from egress-peer " + egress_peer + " through interface " + egress_interface + " to external prefix " + path_destination)
    #print("Path key: " + key)
    #print("Path label stack: " + label_stack)
    
    try:
        document = collection.createDocument()
        document["_key"] = key 
        document["Egress_Peer"] = egress_peer
        document["Egress_Interface"] = egress_interface
        document["Label_Path"] = str(label_stack)
        document["Destination"] = destination
        document.save()
        #print("Successfully created path")
    except CreationError:
        #print("That path already exists!")
        pass
    #print("#############################################################################################################")

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    main()
    exit(0)
