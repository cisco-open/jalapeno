#! /usr/bin/env python
"""This script creates network epe-edge records in an "EPEEdges" collection in ArangoDB.

Given configuration information (database connection parameters, set in arangoconfig)
the "EPEEdges" collection will be created or joined.

Given configuration information (query parameters, set in queryconfig), all "EPEEdge"
documents will then be created from existing (seemingly unrelated) data from various 
Arango collections. Relevant data will be collected and organized, corresponding 
"EPEEdge" documents will be created, and finally, the "EPEEdge" documents will be 
upserted into the "EPEEdges" collection.
"""

from pyArango.connection import *
from configs import arangoconfig, queryconfig
from util import connections
import logging, time, json, sys
from epe_edge_queries import *

def main():
    setup_logging()
    logging.info('Creating connection to Arango')
    connection = connections.ArangoConn()
    database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    logging.info('Creating collection in Arango')
    collection = create_collection(database, queryconfig.collection)  # the collection name is set in queryconfig
    while(True):
        # parse existing collections for relevant fields that correlate to a potential EPEEdge
        epe_edges_data = collect_epe_edges_data(database, collection)
        # for each record of epe-edge data, create & upsert a corresponding EPEEdge document into the EPEEdges collection
        create_epe_edges(database, collection, epe_edges_data)
        print("Done parsing EPEEdges! Next collection begins in 10 seconds.\n") 
        time.sleep(10)
 
def create_collection(db, collection_name):
    """Create new collection in ArangoDB.
    If the collection exists, connect to that collection.
    """
    database = db
    print("Creating " + collection_name + " collection in Arango")
    try:
        collection = database.createCollection(name=collection_name)
    except CreationError:
        print(collection_name + " collection exists: entering collection.")
        collection = database[collection_name]
    return collection


def collect_epe_edges_data(db, collection):
    """ Collect EPEEdge data from existing collections. 
    Return records of epe_edge_data in a list.
    """
    all_epe_edge_data = []

    # collecting list of BorderRouters
    border_routers = get_border_routers_query(db)
    for border_router_index in range(len(border_routers)):
        border_router = border_routers[border_router_index]
        # print("Parsing EPEEdge for BorderRouter: " + border_router)
        border_router_data = get_border_router_data_query(db, border_router)
        border_router_source = border_router_data[0]['Source']
        border_router_asn = border_router_data[0]['SourceASN']
        border_router_sr_node_sid = get_border_router_sr_node_sid(db, border_router)[0]
        # print("Current BorderRouter data: " + border_router_source + " " + border_router_asn + " " + border_router_sr_node_sid)

        # collecting ExternalRouters connected to current BorderRouter
        external_routers = get_external_routers_query(db, border_router)
        for external_router_index in range(len(external_routers)):
            external_router = external_routers[external_router_index]
            # print("Parsing EPEEdge for Border Router: " + border_router + " connected to External Router: " + external_router)
            external_router_hop = external_router
            external_link_edge_data = get_external_link_edge_data_query(db, border_router, external_router)
            external_link_edge_source = external_link_edge_data[0]['Source']
            external_link_edge_destination = external_link_edge_data[0]['Destination']
            external_link_edge_src_intf_ip = external_link_edge_data[0]['SrcInterfaceIP']
            external_link_edge_dst_intf_ip = external_link_edge_data[0]['DstInterfaceIP']
            external_link_edge_label = external_link_edge_data[0]['Label']

            external_prefixes = get_external_prefixes_query(db, external_link_edge_dst_intf_ip, external_router)
            for external_prefix_index in range(len(external_prefixes)):
                external_prefix = external_prefixes[external_prefix_index]
                print("Parsing EPEEdge for border_router " + border_router + " for external_router " + external_router + " with external_prefix " + external_prefix)
                external_prefix_edge_data = get_external_prefix_edge_data_query(db, external_link_edge_dst_intf_ip, external_router, external_prefix)
                external_prefix_edge_src_asn = external_prefix_edge_data[0]['SrcRouterASN']
                external_prefix_edge_src_intf_ip = external_prefix_edge_data[0]['SrcInterfaceIP']
                external_prefix_edge_dst_prefix_asn = external_prefix_edge_data[0]['DstPrefixASN']
                external_prefix_edge_destination = external_prefix_edge_data[0]['Destination']
                external_prefix_edge_dst_prefix = external_prefix_edge_data[0]['DstPrefix']
                external_prefix_edge_source = external_prefix_edge_data[0]['Source']

                if((('Routers/'+border_router) == external_link_edge_source) and external_link_edge_destination == external_prefix_edge_source):
                    epe_edge_data = dict(epe_edge_src=external_link_edge_source, epe_edge_src_asn=border_router_asn,
                                    epe_edge_src_sr_node_sid=border_router_sr_node_sid, epe_edge_src_intf_ip=external_link_edge_src_intf_ip,
                                    epe_edge_src_epe_label=external_link_edge_label, epe_edge_hop_intf_ip=external_link_edge_dst_intf_ip,
   				    epe_edge_hop=external_prefix_edge_source, epe_edge_hop_asn=external_prefix_edge_src_asn, 
			 	    epe_edge_dst=external_prefix_edge_destination, epe_edge_dst_asn=external_prefix_edge_dst_prefix_asn)
                    '''print("Parsed EPEEdge: \n Source: %s %s %s %s %s \n Hop: %s %s %s \n Destination: %s %s" % (external_link_edge_source, 
			border_router_asn, border_router_sr_node_sid, external_link_edge_src_intf_ip, external_link_edge_label, external_link_edge_dst_intf_ip, 
			external_prefix_edge_source, external_prefix_edge_src_asn, external_prefix_edge_destination, external_prefix_edge_dst_prefix_asn))'''
                    all_epe_edge_data.append(epe_edge_data)
                    print("======================Parsed!===========================\n")
    print("========================================================================\n")
    return all_epe_edge_data


