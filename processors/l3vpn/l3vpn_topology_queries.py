#! /usr/bin/env python
"""AQL Queries executed by the L3VPN_Topology Service."""

def get_prefix_data(db):
    aql = """ FOR p in L3VPNPrefix return p """
    bindVars = {}
    prefix_data = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return prefix_data

def get_prefixSID(db, routerID):
    aql = """ FOR e in LSNode FILTER e._key == @ls_node_key return e.PrefixSID """
    bindVars = {'ls_node_key': routerID}
    prefixSID = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return prefixSID

def get_router_igp_id(db, router_id):
    aql = """ FOR l in LSNode filter l.RouterID == @router_id return l.IGPID """
    bindVars = {'router_id': router_id }
    router_igp_id = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return router_igp_id

def get_srgb_start(db, igp_router_id):
    aql = """ FOR l in LSNode filter l._key == @igp_router_id return l.SRGBStart """
    bindVars = {'igp_router_id': igp_router_id }
    srgb_start = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return srgb_start

def get_sid_index(db, igp_router_id):
    aql = """ FOR l in LSPrefix filter l.IGPRouterID == @igp_router_id and l.SIDIndex != null and l.SRFlags != null return l.SIDIndex """
    bindVars = {'igp_router_id': igp_router_id }
    sid_index = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return sid_index

def get_prefix_info(db, igp_router_id):
    aql = """ FOR l in LSPrefix filter l.IGPRouterID == @igp_router_id return {"SIDIndex": l.SIDIndex} """
    bindVars = {'igp_router_id': igp_router_id }
    prefix_info = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return prefix_info

#def get_origin_as(db, origin_as):
#    aql = """ FOR p in L3VPNPrefix filter p.Origin_AS return p.Origin_AS """
#    bindVars = {'origin_as': origin_as }
#    origin_as = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
#    return origin_as

def get_all_rds(db):
    aql = """ RETURN MERGE (For r in L3VPNNode return { ["RDs"]: r.RD}) """
    bindVars = {}
    rds = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return rds

def get_l3vpn_nodes_from_rd(db, rd):
    aql = """ For r in L3VPNNode filter POSITION (r.RD, @route_distinguisher) == TRUE return r._key """
    bindVars = {'route_distinguisher': rd}
    l3vpn_nodes = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return l3vpn_nodes

def get_l3vpn_topology_edge_key(db, l3vpn_topology_edge_key):
    aql = """ FOR e in L3VPN_Topology 
              FILTER e._key == @l3vpn_topology_edge_key
              RETURN e._key """
    bindVars = {'l3vpn_topology_edge_key': l3vpn_topology_edge_key}
    l3vpn_topology_edge_key = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return l3vpn_topology_edge_key

def get_l3vpn_fib_edge_key(db, l3vpn_fib_edge_key):
    aql = """ FOR e in L3VPN_FIB 
              FILTER e._key == @l3vpn_fib_edge_key
              RETURN e._key """
    bindVars = {'l3vpn_fib_edge_key': l3vpn_fib_edge_key}
    l3vpn_fib_edge_key = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return l3vpn_fib_edge_key

def update_node_to_node_topology_edge_query(db, l3vpn_topology_edge_key, rd, source, destination):
    l3vpn_topology_edge_from = 'L3VPNNode/'+source
    l3vpn_topology_edge_to = 'L3VPNNode/'+destination
    aql = """ FOR e in L3VPN_Topology
	      FILTER e._key == @l3vpn_topology_edge_key
	      UPDATE {
                  _key: e._key,
                  _from: @l3vpn_topology_edge_from,
                  _to: @l3vpn_topology_edge_to,
                  SrcIP: @src_ip,
                  DstIP: @dst_ip,
                  Source: @source,
                  Destination: @destination,
                  RD: @rd }
              IN L3VPN_Topology RETURN { before: OLD, after: NEW } """
    bindVars = {'l3vpn_topology_edge_key': l3vpn_topology_edge_key, 'l3vpn_topology_edge_from': l3vpn_topology_edge_from,
                'l3vpn_topology_edge_to': l3vpn_topology_edge_to, 'src_ip': source, 'dst_ip': destination,
                'source': source, 'destination': destination, 'rd': rd}
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated L3VPN_Topology Edge: " + l3vpn_topology_edge_key)
        pass
    else:
        print("Something went wrong while updating L3VPN_Topology Edge")

