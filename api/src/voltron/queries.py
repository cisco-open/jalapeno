from .db import ArangoDBConnection
from .util import ip_network_bandaid

def pathing_epe_bandwidth_get(dst_ip, min_bandwidth=None, peer_preference=None, composite=None):
    aql = """
    FOR edge in ExternalLinkEdges
        FOR v, e, p IN 1..2 OUTBOUND edge._from ExternalLinkEdges, ExternalPrefixEdges
            FILTER p.vertices[2].Prefix == @prefix_ip && p.vertices[2].Length == @prefix_mask
            LET egress_bandwidth = (p.edges[0].speed * 1000) - (p.edges[0].out_octets * 8)
            {bandwidth_filter}
            SORT {peer_preference_filter} egress_bandwidth DESC
            {limit_filter}
            RETURN DISTINCT [p.vertices[0].SRNodeSID, p.edges[0].Label]
    """
    prefix = ip_network_bandaid(dst_ip)
    prefix_ip = str(prefix.network_address)
    prefix_mask = int(prefix.prefixlen)
    bind_vars = { 'prefix_ip': prefix_ip, 'prefix_mask': prefix_mask }

    if composite is not None:
        aql = aql.format(limit_filter='', bandwidth_filter='{bandwidth_filter}', peer_preference_filter='{peer_preference_filter}')
    else:
        aql = aql.format(limit_filter='LIMIT 1', bandwidth_filter='{bandwidth_filter}', peer_preference_filter='{peer_preference_filter}')

    if min_bandwidth is not None:
        aql = aql.format(bandwidth_filter='FILTER egress_bandwidth >= @min_bandwidth', peer_preference_filter='{peer_preference_filter}')
        bind_vars['min_bandwidth'] = min_bandwidth
    else:
        aql = aql.format(bandwidth_filter='', peer_preference_filter='{peer_preference_filter}')

    if peer_preference is not None:
        aql = aql.format(peer_preference_filter='POSITION([@peer_preference], p.vertices[1].PeerType, true) DESC,')
        bind_vars['peer_preference'] = str.capitalize(peer_preference)
    else:
        aql = aql.format(peer_preference_filter='')

    db = ArangoDBConnection()
    label_list = list(db.query_aql(aql, bind_vars))
    #if len(label_list) > 0:
    #    label_list = label_list[0]
    return label_list


def pathing_epe_utilization_get(dst_ip, max_utilization=None, peer_preference=None, composite=None):
    aql = """
    FOR edge in ExternalLinkEdges
        FOR v, e, p IN 1..2 OUTBOUND edge._from ExternalLinkEdges, ExternalPrefixEdges
            FILTER p.vertices[2].Prefix == @prefix_ip && p.vertices[2].Length == @prefix_mask
            {utilization_filter}
            SORT {peer_preference_filter} p.edges[0].percent_util_outbound
            {limit_filter}
            RETURN [p.vertices[0].SRNodeSID, p.edges[0].Label]
    """
    prefix = ip_network_bandaid(dst_ip)
    prefix_ip = str(prefix.network_address)
    prefix_mask = int(prefix.prefixlen)
    bind_vars = { 'prefix_ip': prefix_ip, 'prefix_mask': prefix_mask }

    if composite is not None:
        aql = aql.format(limit_filter='', utilization_filter='{utilization_filter}', peer_preference_filter='{peer_preference_filter}')
    else:
        aql = aql.format(limit_filter='LIMIT 1', utilization_filter='{utilization_filter}', peer_preference_filter='{peer_preference_filter}')

    if max_utilization is not None:
        aql = aql.format(utilization_filter='FILTER p.edges[0].percent_util_outbound <= @max_utilization', peer_preference_filter='{peer_preference_filter}')
        bind_vars['max_utilization'] = max_utilization
    else:
        aql = aql.format(utilization_filter='', peer_preference_filter='{peer_preference_filter}')


    if peer_preference is not None:
        aql = aql.format(peer_preference_filter='POSITION([@peer_preference], p.vertices[1].PeerType, true) DESC,')
        bind_vars['peer_preference'] = str.capitalize(peer_preference)
    else:
        aql = aql.format(peer_preference_filter='')

    db = ArangoDBConnection()
    label_list = list(db.query_aql(aql, bind_vars))
    #if len(label_list) > 0:
    #    label_list = label_list[0]
    return label_list


