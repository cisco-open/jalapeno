#! /usr/bin/env python
"""AQL Queries executed by the EPE-Edge Collector.
"""
    
"""Collect peering-routers from existing collections (InternalRouters and PeeringRouters).
Although this could be done using only the PeeringRouters collection, including InternalRouters
allows for an extra level of validation. The list of all peering-routers is returned."""
def get_peering_routers_query(db):
    aql = """ FOR i in InternalRouters 
              FOR p in PeeringRouters 
              FILTER i.RouterIP == p.RouterIP 
              RETURN i.RouterIP """
    bindVars = {}
    peering_routers = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return peering_routers

def get_peering_router_data_query(db, peering_router):
    aql = """ FOR p in PeeringRouters
	      FILTER p.RouterIP == @peering_router_ip
	      RETURN { Source : p.RouterIP, SourceASN : p.ASN } """
    bindVars = {'peering_router_ip': peering_router}
    peering_router_data = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return peering_router_data

def get_peering_router_sr_node_sid(db, peering_router):
    aql = """ FOR p in PeeringRouters
              FOR i in InternalRouters 
              FILTER p.RouterIP == @peering_router_ip AND i.RouterIP == @peering_router_ip AND p.RouterIP == i.RouterIP
	      RETURN i.SRNodeSID """
    bindVars = {'peering_router_ip': peering_router}
    peering_router_sr_node_sid = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return peering_router_sr_node_sid

 
"""Collect external-routers connected to the given peering-router from existing collections (ExternaLinkEdges).
The list of all external-routers is returned."""
def get_external_routers_query(db, peering_router):
    aql = """ FOR e in ExternalLinkEdges
	      FILTER e._from == @external_link_edge_src
	      RETURN e._to """
    bindVars = {'external_link_edge_src': "Routers/"+peering_router}
    external_routers = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)  
    return external_routers

def get_external_link_edge_data_query(db, peering_router, external_router):
    aql = """ FOR e in ExternalLinkEdges
	      FILTER e._from == @external_link_edge_src AND e._to == @external_link_edge_dst
	      RETURN { Source: e._from, SrcInterfaceIP: e.SrcInterfaceIP, Label: e.Label, DstInterfaceIP: e.DstInterfaceIP, Destination: e._to } """
    bindVars = {'external_link_edge_src': "Routers/"+peering_router, 'external_link_edge_dst': external_router}
    external_link_edge_data = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)  
    return external_link_edge_data


"""Collect external-prefixes connected to the given external-router from existing collections (ExternalPrefixEdges).
The list of all external-prefix is returned."""
def get_external_prefixes_query(db, interface_ip, external_router):
    aql = """ FOR e in ExternalPrefixEdges
	      FILTER e._from == @external_prefix_edge_src AND e.SrcIntfIP == @external_prefix_edge_src_intf_ip 
	      RETURN e._to """
    bindVars = {'external_prefix_edge_src': external_router, 'external_prefix_edge_src_intf_ip': interface_ip}
    external_prefixes = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)  
    return external_prefixes

def get_external_prefix_edge_data_query(db, interface_ip, external_router, external_prefix):
    aql = """ FOR e in ExternalPrefixEdges
	      FILTER e.SrcIntfIP == @external_prefix_edge_intf_ip AND e._from == @external_prefix_edge_src AND e._to == @external_prefix_edge_dst
	      RETURN { Source: e._from, SrcInterfaceIP: e.SrcIntfIP, SrcRouterASN: e.SrcRouterASN, Destination: e._to, DstPrefix: e.DstPrefix, DstPrefixASN: e.DstPrefixASN } """
    bindVars = {'external_prefix_edge_intf_ip': interface_ip, 'external_prefix_edge_src': external_router, 'external_prefix_edge_dst': external_prefix}
    external_prefix_edge_data = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)  
    return external_prefix_edge_data


def get_epe_edge_key(db, key):
    aql = """ FOR e in EPEEdges
	      FILTER e._key == @epe_edge_key
	      RETURN e._key """
    bindVars = {'epe_edge_key': key}
    epe_edge_key = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)  
    return epe_edge_key
    