def create_node_to_node_topology_edge_query(db, l3vpn_topology_edge_key, rd, source, destination):
    l3vpn_topology_edge_from = 'L3VPNNode/'+str(source)
    l3vpn_topology_edge_to = 'L3VPNNode/'+str(destination)
    aql = """ INSERT {
                  _key: @l3vpn_topology_edge_key,
                  _to: @l3vpn_topology_edge_to,
                  _from: @l3vpn_topology_edge_from,
                  SrcIP: @src_ip,
                  DstIP: @dst_ip,
                  Source: @source,
                  Destination: @destination,
                  RD: @rd }
              INTO L3VPN_Topology RETURN NEW._key """
    bindVars = {'l3vpn_topology_edge_key': l3vpn_topology_edge_key, 'l3vpn_topology_edge_from': l3vpn_topology_edge_from,
                'l3vpn_topology_edge_to': l3vpn_topology_edge_to, 'src_ip': source, 'dst_ip': destination,
                'source': source, 'destination': destination, 'rd': rd}
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        print("Successfully created L3VPN_Topology Edge: " + l3vpn_topology_edge_key)
        pass
    else:
        print("Something went wrong while creating L3VPN_Topology Edge")

def update_prefix_to_node_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid):
    l3vpn_topology_edge_from = 'L3VPPrefix/'+str(prefix)
    l3vpn_topology_edge_to = 'L3VPNNode/'+str(router_id)
    aql = """ FOR e in L3VPN_Topology
	      FILTER e._key == @l3vpn_topology_edge_key
	      UPDATE { 
                  _key: @l3vpn_topology_edge_key,
                  _to: @l3vpn_topology_edge_to,
                  _from: @l3vpn_topology_edge_from,
                  SrcIP: @prefix,
                  DstIP: @router_id,
                  VPN_Prefix: @prefix,
                  VPN_Prefix_Len: @prefix_length,
                  RouterID: @router_id,
                  PrefixSID: @prefix_sid,
                  VPN_Label: @vpn_label,
                  RD: @rd,
                  RT: @rt,
                  IPv4: @ipv4,
                  SRv6_SID: @srv6_sid,
                  Source: @prefix,
                  Destination: @router_id }
              IN L3VPN_Topology RETURN { before: OLD, after: NEW } """
    bindVars = {'l3vpn_topology_edge_key': l3vpn_topology_edge_key, 'l3vpn_topology_edge_from': l3vpn_topology_edge_from,
                'l3vpn_topology_edge_to': l3vpn_topology_edge_to, 'prefix': prefix, 'prefix_length': prefix_length,
                'router_id': router_id, 'prefix_sid': prefix_sid, 'vpn_label': vpn_label, 'rd': rd, 'rt': rt, 'ipv4': ipv4, 'srv6_sid': srv6_sid}
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated L3VPN_Topology Edge: " + l3vpn_topology_edge_key)
        pass
    else:
        print("Something went wrong while updating L3VPN_Topology Edge")

