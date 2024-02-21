# Example Arango DB queries

This section will cover some example queries that can be run against the ArangoDB instance.

!!! info
    Many of these queries are specific to a lab setup we run internally. A reference topology has been provided below.

    Please note, the queries will need to be adjusted to fit your environment / needs.

![topology](../img/topology.png "jalapeno topology")

## Link State Collection Queries

```text
for l in ls_node return l
for l in ls_node_edge return l
for l in ls_node_edge  return { from: l._from, to: l._to }
for l in ls_node filter l.router_id == "10.0.0.8" return l
for l in ls_node_edge filter l._key like "%0019%" return l
for l in ls_link filter l.mt_id_tlv.mt_id != 2 return l._key
for l in ls_link filter l.protocol_id == 7 && l.peer_asn == 100000 && l.remote_link_ip == "10.71.0.1" return { epe_sid: l.peer_node_sid.sid } 
for l in ls_prefix filter l.prefix == "10.0.0.8" return l
for l in ls_prefix return l
for l in ls_prefix filter l.prefix_attr_tlvs.ls_prefix_sid != null return l

for l in ls_srv6_sid filter l.igp_router_id == "0000.0000.0018" for m in ls_srv6_sid filter m.igp_router_id == "0000.0000.0017" for n in ls_srv6_sid filter n.igp_router_id == "0000.0000.0016" for o in ls_srv6_sid filter o.igp_router_id == "0000.0000.0021" return [l.srv6_sid, m.srv6_sid, n.srv6_sid, o.srv6_sid]
```

## Query Other Collections

```text
for u in unicast_prefix_v4 return u._key
for l in unicast_prefix_v4 filter l.prefix == "10.10.3.0" filter l.base_attrs.as_path == Null return l
for d in peer filter d.remote_ip == "10.72.0.1"  for l in ls_link filter d.remote_ip == l.remote_link_ip return d
fOR d IN peer filter d.remote_bgp_id == "10.0.0.71" filter d.remote_ip == "10.71.0.1" return d
for l in unicast_prefix_v4 return { key: l._key, prefix: l.prefix, nexthop: l.nexthop, as_path: l.base_attrs.as_path, origin_as: l.origin_as }
for p in unicast_prefix_v4 filter p._key == "10.71.8.0_22_10.71.0.1" return p
for l in l3vpn_v4_prefix filter l.base_attrs.ext_community_list like "%100:100%" return l//.base_attrs.ext_community_list
```

## Shortest path queries

**(Optional) Add synthetic latency:**

```text
for l in ls_node_edge UPDATE l with { link_latency: 5 } in ls_node_edge

for l in ls_node_edge filter l._key == "2_0_0_0_0000.0000.0003_10.1.1.19_0000.0000.0002_10.1.1.18" UPDATE l with { link_latency: 30 } in ls_node_edge
```

**Shortest Path:**

```text
for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0019' TO 'ls_node/2_0_0_0000.0000.0007' ls_node_edge return [v._key, e._key]

for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0019' TO 'ls_node/2_0_0_0000.0000.0007' ls_node_edge filter e.mt_id != 2 return e

for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0019' TO 'ls_node/2_0_0_0000.0000.0007' ls_node_edge filter e.mt_id != 2 return {node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn  } 

for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0019' TO 'ls_node/2_0_0_0000.0000.0007' ls_node_edge OPTIONS {weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency }
```

**Get all paths up to X hops in length:**

```text
for v, e, p IN 1..6 OUTBOUND 'ls_node/2_0_0_0000.0000.0019' ls_node_edge FILTER v._id == 'ls_node/2_0_0_0000.0000.0007' RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.edges[*].link_latency])

for v, e, p IN 1..16 OUTBOUND 'ls_node/2_0_0_0000.0000.0019' lsv4_edge OPTIONS {uniqueVertices: "path", bfs: true} FILTER v._id == 'ls_node/2_0_0_0000.0000.0007' RETURN p.edges[*]._to
```
