#! /usr/bin/env python
"""AQL Queries executed by the LSv6_Topology Service."""

def get_lsv6_topology_keys(db):
    aql = """ FOR l in LSv6_Topology return l._key """
    bindVars = {}
    allLSLinks = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return allLSLinks

def get_disjoint_keys(db, lsv6_topology_keys):
    aql = """ FOR l in LSLinkEdge filter l._key not in @lsv6_topology_keys return l._key """
    bindVars = {'lsv6_topology_keys': lsv6_topology_keys }
    uncreated_ls_link_keys = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return uncreated_ls_link_keys

def get_lsv6_link_keys(db):
    aql = """ FOR l in LSLinkEdge filter CONTAINS(l.local_interface_ip, ":") or CONTAINS(l.remote_interface_ip, ":") return l._key """
    bindVars = {}
    ls_link_keys = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return ls_link_keys

def check_exists_lsv6_topology(db, ls_link_key):
    aql = """ FOR l in LSv6_Topology filter l._key == @ls_link_key RETURN { key: l._key } """
    bindVars = {'ls_link_key': ls_link_key}
    key_exists = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(key_exists) > 0):
        return True
    else:
        return False

def check_exists_lsv6_link(db, lsv6_topology_key):
    aql = """ FOR l in LSLinkEdge filter l._key == @lsv6_topology_key RETURN { key: l._key } """
    bindVars = {'lsv6_topology_key': lsv6_topology_key}
    key_exists = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(key_exists) > 0):
        return True
    else:
        return False

def deleteLSv6TopologyDocument(db, lsv6_topology_key):
    aql = """ FOR l in LSv6_Topology filter l._key == @lsv6_topology_key remove l in LSv6_Topology """
    bindVars = {'lsv6_topology_key': lsv6_topology_key}
    deleted_document = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)

def createBaseLSv6TopologyDocument(db, ls_link_key):
    aql = """ FOR l in LSLinkEdge filter l._key == @ls_link_key insert l into LSv6_Topology RETURN NEW._key """
    bindVars = {'ls_link_key': ls_link_key}
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        #print("Successfully created LSv6_Topology Edge: " + ls_link_key)
        pass
    else:
        print("Something went wrong while creating LSv6_Topology Edge")