def create_prefix_to_node_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid):
    l3vpn_topology_edge_from = 'L3VPNPrefix/'+str(prefix)
    l3vpn_topology_edge_to = 'L3VPNNode/'+str(router_id)
    aql = """ INSERT {
                  _key: @l3vpn_topology_edge_key,
                  _to: @l3vpn_topology_edge_to,
                  _from: @l3vpn_topology_edge_from,
                  SrcIP: @prefix,
                  DstIP: @router_id,
                  VPN_Prefix: @prefix,
                  VPN_Prefix_Len: @prefix_length,
                  RouterID: @router_id,
                  PrefixSID: @prefix_sid,
                  VPN_Label: @vpn_label,
                  RD: @rd,
                  RT: @rt,
                  IPv4: @ipv4,
                  SRv6_SID: @srv6_sid,
                  Source: @prefix,
                  Destination: @router_id }
              INTO L3VPN_Topology RETURN NEW._key """
    bindVars = {'l3vpn_topology_edge_key': l3vpn_topology_edge_key, 'l3vpn_topology_edge_from': l3vpn_topology_edge_from,
                'l3vpn_topology_edge_to': l3vpn_topology_edge_to, 'prefix': prefix, 'prefix_length': prefix_length,
                'router_id': router_id, 'prefix_sid': prefix_sid, 'vpn_label': vpn_label, 'rd': rd, 'rt': rt, 'ipv4': ipv4, 'srv6_sid': srv6_sid}
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        print("Successfully created L3VPN_Topology Edge: " + l3vpn_topology_edge_key)
        pass
    else:
        print("Something went wrong while creating L3VPN_Topology Edge")

def update_node_to_prefix_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid):
    l3vpn_topology_edge_from = 'L3VPNode/'+str(router_id)
    l3vpn_topology_edge_to = 'L3VPNPrefix/'+str(prefix)
    aql = """ FOR e in L3VPN_Topology
	      FILTER e._key == @l3vpn_topology_edge_key
	      UPDATE { 
                  _key: @l3vpn_topology_edge_key,
                  _to: @l3vpn_topology_edge_to,
                  _from: @l3vpn_topology_edge_from,
                  SrcIP: @router_id,
                  DstIP: @prefix,
                  VPN_Prefix: @prefix,
                  VPN_Prefix_Len: @prefix_length,
                  RouterID: @router_id,
                  PrefixSID: @prefix_sid,
                  VPN_Label: @vpn_label,
                  RD: @rd,
                  RT: @rt,
                  IPv4: @ipv4,
                  SRv6_SID: @srv6_sid,
                  Source: @router_id,
                  Destination: @prefix }
              IN L3VPN_Topology RETURN { before: OLD, after: NEW } """
    bindVars = {'l3vpn_topology_edge_key': l3vpn_topology_edge_key, 'l3vpn_topology_edge_from': l3vpn_topology_edge_from,
                'l3vpn_topology_edge_to': l3vpn_topology_edge_to, 'prefix': prefix, 'prefix_length': prefix_length,
                'router_id': router_id, 'prefix_sid': prefix_sid, 'vpn_label': vpn_label, 'rd': rd, 'rt': rt, 'ipv4': ipv4, 'srv6_sid': srv6_sid}
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated L3VPN_Topology Edge: " + l3vpn_topology_edge_key)
        pass
    else:
        print("Something went wrong while updating L3VPN_Topology Edge")

def create_node_to_prefix_topology_edge_query(db, l3vpn_topology_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, srv6_sid):
    l3vpn_topology_edge_from = 'L3VPNNode/'+str(router_id)
    l3vpn_topology_edge_to = 'L3VPNPrefix/'+str(prefix)
    aql = """ INSERT {
                  _key: @l3vpn_topology_edge_key,
                  _to: @l3vpn_topology_edge_to,
                  _from: @l3vpn_topology_edge_from,
                  SrcIP: @router_id,
                  DstIP: @prefix,
                  VPN_Prefix: @prefix,
                  VPN_Prefix_Len: @prefix_length,
                  RouterID: @router_id,
                  PrefixSID: @prefix_sid,
                  VPN_Label: @vpn_label,
                  RD: @rd,
                  RT: @rt,
                  IPv4: @ipv4,
                  SRv6_SID: @srv6_sid,
                  Source: @router_id,
                  Destination: @prefix }
              INTO L3VPN_Topology RETURN NEW._key """
    bindVars = {'l3vpn_topology_edge_key': l3vpn_topology_edge_key, 'l3vpn_topology_edge_from': l3vpn_topology_edge_from,
                'l3vpn_topology_edge_to': l3vpn_topology_edge_to, 'prefix': prefix, 'prefix_length': prefix_length,
                'router_id': router_id, 'prefix_sid': prefix_sid, 'vpn_label': vpn_label, 'rd': rd, 'rt': rt, 'ipv4': ipv4, 'srv6_sid': srv6_sid}
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        print("Successfully created L3VPN_Topology Edge: " + l3vpn_topology_edge_key)
        pass
    else:
        print("Something went wrong while creating L3VPN_Topology Edge")