def pathing_epe_latency_all_get(src_ip, src_transport_ip):
    aql = """
    FOR p in EPEPaths_Latency
        FILTER p.Source == @src_ip
        RETURN  {"source": CONCAT (["Routers/", p.Source]), "target": CONCAT (["Prefixes/", p.Destination]), "SRNodeSID_EPELabel": p.Label_Path, "latency": p.Latency, "Egress_Peer": p.Egress_Peer}
    """
    bind_vars = {'src_ip': src_ip}
    db = ArangoDBConnection()
    label_list = list(db.query_aql(aql, bind_vars))
    return {
        'latencydata': label_list
    }


def pathing_epe_latency_get(src_ip, src_transport_ip, dst_ip, max_latency=None, peer_preference=None, composite=None):
    aql = """
    FOR p IN EPEPaths_Latency
        FILTER p.Source == @source AND p.Destination == @destination
        {latency_filter}
        SORT {peer_preference_filter} p.Latency
        {limit_filter}
        RETURN [p.Label_Path]
    """
    prefix = ip_network_bandaid(dst_ip)
    prefix_ip = str(prefix.network_address)
    bind_vars = {'source': src_ip, 'destination': prefix_ip}

    if composite is not None:
        aql = aql.format(limit_filter='', latency_filter='{latency_filter}', peer_preference_filter='{peer_preference_filter}')
    else:
        aql = aql.format(limit_filter='LIMIT 1', latency_filter='{latency_filter}', peer_preference_filter='{peer_preference_filter}')    

    if max_latency is not None:
        max_latency = str(max_latency/1000)
        aql = aql.format(latency_filter='FILTER p.Latency <= @max_latency', peer_preference_filter='{peer_preference_filter}')
        bind_vars['max_latency'] = max_latency
    else: 
        aql = aql.format(latency_filter='', peer_preference_filter='{peer_preference_filter}')

    if peer_preference is not None:
        aql = aql.format(peer_preference_filter='POSITION([@peer_preference], p.vertices[1].PeerType, true) DESC,')
        bind_vars['peer_preference'] = peer_preference
    else:
        aql = aql.format(peer_preference_filter='')

    db = ArangoDBConnection()
    label_list = list(db.query_aql(aql, bind_vars))
    #if len(label_list) > 0:
    #    label_list = label_list[0]
    return label_list

def pathing_epe_lossless_get(dst_ip, max_loss=None, peer_preference=None, composite=None):
    aql = """
    FOR edge in ExternalLinkEdges
        FOR v, e, p IN 1..2 OUTBOUND edge._from ExternalLinkEdges, ExternalPrefixEdges
            FILTER p.vertices[2].Prefix == @prefix_ip && p.vertices[2].Length == @prefix_mask
            LET total_loss = p.edges[0].out_errors + p.edges[0].out_discards
            {loss_filter}
            SORT {peer_preference_filter} total_loss
            {limit_filter}
            RETURN [p.vertices[0].SRNodeSID, p.edges[0].Label]
    """
    prefix = ip_network_bandaid(dst_ip)
    prefix_ip = str(prefix.network_address)
    prefix_mask = int(prefix.prefixlen)
    bind_vars = { 'prefix_ip': prefix_ip, 'prefix_mask': prefix_mask }

    if composite is not None:
        aql = aql.format(limit_filter='', loss_filter='{loss_filter}', peer_preference_filter='{peer_preference_filter}')
    else:
        aql = aql.format(limit_filter='LIMIT 1', loss_filter='{loss_filter}', peer_preference_filter='{peer_preference_filter}')

    if max_loss is not None:
        aql = aql.format(loss_filter='FILTER total_loss <= @max_loss', peer_preference_filter='{peer_preference_filter}')
        bind_vars['max_loss'] = max_loss
    else:
        aql = aql.format(loss_filter='', peer_preference_filter='{peer_preference_filter}')

    if peer_preference is not None:
        aql = aql.format(peer_preference_filter='POSITION([@peer_preference], p.vertices[1].PeerType, true) DESC,')
        bind_vars['peer_preference'] = str.capitalize(peer_preference)
    else:
        aql = aql.format(peer_preference_filter='')

    db = ArangoDBConnection()
    label_list = list(db.query_aql(aql, bind_vars))
    #if len(label_list) > 0:
    #    label_list = label_list[0]
    return label_list


