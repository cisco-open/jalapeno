#! /usr/bin/env python
"""This script collects all label stack paths from a source to a destination.

Given configuration information set in arangoconfig (database connection parameters)
the "Paths" collection will be created or joined.

Given configuration information set in queryconfig (query parameters), label
stacks will then be gathered from Arango path data.
"""

from pyArango.connection import *
from configs import arangoconfig, queryconfig
import logging
from util import connections

def generate_labels():
    """Connect to Arango using parameters in arangoconfig.
    Create label stacks for all paths in network.
    """
    setup_logging()
    logging.info('Creating connection to Arango')
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    logging.info('Enter collection in Arango')
    collection = join_collection(database)
    label_stacks = generate_label_stacks(database, collection)
    return label_stacks

def join_collection(db):
    """Join collection in ArangoDB."""
    database = db
    collection_name = queryconfig.collection  # the collection name is set in queryconfig
    logging.info("Joining " + collection_name + " collection in Arango")
    try:
        collection = database[collection_name]
    except CreationError:
        logging.info(collection_name + " collection does not exist.")
        exit(1)
    return collection

def generate_label_stacks(db, collection):
    """Generate paths in a network using Arango data.
    Insert generated paths into the specified collection.
    """
    database = db
    source = queryconfig.source  # the source for query purposes is vm-t0
    vmsource = queryconfig.vmsource  # the source for client purposes is vm-vm0
    destination_list = open("configs/prefixes.txt").readlines()  # all prefixes should be listed in file
    all_label_stacks = []
    for dest in destination_list:
        destination = dest.rstrip("\n\r")
        logging.info("\nGenerating the label stacks from " + vmsource + " to " + destination)
        label_stacks = generate_label_stacks_query(database, queryconfig.collection, vmsource, destination)
        all_label_stacks.append(label_stacks)
        print("--------------------------------------------------------------------------------")
    return all_label_stacks

def generate_label_stacks_query(db, collection, source, destination):
    """AQL Query to generate label stacks from Arango data."""
    aql = """FOR p in Paths
             FILTER p.Source == @source
	     AND p.Destination == @destination
	     RETURN {Label_Path: p.Label_Path, Key: p._key, Source: p.Source, Destination: p.Destination}
	  """
    bindVars = {'source' : source, 'destination' : destination}
    label_stacks = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return label_stacks

def setup_logging():
    logging.getLogger().setLevel(logging.INFO)

if __name__ == '__main__':
    generate_labels()
    exit(0)