def create_epe_edges(db, collection, all_epe_edge_data):
    """Create epe-edges in a network using Arango data.
    Insert generated epe-edges into the specified collection.
    """
    for epe_edge_data in all_epe_edge_data:
        all_fields_exist = check_epe_fields(epe_edge_data)
        if(all_fields_exist):
            source = str(epe_edge_data['epe_edge_src'].replace('Routers/', ''))
            source_asn = str(epe_edge_data['epe_edge_src_asn'])
            source_epe_label = str(epe_edge_data['epe_edge_src_epe_label'])
            source_sr_node_sid = str(epe_edge_data['epe_edge_src_sr_node_sid'])
            source_intf_ip = str(epe_edge_data['epe_edge_src_intf_ip'])
            hop_intf_ip = str(epe_edge_data['epe_edge_hop_intf_ip'])
            hop = str(epe_edge_data['epe_edge_hop'].replace('Routers/', ''))
            hop_asn = str(epe_edge_data['epe_edge_hop_asn'])
            destination = str(epe_edge_data['epe_edge_dst'].replace('Prefixes/', ''))
            destination_asn = str(epe_edge_data['epe_edge_dst_asn'])
            epe_edge_key = str(source + "_" + epe_edge_data['epe_edge_src_intf_ip'] + "_" + epe_edge_data['epe_edge_hop_intf_ip'] + "_" + hop + "_" + destination)

            existing_epe_edge = get_epe_edge_key(db, epe_edge_key)
            if len(existing_epe_edge) > 0:
                update_epe_edge_query(db, epe_edge_key, source, source_asn, source_sr_node_sid, source_intf_ip, source_epe_label, hop_intf_ip, hop, hop_asn, 
                                      destination, destination_asn)
            else:
                create_epe_edge_query(db, epe_edge_key, source, source_asn, source_sr_node_sid, source_intf_ip, source_epe_label, hop_intf_ip, hop, hop_asn, 
                                      destination, destination_asn)


def check_epe_fields(epe_edge_data):
    all_fields_exist = False
    if (('epe_edge_src' in epe_edge_data) and (epe_edge_data.get('epe_edge_src') != "") and
       ('epe_edge_src_asn' in epe_edge_data) and (epe_edge_data.get('epe_edge_src_asn') != "") and
       ('epe_edge_src_sr_node_sid' in epe_edge_data) and (epe_edge_data.get('epe_edge_src_sr_node_sid') != "") and
       ('epe_edge_src_intf_ip' in epe_edge_data) and (epe_edge_data.get('epe_edge_src_intf_ip') != "") and
       ('epe_edge_src_epe_label' in epe_edge_data) and (epe_edge_data.get('epe_edge_src_epe_label') != "") and
       ('epe_edge_hop_intf_ip' in epe_edge_data) and (epe_edge_data.get('epe_edge_hop_intf_ip') != "") and
       ('epe_edge_hop' in epe_edge_data) and (epe_edge_data.get('epe_edge_hop') != "") and
       ('epe_edge_hop_asn' in epe_edge_data) and (epe_edge_data.get('epe_edge_hop_asn') != "") and
       ('epe_edge_dst' in epe_edge_data) and (epe_edge_data.get('epe_edge_dst') != "") and
       ('epe_edge_dst_asn' in epe_edge_data) and (epe_edge_data.get('epe_edge_dst_asn') != "")):
        all_fields_exist = True
    return all_fields_exist

     
def setup_logging():
    logging.getLogger().setLevel(logging.WARNING)
    logging.getLogger("requests").setLevel(logging.WARNING)
    logging.getLogger("urllib3").setLevel(logging.WARNING)

if __name__ == '__main__':
    main()
    exit(0)