def topology_get():
    aql_node_router_internal = """
    FOR router IN Routers
        FOR internal_router IN InternalRouters
            FILTER router._key == internal_router._key
                RETURN {
                    "id": router._id,
                    "label": internal_router._id,
                    "value": 1
                }
    """
    aql_node_router_external = """
    FOR router IN Routers
        FOR external_router IN ExternalRouters
            FILTER router._key == external_router._key
                RETURN {
                    "id": router._id,
                    "label": external_router._id,
                    "value": 2
                }
    """
    aql_node_prefix_external = """
    FOR prefix IN Prefixes
        FOR external_prefix IN ExternalPrefixes
            FILTER prefix._key == external_prefix._key
                RETURN {
                    "id": prefix._id,
                    "label": external_prefix._id,
                    "value": 3
                }
    """
    aql_link_external_link_edge = """
    FOR ext_link IN ExternalLinkEdges
        RETURN distinct {
            "source": ext_link._from,
            "target": ext_link._to,
            "value": ext_link.Label
        }
    """
    aql_link_internal_link_edge = """
    FOR int_link IN InternalLinkEdges
        RETURN distinct {
            "source": int_link._from,
            "target": int_link._to,
            "value": int_link.Label
        }
    """

    #aql_link_epe_edge = """
    #FOR epe_edge IN EPEEdges
    #    RETURN {
    #        "source": epe_edge._from,
    #        "target": epe_edge._to,
    #        "value": epe_edge.DestinationASN
    #    }
    #"""

    aql_link_ext_prefix_edge = """
    FOR prefix_edge IN ExternalPrefixEdges
        RETURN distinct {
            "source": prefix_edge._from,
            "target": prefix_edge._to,
            "value": prefix_edge.DstPrefixASN
        }
    """

    db = ArangoDBConnection()
    nodes = []
    node_router_internal = list(db.query_aql(aql_node_router_internal))
    for node in node_router_internal:
        nodes.append(node)
    node_router_external = list(db.query_aql(aql_node_router_external))
    for node in node_router_external:
        nodes.append(node)
    node_prefix_external = list(db.query_aql(aql_node_prefix_external))
    for node in node_prefix_external:
        nodes.append(node)
    links = []
    link_external_link = list(db.query_aql(aql_link_external_link_edge))
    for link in link_external_link:
        link['value'] = int(link['value'])
        links.append(link)
    link_internal_link = list(db.query_aql(aql_link_internal_link_edge))
    for link in link_internal_link:
        link['value'] = int(link['value'])
        links.append(link)
    #link_epe = list(db.query_aql(aql_link_epe_edge))
    #for link in link_epe:
    #    link['value'] = int(link['value'])
    #    links.append(link)
    link_ext_prefix = list(db.query_aql(aql_link_ext_prefix_edge))
    for link in link_ext_prefix:
        link['value'] = int(link['value'])
        links.append(link)
    return {
        'nodes': nodes,
        'edges': links
    }