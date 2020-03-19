#! /usr/bin/env python
"""AQL Queries executed by the L3VPN_Topology Service."""

def get_ls_topology_keys(db):
    aql = """ FOR l in LS_Topology return l._key """
    bindVars = {}
    allLSLinks = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return allLSLinks 

def get_disjoint_keys(db, ls_topology_keys):
    aql = """ FOR l in LSLink filter l._key not in @ls_topology_keys return l._key """
    bindVars = {'ls_topology_keys': ls_topology_keys }
    uncreated_ls_link_keys = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return uncreated_ls_link_keys

def get_ls_link_keys(db):
    aql = """ FOR l in LSLink return l._key """
    bindVars = {}
    ls_link_keys = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return ls_link_keys 

def check_exists_ls_topology(db, ls_link_key):
    aql = """ FOR l in LS_Topology filter l._key == @ls_link_key RETURN { key: l._key } """
    bindVars = {'ls_link_key': ls_link_key}
    key_exists = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(key_exists) > 0):
        return True
    else:
        return False

def createBaseLSTopologyDocument(db, ls_link_key):
    aql = """ FOR l in LSLink filter l._key == @ls_link_key insert l into LS_Topology RETURN NEW._key """
    bindVars = {'ls_link_key': ls_link_key}
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        print("Successfully created LS_Topology Edge: " + ls_link_key)
        pass
    else:
        print("Something went wrong while creating LS_Topology Edge")

def updateBaseLSTopologyDocument(db, ls_link_key):
    aql = """ FOR l in LSLink filter l._key == @ls_link_key update l into LS_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_link_key': ls_link_key }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LS_Topology Edge: " + ls_link_key)
        pass
    else:
        print("Something went wrong while updated LS_Topology Edge")

def enhance_ls_topology_document(db, ls_topology_key):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key UPDATE { _key: l._key, "In-Octets": "", "Out-Octets": "",  
    "In-Discards": "", "Out-Discards": "", "Link-Delay": "", "LocalMaxSIDDepth": "", "RemoteMaxSIDDepth": "", "PQResvBW": "", "AppResvBW": "" } in LS_Topology 
    RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_topology_key': ls_topology_key }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully enhanced LS_Topology Edge: " + ls_topology_key)
        pass
    else:
        print("Something went wrong while enhancing LS_Topology Edge")

def get_prefix_sid(db, ls_node_key):
    aql = """ FOR l in LSNode filter l._key == @ls_node_key return l.PrefixSID """
    bindVars = {'ls_node_key': ls_node_key }
    prefix_sid = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return prefix_sid

def get_local_node(db, ls_topology_key):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key return l.LocalRouterID """
    bindVars = {'ls_topology_key': ls_topology_key }
    local_node_id = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return local_node_id

def get_remote_node(db, ls_topology_key):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key return l.RemoteRouterID """
    bindVars = {'ls_topology_key': ls_topology_key }
    remote_node_id = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return remote_node_id

def update_prefix_sid(db, ls_topology_key, local_prefix_sid, remote_prefix_sid):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key UPDATE { _key: l._key, "LocalPrefixSID": @local_prefix_sid, "RemotePrefixSID": @remote_prefix_sid  } in LS_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_topology_key': ls_topology_key, 'local_prefix_sid': local_prefix_sid, 'remote_prefix_sid': remote_prefix_sid }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully enhanced LS_Topology Edge: " + ls_topology_key)
        pass
    else:
        print("Something went wrong while enhancing LS_Topology Edge")