def updateBaseLSv6TopologyDocument(db, ls_link_key):
    aql = """ FOR l in LSLinkEdge filter l._key == @ls_link_key update l into LSv6_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'ls_link_key': ls_link_key }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        #print("Successfully updated LSv6_Topology Edge: " + ls_link_key)
        pass
    else:
        print("Something went wrong while updated LSv6_Topology Edge")

def enhance_lsv6_topology_document(db, lsv6_topology_key):
    aql = """ FOR l in LSv6_Topology filter l._key == @lsv6_topology_key UPDATE { _key: l._key, "link_delay": "", "local_msd": "",
    "remote_msd": "", "pq_resv_bw": "", "app_resv_bw": "" } in LSv6_Topology
    RETURN { before: OLD, after: NEW }"""
    bindVars = {'lsv6_topology_key': lsv6_topology_key }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        #print("Successfully enhanced LSv6_Topology Edge: " + lsv6_topology_key)
        pass
    else:
        print("Something went wrong while enhancing LSv6_Topology Edge")

def get_max_sid_depth(db, igp_router_id):
    aql = """ FOR l in LSNode filter l._key == @igp_router_id return l.node_msd """
    bindVars = {'igp_router_id': igp_router_id }
    max_sid_depth = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return max_sid_depth

#def get_prefix_info(db, igp_router_id):
#    aql = """ FOR l in LSPrefix filter l.IGPRouterID == @igp_router_id and l.SIDIndex != null return {"Prefix": l.Prefix, "Length": l.Length, "SIDIndex": l.SIDIndex, "SRFlag": l.SRFlags } """
#    bindVars = {'igp_router_id': igp_router_id }
#    prefix_info = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
#    return prefix_info

def get_local_igpid(db, lsv6_topology_key):
    aql = """ FOR l in LSv6_Topology filter l._key == @lsv6_topology_key return {"local_igp_id": l.local_igp_id} """
    bindVars = {'lsv6_topology_key': lsv6_topology_key }
    local_igp_id = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return local_igp_id

def get_remote_igpid(db, lsv6_topology_key):
    aql = """ FOR l in LSv6_Topology filter l._key == @lsv6_topology_key return {"remote_igp_id": l.remote_igp_id} """
    bindVars = {'lsv6_topology_key': lsv6_topology_key }
    remote_igp_id = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return remote_igp_id

def get_srv6_info(db, igp_router_id):
    aql = """ FOR l in LSSRv6SID filter l._key == @igp_router_id return {"protocol": l.protocol, "mt_id": l.mt_id, "srv6_sid": l.srv6_sid, "srv6_endpoint_behavior": l.srv6_endpoint_behavior, "srv6_sid_structure": l.srv6_sid_structure} """
    bindVars = {'igp_router_id': igp_router_id}
    srv6_info = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return srv6_info

#def update_prefix_sid(db, lsv6_topology_key, local_prefix_sid, remote_prefix_sid):
#    aql = """ FOR l in LSv6_Topology filter l._key == @lsv6_topology_key UPDATE { _key: l._key, "LocalPrefixSID": @local_prefix_sid, "RemotePrefixSID": @remote_prefix_sid  } in LSv6_Topology RETURN { before: OLD, after: NEW }"""
#    bindVars = {'lsv6_topology_key': lsv6_topology_key, 'local_prefix_sid': local_prefix_sid, 'remote_prefix_sid': remote_prefix_sid }
#    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
#    if(len(updated_edge) > 0):
#        print("Successfully updated LSv6_Topology Edge: " + lsv6_topology_key + " with LocalPrefixSID " + str(local_prefix_sid) + " and RemotePrefixSID " + str(remote_prefix_sid))
#        pass
#    else:
#        print("Something went wrong while updating LSv6_Topology Edge with PrefixSIDs")

#def update_prefix_info(db, lsv6_topology_key, local_prefix_info, remote_prefix_info):
#    aql = """ FOR l in LSv6_Topology filter l._key == @lsv6_topology_key UPDATE { _key: l._key, "LocalPrefixInfo": @local_prefix_info, "RemotePrefixInfo": @remote_prefix_info  } in LSv6_Topology RETURN { before: OLD, after: NEW }"""
#    bindVars = {'lsv6_topology_key': lsv6_topology_key, 'local_prefix_info': local_prefix_info, 'remote_prefix_info': remote_prefix_info }
#    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
#    if(len(updated_edge) > 0):
#        print("Successfully updated LSv6_Topology Edge: " + lsv6_topology_key + " with LocalPrefixInfo " + str(local_prefix_info) + " and RemotePrefixInfo " + str(remote_prefix_info))
#        pass
#    else:
#        print("Something went wrong while updating LSv6_Topology Edge with PrefixInfo")

def update_lsv6_topology_document(db, lsv6_topology_key, local_srv6_info, remote_srv6_info):
    aql = """ FOR l in LSv6_Topology filter l._key == @lsv6_topology_key UPDATE { _key: l._key,
              "local_protocol": @local_protocol, "local_mt_id": @local_mt_id, "local_srv6_sid": @local_srv6_sid, "local_srv6_sid_endpoint_behavior": @local_srv6_sid_endpoint_behavior,
              "local_srv6_sid_structure": @local_srv6_sid_structure, "remote_protocol": @remote_protocol, "remote_mt_id": @remote_mt_id, "remote_srv6_sid": @remote_srv6_sid,
              "remote_srv6_sid_endpoint_behavior": @remote_srv6_sid_endpoint_behavior, "remote_srv6_sid_structure": @remote_srv6_sid_structure}
              in LSv6_Topology RETURN { before: OLD, after: NEW }"""
    bindVars = {'lsv6_topology_key': lsv6_topology_key, 'local_protocol': local_srv6_info["protocol"], 'local_mt_id': local_srv6_info["mt_id"],
                'local_srv6_sid': local_srv6_info["srv6_sid"], 'local_srv6_sid_endpoint_behavior': local_srv6_info["srv6_endpoint_behavior"], "local_srv6_sid_structure": local_srv6_info["srv6_sid_structure"],
                'remote_protocol': remote_srv6_info["protocol"], 'remote_mt_id': remote_srv6_info["mt_id"], 'remote_srv6_sid': remote_srv6_info["srv6_sid"], 
                'remote_srv6_sid_endpoint_behavior': remote_srv6_info["srv6_endpoint_behavior"], 'remote_srv6_sid_structure': remote_srv6_info["srv6_sid_structure"] }
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated LSv6_Topology Edge: " + lsv6_topology_key + " with srv6 sid " + str(local_srv6_info["srv6_sid"]))
        pass
    else:
        print("Something went wrong while updating LSv6_Topology Edge: " + lsv6_topology_key)

#def update_lsv6_topology_document(db, lsv6_topology_key, local_prefix_sid, remote_prefix_sid, local_prefix_info, remote_prefix_info, local_max_sid_depth, remote_max_sid_depth, local_srv6_info, remote_srv6_info):
#    aql = """ FOR l in LSv6_Topology filter l._key == @lsv6_topology_key UPDATE { _key: l._key, "LocalPrefixSID": @local_prefix_sid, "RemotePrefixSID": @remote_prefix_sid,
#              "LocalPrefixInfo": @local_prefix_info, "RemotePrefixInfo": @remote_prefix_info, "LocalMaxSIDDepth": @local_max_sid_depth, "RemoteMaxSIDDepth": @remote_max_sid_depth,
#              "Local_Protocol": @local_protocol, "Local_MT_ID": @local_mt_id, "Local_SRv6_SID": @local_srv6_sid, "Local_SRv6_SID_Endpoint_Behavior": @local_srv6_sid_endpoint_behavior,
#              "Local_SRv6_SID_Structure": @local_srv6_sid_structure, "Remote_Protocol": @remote_protocol, "Remote_MT_ID": @remote_mt_id, "Remote_SRv6_SID": @remote_srv6_sid,
#              "Remote_SRv6_SID_Endpoint_Behavior": @remote_srv6_sid_endpoint_behavior, "Remote_SRv6_SID_Structure": @remote_srv6_sid_structure}
#              in LSv6_Topology RETURN { before: OLD, after: NEW }"""
#    bindVars = {'lsv6_topology_key': lsv6_topology_key, 'local_prefix_sid': local_prefix_sid, 'remote_prefix_sid': remote_prefix_sid, 'local_prefix_info': local_prefix_info, 'remote_prefix_info': remote_prefix_info,
#                'local_max_sid_depth': local_max_sid_depth, 'remote_max_sid_depth': remote_max_sid_depth, 'local_protocol': local_srv6_info["Protocol"],  'local_mt_id': local_srv6_info["MT_ID"],
#                'local_srv6_sid': local_srv6_info["SRv6_SID"], 'local_srv6_sid_endpoint_behavior': local_srv6_info["SRv6_Endpoint_Behavior"], "local_srv6_sid_structure": local_srv6_info["SRv6_SID_Structure"],
#                'remote_protocol': remote_srv6_info["Protocol"], 'remote_mt_id': remote_srv6_info["MT_ID"], 'remote_srv6_sid': remote_srv6_info["SRv6_SID"], 
#                'remote_srv6_sid_endpoint_behavior': remote_srv6_info["SRv6_Endpoint_Behavior"], 'remote_srv6_sid_structure': remote_srv6_info["SRv6_SID_Structure"] }
#    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
#    if(len(updated_edge) > 0):
#        print("Successfully updated LSv6_Topology Edge: " + lsv6_topology_key + " with LocalPrefixSID " + str(local_prefix_sid) + " and RemotePrefixSID " + str(remote_prefix_sid))
#        pass
#    else:
#        print("Something went wrong while updating LSv6_Topology Edge: " + lsv6_topology_key)

