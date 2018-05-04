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
    source = queryconfig.source  # the source for query purposes is vm-t0
    vmsource = queryconfig.vmsource  # the source for client purposes is vm-vm0
    destination_list = open("configs/prefixes.txt").readlines()  # all prefixes should be listed in file
    for dest in destination_list:
        destination = dest.rstrip("\n\r")
        print("\nGenerating all paths from " + vmsource + " to " + destination)
        paths = generate_paths_query(database, source, destination)
        for path in paths:
            create_path_record(collection, path, vmsource, destination)  # insert path into collection
        clean_paths_collection(database, paths, vmsource, destination)
    time.sleep(10)

def generate_paths_query(db, source, destination):
    """AQL Query to generate paths from Arango data."""
    aql = """ FOR v,e,p in 4
        OUTBOUND @source
        LinkEdgesV4, PrefixEdges
        OPTIONS {bfs: False, uniqueEdges: "path", uniqueVertices: "path"}
        FILTER p.vertices[-1]._id == @destination
        RETURN [p.edges[* FILTER CURRENT.Label != null]._from, p.edges[* FILTER CURRENT.Label != null]._to,
            p.edges[* FILTER CURRENT.Label != null].FromIP, p.edges[* FILTER CURRENT.Label != null].Label] """
    bindVars = {'source': 'Routers/'+source, 'destination': 'Prefixes/'+destination}
    paths = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return paths

def clean_paths_collection(db, paths, source, destination):
    """Remove any paths in the Paths collection that do not exist in reality."""
    real_paths = []
    for path in paths:
	route = ', '.join(utilities.uniqify(path[0] + path[1])).replace('Routers/', '')  # formatting route and removing duplicates
        key_route = route.replace(', ', "_to_")  # formatting document key
	real_paths.append("Path:" + source + "_to_" + key_route + "_to_" + destination)
    aql = """FOR p in Paths
	FILTER p.Destination == @destination
        RETURN p._key """
    bindVars = {'destination': destination}
    existing_path_collection = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    for current_path in existing_path_collection:
	if current_path not in real_paths:
	    print("Path " + str(current_path) + " does not exist anymore. Removing from Paths collection.")
	    aql = """REMOVE @key IN Paths """
            bindVars = {'key': str(current_path)}
            db.AQLQuery(aql, rawResults=True, bindVars=bindVars)


def create_path_record(collection, path, source, destination):
    """Create new path record and insert into collection.
    A sample path record has the following structure:
        key: Path:10.1.2.1_to_10.1.1.0_to_10.1.1.2_to_10.1.1.4_to_10.100.100.8_to_10.13.0.0_24
        "Latency": 0, "Label_Path": "24006_24007_24007", "Source": "10.1.2.1",
        "Path": "10.1.1.0, 10.1.1.2, 10.1.1.4, 10.100.100.8", "Destination": "10.13.0.0_24",
        "Interface_Path": "2.2.2.4, 2.2.2.12, 2.2.2.24"
    """
    route = ', '.join(utilities.uniqify(path[0] + path[1])).replace('Routers/', '')  # formatting route and removing duplicates
    key_route = route.replace(', ', "_to_")  # formatting document key
    interfaces = ', '.join(path[2])  # formatting path interfaces
    label_stack = ', '.join(path[3]).replace(', ', "_")  # formatting label stack
    print("Creating path: " + source + ", " + route + ", " + destination)

    '''big question here. is it reasonable to say that from VM0 to C9, a path (t0, l1, p3, ep5) would be a singular path?
       is there a chance that one path could have two sets of interfaces, labels, latencies?
       for example, instead of interfaces 2.2.2.2, 2.2.2.2, 2.2.2.24, we get u'2.2.2.2', u'10.10.10.6', u'2.2.2.16
       would the interface matter?
    '''
    try:
        document = collection.createDocument()
        document["_key"] = "Path:" + source + "_to_" + key_route + "_to_" + destination
        document["Source"] = source
        document["Destination"] = destination
        document["Path"] = str(route)
        document["Interface_Path"] = str(interfaces)
        document["Label_Path"] = str(label_stack)
        document["Latency"] = 0
        document.save()
    except CreationError:
        print("That path already exists!")
	pass

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    main()
    exit(0)