"""Given a list of peering routers, collect data from two different collections (ExternalLinkEdge and InternalRouter) for each
peering router. This allows for the marriage of data in ExternalLinkEdge (interface_ips, etc.) and in InternalRouter (sr_node_sid).
The data is compiled into an "epe_edge_data" record. The record is then added to a list of all epe_edge_data records. 
Finally, the list of all records is returned."""
def get_epe_edges_data_query(db, peering_routers):
    all_epe_edge_data = []
    for peering_router in peering_routers:
        aql = """ FOR e in ExternalLinkEdges
                   FOR i in InternalRouters
                   FILTER e._from == @external_edge_src_router_ip AND i.RouterIP == @src_router_ip
                   RETURN { src : e._from, dst : e._to, src_interface_ip : e.SrcInterfaceIP, dst_interface_ip: e.DstInterfaceIP, label: e.Label, src_router_ip : i.RouterIP, src_asn : i.ASN, src_igp: i.IGPID, src_name : i.Name, src_sr_node_sid : i.SRNodeSID } """
        bindVars = {'external_edge_src_router_ip': 'Routers/'+peering_router, 'src_router_ip': peering_router}
        peering_router_epe_edge_data = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
        for epe_edge_data in peering_router_epe_edge_data:
            all_epe_edge_data.append(epe_edge_data)
    return all_epe_edge_data



def update_epe_edge_query(db, epe_edge_key, source, source_asn, source_sr_node_sid, source_intf_ip, source_epe_label, hop_intf_ip, hop, hop_asn, destination, destination_asn):
    epe_edge_from = 'Routers/'+source
    epe_edge_to = 'Prefixes/'+destination
    aql = """ FOR e in EPEEdges
	      FILTER e._key == @epe_edge_key
	      UPDATE { _key: e._key, _to: @epe_edge_to, _from: @epe_edge_from, Source: @epe_edge_src, SourceASN: @epe_edge_src_asn, 
                       SourceSRNodeSID: @epe_edge_src_sr_node_sid, SourceInterfaceIP: @epe_edge_src_intf_ip, EPELabel: @epe_edge_src_epe_label, 
                       HopInterfaceIP: @epe_edge_hop_intf_ip, Hop: @epe_edge_hop, HopASN: @epe_edge_hop_asn, Destination: @epe_edge_dst, DestinationASN: @epe_edge_dst_asn }
              IN EPEEdges RETURN { before: OLD, after: NEW } """
    bindVars = {'epe_edge_key': epe_edge_key, 'epe_edge_from': epe_edge_from, 'epe_edge_to': epe_edge_to, 'epe_edge_src': source, 
                'epe_edge_src_asn': source_asn, 'epe_edge_src_sr_node_sid': source_sr_node_sid, 
		'epe_edge_src_intf_ip': source_intf_ip, 'epe_edge_src_epe_label': source_epe_label, 'epe_edge_hop_intf_ip': hop_intf_ip,
		'epe_edge_hop': hop, 'epe_edge_hop_asn': hop_asn, 'epe_edge_dst': destination, 'epe_edge_dst_asn': destination_asn}
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        pass
	#print("Successfully updated EPEEdge: " + epe_edge_key)
    else:
	print("Something went wrong while updating EPEEdge")


def create_epe_edge_query(db, epe_edge_key, source, source_asn, source_sr_node_sid, source_intf_ip, source_epe_label, hop_intf_ip, hop, hop_asn, destination, destination_asn):
    epe_edge_from = 'Routers/'+source
    epe_edge_to = 'Prefixes/'+str(destination)
    aql = """ INSERT { _key: @epe_edge_key, _to: @epe_edge_to, _from: @epe_edge_from, Source: @epe_edge_src, SourceASN: @epe_edge_src_asn, 
                       SourceSRNodeSID: @epe_edge_src_sr_node_sid, SourceInterfaceIP: @epe_edge_src_intf_ip, EPELabel: @epe_edge_src_epe_label, 
                       HopInterfaceIP: @epe_edge_hop_intf_ip, Hop: @epe_edge_hop, HopASN: @epe_edge_hop_asn, Destination: @epe_edge_dst, DestinationASN: @epe_edge_dst_asn }
              INTO EPEEdges RETURN NEW._key """
    bindVars = {'epe_edge_key': epe_edge_key, 'epe_edge_from': epe_edge_from, 'epe_edge_to': epe_edge_to, 'epe_edge_src': source, 
                'epe_edge_src_asn': source_asn, 'epe_edge_src_sr_node_sid': source_sr_node_sid, 
		'epe_edge_src_intf_ip': source_intf_ip, 'epe_edge_src_epe_label': source_epe_label, 'epe_edge_hop_intf_ip': hop_intf_ip,
		'epe_edge_hop': hop, 'epe_edge_hop_asn': hop_asn, 'epe_edge_dst': destination, 'epe_edge_dst_asn': destination_asn}
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        pass
	#print("Successfully created EPEEdge: " + epe_edge_key)
    else:
	print("Something went wrong while creating EPEEdge")