def update_l3vpn_fib_edge_query(db, l3vpn_fib_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, origin_as, srv6_sid):
    l3vpn_fib_edge_from = 'L3VPNode/'+str(router_id)
    l3vpn_fib_edge_to = 'L3VPNPrefix/'+str(prefix)
    aql = """ FOR e in L3VPN_FIB
	      FILTER e._key == @l3vpn_fib_edge_key
	      UPDATE { 
                  _key: @l3vpn_fib_edge_key,
                  _to: @l3vpn_fib_edge_to,
                  _from: @l3vpn_fib_edge_from,
                  SrcIP: @router_id,
                  DstIP: @prefix,
                  VPN_Prefix: @prefix,
                  VPN_Prefix_Len: @prefix_length,
                  RouterID: @router_id,
                  PrefixSID: @prefix_sid,
                  VPN_Label: @vpn_label,
                  RD: @rd,
                  RT: @rt,
                  IPv4: @ipv4,
                  Origin_AS: @origin_as,
                  SRv6_SID: @srv6_sid }
              IN L3VPN_FIB RETURN { before: OLD, after: NEW } """
    bindVars = {'l3vpn_fib_edge_key': l3vpn_fib_edge_key, 'l3vpn_fib_edge_from': l3vpn_fib_edge_from,
                'l3vpn_fib_edge_to': l3vpn_fib_edge_to, 'prefix': prefix, 'prefix_length': prefix_length,
                'router_id': router_id, 'prefix_sid': prefix_sid, 'vpn_label': vpn_label, 'rd': rd, 'rt': rt, 'ipv4': ipv4, 'origin_as': origin_as, 'srv6_sid': srv6_sid}
    updated_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(updated_edge) > 0):
        print("Successfully updated L3VPN_FIB Edge: " + l3vpn_fib_edge_key)
        pass
    else:
        print("Something went wrong while updating L3VPN_FIB Edge")

def create_l3vpn_fib_edge_query(db, l3vpn_fib_edge_key, prefix, prefix_length, router_id, prefix_sid, vpn_label, rd, rt, ipv4, origin_as, srv6_sid):
    l3vpn_fib_edge_from = 'L3VPNNode/'+str(router_id)
    l3vpn_fib_edge_to = 'L3VPNPrefix/'+str(prefix)
    aql = """ INSERT {
                  _key: @l3vpn_fib_edge_key,
                  _to: @l3vpn_fib_edge_to,
                  _from: @l3vpn_fib_edge_from,
                  SrcIP: @router_id,
                  DstIP: @prefix,
                  VPN_Prefix: @prefix,
                  VPN_Prefix_Len: @prefix_length,
                  RouterID: @router_id,
                  PrefixSID: @prefix_sid,
                  VPN_Label: @vpn_label,
                  RD: @rd,
                  RT: @rt,
                  IPv4: @ipv4,
                  Origin_AS: @origin_as,
                  SRv6_SID: @srv6_sid }
              INTO L3VPN_FIB RETURN NEW._key """
    bindVars = {'l3vpn_fib_edge_key': l3vpn_fib_edge_key, 'l3vpn_fib_edge_from': l3vpn_fib_edge_from,
                'l3vpn_fib_edge_to': l3vpn_fib_edge_to, 'prefix': prefix, 'prefix_length': prefix_length,
                'router_id': router_id, 'prefix_sid': prefix_sid, 'vpn_label': vpn_label, 'rd': rd, 'rt': rt, 'ipv4': ipv4, 'origin_as': origin_as, 'srv6_sid': srv6_sid}
    created_edge = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    if(len(created_edge) > 0):
        print("Successfully created L3VPN_FIB Edge: " + l3vpn_fib_edge_key)
        pass
    else:
        print("Something went wrong while creating L3VPN_FIB Edge")

