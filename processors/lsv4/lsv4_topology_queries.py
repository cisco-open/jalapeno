#! /usr/bin/env python
"""AQL Queries executed by the LSv4_Topology Service."""

def get_lsv4_topology_keys(db):
    aql = """ FOR l in LSv4_Topology return l._key """
    bindVars = {}
    allLSLinks = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return allLSLinks

def get_disjoint_keys(db, ls_topology_keys):
    aql = """ FOR l in LSLinkEdge filter l._key not in @lsv4_topology_keys return l._key """
    bindVars = {'lsv4_topology_keys': ls_topology_keys }
    uncreated_ls_link_keys = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return uncreated_ls_link_keys

def get_lsv4_link_keys(db):
    aql = """ FOR l in LSLinkEdge filter NOT CONTAINS(l.local_interface_ip, ":") and NOT CONTAINS(l.remote_interface_ip, ":") return l._key """
    bindVars = {}
    ls_link_keys = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return ls_link_keys

def check_exists_lsv4_topology(db, ls_link_key):
    aql = """ FOR l in LSv4_Topology filter l._key == @ls_link_key RETURN { key: l._key } """
    bindVars = {'ls_link_key': ls_link_key}
    key_exists = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(key_exists) > 0):
        return True
    else:
        return False

def check_exists_lsv4_link(db, lsv4_topology_key):
    aql = """ FOR l in LSLinkEdge filter l._key == @lsv4_topology_key RETURN { key: l._key } """
    bindVars = {'lsv4_topology_key': lsv4_topology_key}
    key_exists = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(key_exists) > 0):
        return True
    else:
        return False

def deleteLSv4TopologyDocument(db, lsv4_topology_key):
    aql = """ FOR l in LSv4_Topology filter l._key == @lsv4_topology_key remove l in LSv4_Topology """
    bindVars = {'lsv4_topology_key': lsv4_topology_key}
    deleted_document = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)

def createBaseLSv4TopologyDocument(db, ls_link_key):
    aql = """ FOR l in LSLinkEdge filter l._key == @ls_link_key insert l into LSv4_Topology RETURN NEW._key """
    bindVars = {'ls_link_key': ls_link_key}
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        #print("Successfully created LSv4_Topology Edge: " + ls_link_key)
        pass
    else:
        print("Something went wrong while creating LSv4_Topology Edge")

def updateBaseLSv4TopologyDocument(db, ls_link_key):
    aql = """ FOR l in LSLinkEdge filter l._key == @ls_link_key update l into LSv4_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_link_key': ls_link_key }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        #print("Successfully updated LSv4_Topology Edge: " + ls_link_key)
        pass
    else:
        print("Something went wrong while updated LSv4_Topology Edge")

def enhance_lsv4_topology_document(db, lsv4_topology_key):
    aql = """ FOR l in LSv4_Topology filter l._key == @lsv4_topology_key UPDATE { _key: l._key, "link_delay": "", "local_msd": "",
    "remote_msd": "", "pq_resv_bw": "", "app_resv_bw": "" } in LSv4_Topology
    RETURN { before: OLD, after: NEW }"""
    bindVars = {'lsv4_topology_key': lsv4_topology_key }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        #print("Successfully enhanced LSv4_Topology Edge: " + lsv4_topology_key)
        pass
    else:
        print("Something went wrong while enhancing LSv4_Topology Edge")

def get_srgb_start(db, ls_node_key):
    aql = """ FOR l in LSNode filter l._key == @ls_node_key return l.srgb_start """
    bindVars = {'ls_node_key': ls_node_key }
    srgb_start = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return srgb_start

def get_max_sid_depth(db, igp_router_id):
    aql = """ FOR l in LSNode filter l._key == @igp_router_id return l.node_msd """
    bindVars = {'igp_router_id': igp_router_id }
    max_sid_depth = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return max_sid_depth

def get_prefix_info(db, igp_router_id):
    aql = """ FOR l in LSPrefix filter l.igp_router_id == @igp_router_id and l.prefix_sid != null for z in l.prefix_sid filter z.algo == 0 return {"prefix": l.prefix, "length": l.length, "flags": z.flags, "sid_index":  z.prefix_sid}"""
    bindVars = {'igp_router_id': igp_router_id }
    prefix_info = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return prefix_info

def get_local_igpid(db, lsv4_topology_key):
    aql = """ FOR l in LSv4_Topology filter l._key == @lsv4_topology_key return {"local_igp_id": l.local_igp_id} """
    bindVars = {'lsv4_topology_key': lsv4_topology_key }
    local_igpid = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return local_igpid

def get_remote_igpid(db, lsv4_topology_key):
    aql = """ FOR l in LSv4_Topology filter l._key == @lsv4_topology_key return {"remote_igp_id": l.remote_igp_id} """
    bindVars = {'lsv4_topology_key': lsv4_topology_key }
    remote_igpid = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return remote_igpid

def get_sr_flags(db, igp_id):
    aql = """ FOR l in LSPrefix filter l.SRFlags != null and l.IGPRouterID == @igp_id return l.SRFlags[0] """
    bindVars = {'igp_id': igp_id}
    sr_flags = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return sr_flags

def update_prefix_sid(db, lsv4_topology_key, local_prefix_sid, remote_prefix_sid):
    aql = """ FOR l in LSv4_Topology filter l._key == @lsv4_topology_key UPDATE { _key: l._key, "local_prefix_sid": @local_prefix_sid, "remote_prefix_sid": @remote_prefix_sid  } in LSv4_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'lsv4_topology_key': lsv4_topology_key, 'local_prefix_sid': local_prefix_sid, 'remote_prefix_sid': remote_prefix_sid }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LSv4_Topology Edge: " + lsv4_topology_key + " with local_prefix_sid " + str(local_prefix_sid) + " and remote_prefix_sid " + str(remote_prefix_sid))
        pass
    else:
        print("Something went wrong while updating LSv4_Topology Edge with PrefixSIDs")

