#! /usr/bin/env python
"""AQL Queries executed by the LS_Topology Service."""

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
        #print("Successfully created LS_Topology Edge: " + ls_link_key)
        pass
    else:
        print("Something went wrong while creating LS_Topology Edge")

def updateBaseLSTopologyDocument(db, ls_link_key):
    aql = """ FOR l in LSLink filter l._key == @ls_link_key update l into LS_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_link_key': ls_link_key }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        #print("Successfully updated LS_Topology Edge: " + ls_link_key)
        pass
    else:
        print("Something went wrong while updated LS_Topology Edge")

def enhance_ls_topology_document(db, ls_topology_key):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key UPDATE { _key: l._key, "Link-Delay": "", "LocalMaxSIDDepth": "",
    "RemoteMaxSIDDepth": "", "PQResvBW": "", "AppResvBW": "" } in LS_Topology
    RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_topology_key': ls_topology_key }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        #print("Successfully enhanced LS_Topology Edge: " + ls_topology_key)
        pass
    else:
        print("Something went wrong while enhancing LS_Topology Edge")

def get_srgb_start(db, ls_node_key):
    aql = """ FOR l in LSNode filter l._key == @ls_node_key return l.SRGBStart """
    bindVars = {'ls_node_key': ls_node_key }
    srgb_start = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return srgb_start

def get_max_sid_depth(db, igp_router_id):
    aql = """ FOR l in LSNode filter l._key == @igp_router_id return l.NodeMaxSIDDepth """
    bindVars = {'igp_router_id': igp_router_id }
    max_sid_depth = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return max_sid_depth

def get_prefix_info(db, igp_router_id):
    aql = """ FOR l in LSPrefix filter l.IGPRouterID == @igp_router_id return {"Prefix": l.Prefix, "Length": l.Length, "SIDIndex": l.SIDIndex, "SRFlag": l.SRFlags } """
    bindVars = {'igp_router_id': igp_router_id }
    prefix_info = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return prefix_info

def get_local_igpid(db, ls_topology_key):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key return {"LocalIGPID": l.LocalIGPID} """
    bindVars = {'ls_topology_key': ls_topology_key }
    local_igpid = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return local_igpid

def get_remote_igpid(db, ls_topology_key):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key return {"RemoteIGPID": l.RemoteIGPID} """
    bindVars = {'ls_topology_key': ls_topology_key }
    remote_igpid = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return remote_igpid


def update_prefix_sid(db, ls_topology_key, local_prefix_sid, remote_prefix_sid):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key UPDATE { _key: l._key, "LocalPrefixSID": @local_prefix_sid, "RemotePrefixSID": @remote_prefix_sid  } in LS_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_topology_key': ls_topology_key, 'local_prefix_sid': local_prefix_sid, 'remote_prefix_sid': remote_prefix_sid }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LS_Topology Edge: " + ls_topology_key + " with LocalPrefixSID " + str(local_prefix_sid) + " and RemotePrefixSID " + str(remote_prefix_sid))
        pass
    else:
        print("Something went wrong while updating LS_Topology Edge with PrefixSIDs")

def update_prefix_info(db, ls_topology_key, local_prefix_info, remote_prefix_info):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key UPDATE { _key: l._key, "LocalPrefixInfo": @local_prefix_info, "RemotePrefixInfo": @remote_prefix_info  } in LS_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_topology_key': ls_topology_key, 'local_prefix_info': local_prefix_info, 'remote_prefix_info': remote_prefix_info }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LS_Topology Edge: " + ls_topology_key + " with LocalPrefixInfo " + str(local_prefix_info) + " and RemotePrefixInfo " + str(remote_prefix_info))
        pass
    else:
        print("Something went wrong while updating LS_Topology Edge with PrefixInfo")

def update_max_sid_depths(db, ls_topology_key, local_max_sid_depth, remote_max_sid_depth):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key UPDATE { _key: l._key, "LocalMaxSIDDepth": @local_max_sid_depth, "RemoteMaxSIDDepth": @remote_max_sid_depth  } in LS_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_topology_key': ls_topology_key, 'local_max_sid_depth': local_max_sid_depth, 'remote_max_sid_depth': remote_max_sid_depth }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LS_Topology Edge: " + ls_topology_key + " with LocalMaxSIDDepth " + str(local_max_sid_depth) + " and RemoteMaxSIDDepth " + str(remote_max_sid_depth))
        pass
    else:
        print("Something went wrong while updating LS_Topology Edge with MaxSIDDepths")

def update_ls_topology_document(db, ls_topology_key, local_prefix_sid, remote_prefix_sid, local_prefix_info, remote_prefix_info, local_max_sid_depth, remote_max_sid_depth):
    aql = """ FOR l in LS_Topology filter l._key == @ls_topology_key UPDATE { _key: l._key, "LocalPrefixSID": @local_prefix_sid, "RemotePrefixSID": @remote_prefix_sid,
              "LocalPrefixInfo": @local_prefix_info, "RemotePrefixInfo": @remote_prefix_info, "LocalMaxSIDDepth": @local_max_sid_depth, "RemoteMaxSIDDepth": @remote_max_sid_depth 
              } in LS_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_topology_key': ls_topology_key, 'local_prefix_sid': local_prefix_sid, 'remote_prefix_sid': remote_prefix_sid, 'local_prefix_info': local_prefix_info, 'remote_prefix_info': remote_prefix_info,
                'local_max_sid_depth': local_max_sid_depth, 'remote_max_sid_depth': remote_max_sid_depth }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LS_Topology Edge: " + ls_topology_key + " with LocalPrefixSID " + str(local_prefix_sid) + " and RemotePrefixSID " + str(remote_prefix_sid))
        pass
    else:
        print("Something went wrong while updating LS_Topology Edge: " + ls_topology_key)

