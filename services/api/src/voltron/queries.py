from .db import ArangoDBConnection
from .util import ip_network_bandaid

def pathing_epe_bandwidth_get(dst_ip, min_bandwidth=None):
    aql = """
    FOR edge in ExternalLinkEdges
        FOR v, e, p IN 1..2 OUTBOUND edge._from ExternalLinkEdges, ExternalPrefixEdges
            FILTER p.vertices[2].Prefix == @prefix_ip && p.vertices[2].Length == @prefix_mask
            LET egress_bandwidth = (p.edges[0].speed * 1000) - (p.edges[0].out_octets * 8)
            {bandwidth_filter}
            SORT egress_bandwidth DESC
            LIMIT 1
            RETURN [p.vertices[0].SRNodeSID, p.edges[0].Label]
    """
    prefix = ip_network_bandaid(dst_ip)
    prefix_ip = str(prefix.network_address)
    prefix_mask = int(prefix.prefixlen)
    bind_vars = { 'prefix_ip': prefix_ip, 'prefix_mask': prefix_mask }
    if min_bandwidth is not None:
        aql = aql.format(bandwidth_filter='FILTER egress_bandwidth >= @min_bandwidth')
        bind_vars['min_bandwidth'] = min_bandwidth
    db = ArangoDBConnection()
    label_list = list(db.query_aql(aql, bind_vars))
    if len(label_list) > 0:
        label_list = label_list[0]
    return label_list

def pathing_epe_utilization_get(dst_ip, max_utilization=None):
    aql = """
    FOR edge in ExternalLinkEdges
        FOR v, e, p IN 1..2 OUTBOUND edge._from ExternalLinkEdges, ExternalPrefixEdges
            FILTER p.vertices[2].Prefix == @prefix_ip && p.vertices[2].Length == @prefix_mask
            {utilization_filter}
            SORT p.edges[0].percent_util_outbound
            LIMIT 1
            RETURN [p.vertices[0].SRNodeSID, p.edges[0].Label]
    """
    prefix = ip_network_bandaid(dst_ip)
    prefix_ip = str(prefix.network_address)
    prefix_mask = int(prefix.prefixlen)
    bind_vars = { 'prefix_ip': prefix_ip, 'prefix_mask': prefix_mask }
    if max_utilization is not None:
        aql = aql.format(utilization_filter='FILTER p.edges[0].percent_util_outbound <= @max_utilization')
        bind_vars['max_utilization'] = max_utilization
    db = ArangoDBConnection()
    label_list = list(db.query_aql(aql, bind_vars))
    if len(label_list) > 0:
        label_list = label_list[0]
    return label_list

def pathing_epe_latency_get(src_ip, src_transport_ip, dst_ip, max_latency=None):
    aql = """
    FOR p IN EPEPaths_Latency
        FILTER p.Source == @source AND p.Destination == @destination
        {latency_filter}
        SORT p.Latency
        LIMIT 1
        RETURN [p._key, p.Label_Path]
    """
    prefix = ip_network_bandaid(dst_ip)
    prefix_ip = str(prefix.network_address)
    bind_vars = {'source': src_ip, 'destination': prefix_ip}
    if max_latency is not None:
        aql = aql.format(latency_filter='FILTER p.Latency <= @max_latency')
        bind_vars['max_latency'] = max_latency
    db = ArangoDBConnection()
    label_list = list(db.query_aql(aql, bind_vars))
    if len(label_list) > 0:
        label_list = label_list[0]
    return label_list

def pathing_epe_lossless_get(dst_ip, max_loss=None):
    aql = """
    FOR edge in ExternalLinkEdges
        FOR v, e, p IN 1..2 OUTBOUND edge._from ExternalLinkEdges, ExternalPrefixEdges
            FILTER p.vertices[2].Prefix == @prefix_ip && p.vertices[2].Length == @prefix_mask
            LET total_loss = p.edges[0].out_errors + p.edges[0].out_discards
            {loss_filter}
            SORT total_loss
            LIMIT 1
            RETURN [p.vertices[0].SRNodeSID, p.edges[0].Label]
    """
    prefix = ip_network_bandaid(dst_ip)
    prefix_ip = str(prefix.network_address)
    prefix_mask = int(prefix.prefixlen)
    bind_vars = { 'prefix_ip': prefix_ip, 'prefix_mask': prefix_mask }
    if max_loss is not None:
        aql = aql.format(loss_filter='FILTER total_loss <= @max_loss')
        bind_vars['max_loss'] = max_loss
    label_list = list(db.query_aql(aql, bind_vars))
    if len(label_list) > 0:
        label_list = label_list[0]
    return label_list

def topology_get():
    aql_node_router = """
    FOR router IN Routers
        RETURN router._id
    """
    aql_node_prefix = """
    FOR prefix IN Prefixes
        RETURN prefix._id
    """
    aql_link_external_link_edge = """
    FOR ext_link IN ExternalLinkEdges
        RETURN {
            "source": ext_link._from,
            "target": ext_link._to,
            "value": ext_link.Label
        }
    """
    aql_link_internal_link_edge = """
    FOR int_link IN InternalLinkEdges
        RETURN {
            "source": int_link._from,
            "target": int_link._to,
            "value": int_link.Label
        }
    """
    aql_link_epe_edge = """
    FOR epe_edge IN EPEEdges
        RETURN {
            "source": epe_edge._from,
            "target": epe_edge._to,
            "value": epe_edge.DestinationASN
        }
    """
    aql_link_ext_prefix_edge = """
    FOR prefix_edge IN ExternalPrefixEdges
        RETURN {
            "source": prefix_edge._from,
            "target": prefix_edge._to,
            "value": prefix_edge.DstPrefixASN
        }
    """
    db = ArangoDBConnection()
    nodes = []
    node_router = list(db.query_aql(aql_node_router))
    for node in node_router:
        nodes.append({'id': node, 'group': 1})
    node_prefix = list(db.query_aql(aql_node_prefix))
    for node in node_prefix:
        nodes.append({'id': node, 'group': 2})
    links = []
    link_external_link = list(db.query_aql(aql_link_external_link_edge))
    for link in link_external_link:
        link['value'] = int(link['value'])
        links.append(link)
    link_internal_link = list(db.query_aql(aql_link_internal_link_edge))
    for link in link_internal_link:
        link['value'] = int(link['value'])
        links.append(link)
    link_epe = list(db.query_aql(aql_link_epe_edge))
    for link in link_epe:
        link['value'] = int(link['value'])
        links.append(link)
    link_ext_prefix = list(db.query_aql(aql_link_ext_prefix_edge))
    for link in link_ext_prefix:
        link['value'] = int(link['value'])
        links.append(link)
    return {
        'nodes': nodes,
        'links': links
    }