def update_prefix_info(db, lsv4_topology_key, local_prefix_info, remote_prefix_info):
    aql = """ FOR l in LSv4_Topology filter l._key == @lsv4_topology_key UPDATE { _key: l._key, "local_prefix_info": @local_prefix_info, "remote_prefix_info": @remote_prefix_info  } in LSv4_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'lsv4_topology_key': lsv4_topology_key, 'local_prefix_info': local_prefix_info, 'remote_prefix_info': remote_prefix_info }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LSv4_Topology Edge: " + lsv4_topology_key + " with LocalPrefixInfo " + str(local_prefix_info) + " and RemotePrefixInfo " + str(remote_prefix_info))
        pass
    else:
        print("Something went wrong while updating LSv4_Topology Edge with PrefixInfo")

def update_max_sid_depths(db, lsv4_topology_key, local_max_sid_depth, remote_max_sid_depth):
    aql = """ FOR l in LSv4_Topology filter l._key == @lsv4_topology_key UPDATE { _key: l._key, "local_msd": @local_msd, "remote_msd": @remote_msd  } in LSv4_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'lsv4_topology_key': lsv4_topology_key, 'local_msd': local_msd, 'remote_msd': remote_msd }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LSv4_Topology Edge: " + lsv4_topology_key + " with LocalMaxSIDDepth " + str(local_max_sid_depth) + " and RemoteMaxSIDDepth " + str(remote_max_sid_depth))
        pass
    else:
        print("Something went wrong while updating LSv4_Topology Edge with MaxSIDDepths")

def update_lsv4_topology_document(db, lsv4_topology_key, local_prefix_sid, remote_prefix_sid, local_prefix_info, remote_prefix_info, local_msd, remote_msd):
    aql = """ FOR l in LSv4_Topology filter l._key == @lsv4_topology_key UPDATE { _key: l._key, "local_prefix_sid": @local_prefix_sid, "remote_prefix_sid": @remote_prefix_sid,
              "LocalPrefixInfo": @local_prefix_info, "remote_prefix_info": @remote_prefix_info, "local_msd": @local_msd, "remote_msd": @remote_msd 
              } in LSv4_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'lsv4_topology_key': lsv4_topology_key, 'local_prefix_sid': local_prefix_sid, 'remote_prefix_sid': remote_prefix_sid, 'local_prefix_info': local_prefix_info, 'remote_prefix_info': remote_prefix_info, 'local_msd': local_msd, 'remote_msd': remote_msd }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LSv4_Topology Edge: " + lsv4_topology_key + " with local_prefix_sid " + str(local_prefix_sid) + " and remote_prefix_sid " + str(remote_prefix_sid))
        pass
    else:
        print("Something went wrong while updating LSv4_Topology Edge: " + lsv4_topology_key)

