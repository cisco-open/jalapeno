#! /usr/bin/env python
"""AQL Queries executed by the LS_Topology Service."""

def getLSLinkKeys(db):
    aql = """ FOR l in LSLink return l._key """
    bindVars = {}
    allLSLinks = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return allLSLinks

def getLSTopologyKeys(db):
    aql = """ FOR l in LS_Topology return l._key """
    bindVars = {}
    allLSLinks = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return allLSLinks  

def getDisjointKeys(db, lsTopologyKeys):
    aql = """ FOR l in LSLink filter l._key not in @lsTopologyKeys return l._key """
    bindVars = {'lsTopologyKeys': lsTopologyKeys }
    uncreatedLSLinkKeys = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return uncreatedLSLinkKeys

def createBaseLSTopologyDocument(db, lsLinkKey):
    aql = """ FOR l in LSLink filter l._key == @lsLinkKey insert l into LS_Topology RETURN NEW._key """
    bindVars = {'lsLinkKey': lsLinkKey }
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        print("Successfully created LS_Topology Edge: " + lsLinkKey)
        pass
    else:
        print("Something went wrong while creating LS_Topology Edge")

def updateBaseLSTopologyDocument(db, lsLinkKey):
    aql = """ FOR l in LSLink filter l._key == @lsLinkKey update l into LS_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'lsLinkKey': lsLinkKey }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LS_Topology Edge: " + lsLinkKey)
        pass
    else:
        print("Something went wrong while updated LS_Topology Edge")

# def update_node_to_prefix_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt):
#     l3vpn_topology_edge_from = 'L3VPNode/'+str(router_id)
#     l3vpn_topology_edge_to = 'L3VPNPrefix/'+str(prefix)
#     aql = """ FOR e in L3VPN_Topology
# 	      FILTER e._key == @l3vpn_topology_edge_key
# 	      UPDATE { 
#                   _key: @l3vpn_topology_edge_key,
#                   _to: @l3vpn_topology_edge_to,
#                   _from: @l3vpn_topology_edge_from,
#                   SrcIP: @router_id,
#                   DstIP: @prefix,
#                   VPN_Prefix: @prefix,
#                   VPN_Prefix_Len: @prefix_length,
#                   RouterID: @router_id,
#                   PrefixSID: @prefix_sid,
#                   VPN_Label: @vpn_label,
#                   RD: @rd,
#                   RT: @rt,
#                   Source: @router_id,
#                   Destination: @prefix }
#               IN L3VPN_Topology RETURN { before: OLD, after: NEW } """
#     bindVars = {'l3vpn_topology_edge_key': l3vpn_topology_edge_key, 'l3vpn_topology_edge_from': l3vpn_topology_edge_from,
#                 'l3vpn_topology_edge_to': l3vpn_topology_edge_to, 'prefix': prefix, 'prefix_length': prefix_length,
#                 'router_id': router_id, 'prefix_sid': prefix_sid, 'vpn_label': vpn_label, 'rd': rd, 'rt': rt}
#     updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
#     if(len(updated_edge) > 0):
#         print("Successfully updated L3VPN_Topology Edge: " + l3vpn_topology_edge_key)
#         pass
#     else:
#         print("Something went wrong while updating L3VPN_Topology Edge")
