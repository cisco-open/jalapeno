### Example Queries part 2
#### Note: you can use golang-style // to comment lines out

for v, e in outbound shortest_path 'sr_node/2_0_0_0000.0000.0025' TO 'unicast_prefix_v4/10.10.3.0_24_10.0.0.29' sr_topology return  { prefix: v.prefix, name: v.name, sid: e.srv6_sid, latency: e.latency }

for v, e in outbound shortest_path 'sr_node/2_0_0_0000.0000.0025' to 'unicast_prefix_v4/10.10.3.0_24_10.0.0.29' sr_topology OPTIONS {weightAttribute: 'latency' } return  { prefix: v.prefix, name: v.name, sid: e.srv6_sid, latency: e.latency }

for v, e, p IN 1..6 outbound 'sr_node/2_0_0_0000.0000.0025' sr_topology OPTIONS {uniqueVertices: "path", bfs: true} FILTER v._id == 'unicast_prefix_v4/10.10.3.0_24_10.0.0.29' return { path: p.edges[*].remote_node_name, sid: p.edges[*].srv6_sid, country_list: p.edges[*].country_codes[*], latency: sum(p.edges[*].latency), percent_util_out: avg(p.edges[*].percent_util_out)} 
    
for v, e in outbound shortest_path 'sr_node/2_0_0_0000.0000.0025' to 'unicast_prefix_v4/10.10.3.0_24_10.0.0.29' sr_topology OPTIONS {weightAttribute: 'latency' }  return  { prefix: v.prefix, name: v.name, sid: e.srv6_sid, latency: e.latency, cc: e.country_codes }

for p in outbound k_shortest_paths  'sr_node/2_0_0_0000.0000.0025' to 'unicast_prefix_v4/10.10.3.0_24_10.0.0.29' sr_topology  filter p.edges[*].country_codes !like "%FRA%" return { path: p.edges[*].remote_node_name, sid: p.edges[*].srv6_sid, country_list: p.edges[*].country_codes[*], latency: sum(p.edges[*].latency), percent_util_out: avg(p.edges[*].percent_util_out)} 

Basic queries

for l in ls_srv6_sid_edge return l
for l in ls_prefix return l
for l in ls_node_edge return l 
for l in ls_node_edge return l
for l in ls_node_edge filter l.protocol_id == 7 return l._key
for p in unicast_prefix_v4 filter p._key == "10.71.8.0_22_10.71.0.1" return p
FOR d IN peer filter d.remote_bgp_id == "10.0.0.71" filter d.remote_ip == "10.71.0.1" return d

for l in UnicastPrefixV4 filter l.peer_ip == "10.2.2.3" return { UnicastPrefixV4: l, LSLink: (for s in LSLink filter s.remote_link_ip == "10.2.2.3"return s)}
        
for u in UnicastPrefixV4 for l in LSLink filter u.peer_ip == l.remote_link_ip filter u.peer_ip == "10.71.1.1"  return {u: u._id, l: l.remote_link_ip}

FOR d IN peer filter d.remote_ip == "10.71.0.1" FOR l in ls_link  filter d.remote_ip == l.remote_link_ip  return d

for d in peer filter d.remote_ip == "10.1.40.3" for l in ls_link  filter d.remote_ip == l.remote_link_ip  return { d, l }

//FOR v, e, p IN 1..10 ANY "ls_node/2_0_0_0000.0000.0019" lsv4_edge 
    //FILTER v._id == "ls_node/2_0_0_0000.0000.0009" 
  //  FILTER v._id == "unicast_prefix_v4/10.71.2.0_24_10.71.0.1"
    //FILTER e.mt_id_tlv.mt_id == null 
    //RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.vertices[*].router_id, p.edges[*]._key])

for u in unicast_prefix_v4 return u._key

for v, e in outbound 'ls_node/2_0_0_0000.0000.0002' GRAPH ls_node return [v._key, e._key]

//for d in ls_link filter d.mt_id_tlv.mt_id != 2 return d._key


//for v, e in outbound shortest_path 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge filter e.mt_id != 2 return e

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' ls_node_edge filter e.mt_id != 2 return {node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn  } 

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' ls_node_edge OPTIONS {weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency }

//for l in ls_link filter l.protocol_id == 7 && l.peer_asn == 100000 && l.remote_link_ip == "10.73.0.1" return { epe_sid: l.peer_node_sid.prefix_sid } 

//for l in ls_link filter l.protocol_id == 7 && l.peer_asn == 100000 && l.remote_link_ip == "10.73.0.1" return { epe_sid: l.peer_node_sid.sid } 
//for l in ls_link filter l.protocol_id == 7 return l.peer_node_sid.sid//&&  l.remote_link_ip == "10.73.0.1" return l
//for l in ls_link filter l._key == "7_0_0_46489_10.0.0.43_10.73.0.0_10.0.0.73_10.73.0.1" return l
//for l in ls_link filter l.protocol_id == 7 return [l._key, l.remote_link_ip, l.peer_node_sid.sid]


//for l in lsv4_edge return [l._key, l.link_latency]
//FOR p in lsv4_edge filter p._key == "2_0_0_0_0000.0000.0004_10.1.1.65_0000.0000.0007_10.1.1.64" UPDATE p with { link_latency: 50 } in lsv4_edge 

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' lsv4_edge filter e.mt_id != 2 return {node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn  } 

for l in l3vpn_v4_prefix filter l.base_attrs.ext_community_list like "%100:100%" return l//.base_attrs.ext_community_list
//for l in l3vpn_v4 SEARCH  l.base_attrs.ext_community_list == "rt=100:100" filter l.base_attrs.local_pref != null return l//for d in l3vpn_v4 filter d.nexthop == "10.0.0.9" return d._key//filter l.nexthop == "10.0.0.9" return l

//for l in ls_node filter l.igp_router_id == "0000.0000.0021" return l._id
//for l in unicast_prefix_v4 filter l.prefix =="10.10.21.0" return l._id

//for l in ls_node for u in unicast_prefix_v4 filter u._key == "10.10.3.0_24_10.0.0.3" filter l.igp_router_id == "0000.0000.0003" INSERT { _from: l._id, _to: u._id, _key: "10.10.3.0_24_10.0.0.3" } INTO lsv4_edge
 
//for l in lsv4_edge  filter l._key == "10.10.3.0_24_10.0.0.3" return l

//for l in lsv4_edge filter l._key == "10.10.3.0_24_10.0.0.3" UPDATE l with { prefix: "10.10.3.0", prefix_len: 24, nexthop: "10.0.0.3", labels: 24031 } in lsv4_edge

//FOR v, e, p IN 1..16 OUTBOUND 'ls_node/2_0_0_0000.0000.0019' lsv4_edge OPTIONS {uniqueVertices: "path", bfs: true} FILTER v._id == 'unicast_prefix_v4/10.10.21.0_24_10.0.0.21' RETURN p.edges[*].remote_igp_id//._to 


//for l in ls_srv6_sid filter l.igp_router_id == "0000.0000.0018" for m in ls_srv6_sid filter m.igp_router_id == "0000.0000.0017" for n in ls_srv6_sid filter n.igp_router_id == "0000.0000.0016" for o in ls_srv6_sid filter o.igp_router_id == "0000.0000.0021" return [l.srv6_sid, m.srv6_sid, n.srv6_sid, o.srv6_sid]


//for l in ls_node insert l in ls_node_meta options { ignoreErrors: true }

//for l in ls_node filter l.igp_router_id == "0000.0000.0001" return l

//for l in ls_prefix filter l.prefix_attr_tlvs.ls_prefix_sid != null return l

//for l in RT_L3VPNV4 return l

//for l in lsv4_edge return l

//for l in lsv4_edge filter l.protocol_id == 7 return l

//for l in lsv4_edge filter l._to like "%0009%" return l._key

//for l in lsv4_edge UPDATE l with { link_latency: 10 } in lsv4_edge

//for l in lsv4_edge filter l._key =="10.71.4.0_23_10.72.0.1" UPDATE l with { link_latency: 10 } in lsv4_edge

//for l in lsv4_edge filter l._key == "13400517" UPDATE l with { link_latency: 5 } in lsv4_edge

//for l in lsv4_edge filter l._key == "2_0_0_0_0000.0000.0018_10.1.1.48_0000.0000.0022_10.1.1.49" UPDATE l with { link_latency: 90 } in lsv4_edge

//for l in lsv4_edge filter l._key == "2_0_0_0_0000.0000.0017_10.1.1.45_0000.0000.0016_10.1.1.44" UPDATE l with { link_latency: 90 } in lsv4_edge

//for l in lsv4_edge filter l._key == "2_0_0_0_0000.0000.0020_10.1.1.7_0000.0000.0009_10.1.1.6" UPDATE l with { link_latency: 5 } in lsv4_edge

//for l in lsv4_edge filter l._key like "%0019%" return { key: l._key, latency: l.link_latency }

//for l in lsv4_edge return { key: l._key, latency: l.link_latency }

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0019' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' lsv4_edge OPTIONS {weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency }

//for l in lsv4_edge filter l._from == "ls_prefix/2_0_0_0_0_10.0.0.7_32_0000.0000.0007" return l
//for l in l3vpn_v4_prefix_edge return l._from
//for l in ls_srv6_sid_edge return l
//for l in ls_prefix return l
//for l in LSNode_Edge return l 
//FOR d IN peer filter d.remote_bgp_id == "10.0.0.71" filter d.remote_ip == "10.71.0.1" return d
//for l in ls_node_edge return l
//for l in ls_node_edge return l//filter l.protocol_id == 7 return l//l.link_latency == 50 return l._key
//for l in L3Underlay_Edge return l

//FOR n in LSPrefix FILTER n.prefix == "10.0.0.6" RETURN n.prefix_attr_tlvs.ls_prefix_sid[*].prefix_sid

//for p in unicast_prefix_v4 filter p._key == "10.71.8.0_22_10.71.0.1" return p

//for l in epe_link return { key: l._key, latency: l.link_latency }

//FOR p in epe_link FILTER p._key == "7_0_0_100000_10.0.0.7_10.71.1.0_10.0.0.71_10.71.1.1" UPDATE p with { link_latency: 60 } in epe_link

//FOR p in epe_link UPDATE p with { link_latency: 20 } in epe_link

//for l in UnicastPrefixV4
  //  filter l.peer_ip == "10.2.2.3" 
    //return { UnicastPrefixV4: l,
      //  LSLink: (for s in LSLink
        //    filter s.remote_link_ip == "10.2.2.3"
          //  return s)
    //    }
        
//for u in UnicastPrefixV4 for l in LSLink filter u.peer_ip == l.remote_link_ip filter u.peer_ip == "10.71.1.1"  return {u: u._id, l: l.remote_link_ip}

//FOR d IN peer filter d.remote_ip == "10.71.0.1" FOR l in ls_link  filter d.remote_ip == l.remote_link_ip  return d
    
//for l in LSNode_Edge return l//filter l._from == "Peer/10.0.0.72_10.72.0.1" return l

//FOR d IN Peer filter d.remote_ip == "10.71.1.1" FOR l in LSLink  filter d.remote_ip == l.remote_link_ip  return { d, l }

//for l in UnicastPrefixV4 //filter l.peer_ip == "10.71.1.1" return l
//for s in LSLink filter l.peer_ip == s.remote_link_ip
//return {l, s}

//for l in UnicastPrefixV4 filter l._key == "10.0.0.35_32_10.2.2.3" return l

//FOR d IN UnicastPrefixV4 FOR l in LSLink  filter d.peer_ip == l.remote_link_ip  filter d.peer_ip == "10.72.0.1" return d._key

//for d in LSLink for l in UnicastPrefixV4 filter l.prefix == "10.0.0.35" filter d.remote_link_ip == l.peer_ip return d._key
 
//for l in LSLink filter l.protocol_id ==7 return l 

//FOR d IN LSNode filter d.router_id == "10.0.0.7" filter d.domain_id == 0 return d
    
//RETURN LENGTH(FOR v IN OUTBOUND SHORTEST_PATH 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge Return v)

//FOR v, e, p IN 1..6 OUTBOUND 'ls_node/2_0_0_0000.0000.0019' lsv4_edge FILTER v._id == 'unicast_prefix_v4/10.10.21.0_24_10.0.0.21' RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.edges[*].epe_peer])

//FOR v, e, p IN 1..10 ANY "ls_node/2_0_0_0000.0000.0019" lsv4_edge 
    //FILTER v._id == "ls_node/2_0_0_0000.0000.0009" 
  //  FILTER v._id == "unicast_prefix_v4/10.71.2.0_24_10.71.0.1"
    //FILTER e.mt_id_tlv.mt_id == null 
    //RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.vertices[*].router_id, p.edges[*]._key])

//for u in unicast_prefix_v4 return u._key
//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' ls_node_edge return [v._key, e._key]

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' lsv4_edge return [v._key, e._key]

//for v, e in outbound 'ls_node/2_0_0_0000.0000.0002' GRAPH ls_node return [v._key, e._key]

//for d in ls_link filter d.mt_id_tlv.mt_id != 2 return d._key


//FOR p in LSNode_Edge filter p._key == "2_0_0_0_0000.0000.0004_10.1.1.21_0000.0000.0003_10.1.1.20" UPDATE p with { link_latency: 20 } in LSNode_Edge

//for l in LSNode_Edge return l //filter l._key == "2_0_0_0_0000.0000.0004_10.1.1.65_0000.0000.0007_10.1.1.64" return l

//for v, e in outbound shortest_path 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge filter e.mt_id != 2 return e

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' ls_node_edge filter e.mt_id != 2 return {node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn  } 

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' ls_node_edge OPTIONS {weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency }

//for l in ls_link filter l.protocol_id == 7 && l.peer_asn == 100000 && l.remote_link_ip == "10.73.0.1" return { epe_sid: l.peer_node_sid.prefix_sid } 

//for l in ls_link filter l.protocol_id == 7 && l.peer_asn == 100000 && l.remote_link_ip == "10.73.0.1" return { epe_sid: l.peer_node_sid.sid } 
//for l in ls_link filter l.protocol_id == 7 return l.peer_node_sid.sid//&&  l.remote_link_ip == "10.73.0.1" return l
//for l in ls_link filter l._key == "7_0_0_46489_10.0.0.43_10.73.0.0_10.0.0.73_10.73.0.1" return l
//for l in ls_link filter l.protocol_id == 7 return [l._key, l.remote_link_ip, l.peer_node_sid.sid]


//for l in lsv4_edge return [l._key, l.link_latency]
//FOR p in lsv4_edge filter p._key == "2_0_0_0_0000.0000.0004_10.1.1.65_0000.0000.0007_10.1.1.64" UPDATE p with { link_latency: 50 } in lsv4_edge 

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' lsv4_edge filter e.mt_id != 2 return {node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn  } 

//for l in l3vpn_v4_prefix return l.base_attrs.ext_community_list
//for l in l3vpn_v4 SEARCH  l.base_attrs.ext_community_list == "rt=100:100" filter l.base_attrs.local_pref != null return l//for d in l3vpn_v4 filter d.nexthop == "10.0.0.9" return d._key//filter l.nexthop == "10.0.0.9" return l

//for l in ls_node filter l.igp_router_id == "0000.0000.0021" return l._id
//for l in unicast_prefix_v4 filter l.prefix =="10.10.21.0" return l._id

//for l in ls_node for u in unicast_prefix_v4 filter u._key == "10.10.3.0_24_10.0.0.3" filter l.igp_router_id == "0000.0000.0003" INSERT { _from: l._id, _to: u._id, _key: "10.10.3.0_24_10.0.0.3" } INTO lsv4_edge
 
//for l in lsv4_edge  filter l._key == "10.10.3.0_24_10.0.0.3" return l

//for l in lsv4_edge filter l._key == "10.10.3.0_24_10.0.0.3" UPDATE l with { prefix: "10.10.3.0", prefix_len: 24, nexthop: "10.0.0.3", labels: 24031 } in lsv4_edge

FOR v, e, p IN 1..16 OUTBOUND 'ls_node/2_0_0_0000.0000.0019' lsv4_edge 
    OPTIONS {uniqueVertices: "path", bfs: true}
    FILTER v._id == 'unicast_prefix_v4/10.10.21.0_24_10.0.0.21' 
    FILTER v.
    RETURN p.edges[*]._to //CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.edges[*].epe_peer])



              
                


//for l in ipv4_edge return l._to
//for l in lsv4_edge return l
//for l in lsv4_edge return l filter l.protocol_id == 7 filter l.local_link_ip LIKE "%:%" return l
//for l in lsv4_edge filter l.protocol_id == 7 return l._key
//for l in lsv4_edge filter l._key == "2_0_0_0_0000.0000.0010_0.0.0.20_0000.0000.0014_0.0.0.18" return l

//for l in lsv4_edge filter l.prefix like "128.107.20.0" return { To: l._to, latency: l.link_latency }

//for l in lsv4_edge filter l.protocol_id == 2 or l.origin_as == 11404 return l

//for l in unicast_prefix_v4  COLLECT WITH COUNT INTO length RETURN { v4: length }
//for l in unicast_prefix_v4 filter l.nexthop =="198.62.154.19" COLLECT WITH COUNT INTO length RETURN {v4: length}

//for l in lsv4_edge filter l.prefix == "128.107.20.0" return l//update l with { link_latency: 25 } in lsv4_edge 

//for l in lsv4_edge COLLECT WITH COUNT INTO length RETURN length

//for l in lsv4_edge filter l.prefix LIKE "128.107.0%" return l

//for l in ls_link filter l.protocol_id == 7 return l

//for l in ls_node_edge return l

//FOR p in lsv4_edge filter p.protocol_id == 7 UPDATE p with { link_latency: 10 } in lsv4_edge
//FOR p in lsv4_edge filter p._key == "7_0_0_65000_10.0.0.14_198.62.154.18_10.0.0.15_198.62.154.19" UPDATE p with { link_latency: 25 } in lsv4_edge

//for l in lsv4_edge return [l._key, l.link_latency]
//FOR p in lsv4_edge filter p._key == "2_0_0_0_0000.0000.0010_10.1.1.0_0000.0000.0014_10.1.1.1" UPDATE p with { link_latency: 50 } in lsv4_edge 


//for v, e in outbound shortest_path "ls_node/2_0_0_0000.0000.0008" TO "unicast_prefix_v4/128.107.20.0_23_198.62.154.19" lsv4_edge OPTIONS { weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn, adj_sid: e.peer_node_sid.sid  }

//for v, e in outbound shortest_path "ls_node/2_0_0_0000.0000.0008" TO LIKE("unicast_prefix_v4/128.107.20.0_23_198.62.154.%") lsv4_edge OPTIONS { weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn, adj_sid: e.peer_node_sid.sid  }

//for v, e in outbound shortest_path "ls_node/2_0_0_0000.0000.0008" TO "unicast_prefix_v4/103.107.187.0_24_198.62.154.1" lsv4_edge return {vertex: v._key, edge: e._key }

//for l in ls_node filter l._key == "10.0.0.1_198.62.154.1"  
//for p in peer filter p._key == "10.0.0.1_198.62.154.1" return p


//for l in ls_node filter l._key == "2_0_0_0000.0000.0010" return l.router_id


//for l in ls_prefix filter l.prefix == "10.0.0.12" filter l.protocol_id == 2  RETURN  l.prefix_attr_tlvs.ls_prefix_sid[0].prefix_sid

//for n in ls_node filter n._key == "2_0_0_0000.0000.0012" RETURN { prefix: n.router_id, srgb_start: n.ls_sr_capabilities.sr_capability_subtlv[*].sid }


//for l in ls_node_edge return l

//for l in lsv4_edge return l

//for l in lsv4_edge filter l.origin_as == 11404 return l

//for l in lsv4_edge filter l.protocol_id == 2 or l.origin_as == 11404 return l

//for l in lsv4_edge filter l.prefix LIKE "128.107.0%" return l

//for l in ls_link filter l.protocol_id == 7 return l

//for l in ls_node_edge return l._key

//FOR p in lsv4_edge UPDATE p with { link_latency: 10 } in lsv4_edge

//for l in ls_node_edge return l//filter l.protocol_id == 7 return l//l.link_latency == 50 return l._key
//for l in lsv4_edge return l
//for l in l3vpn_v4_prefix_edge return l._from
//for l in ls_srv6_sid_edge return l
//for l in ls_prefix return l
//for l in LSNode_Edge return l 
//FOR d IN peer return d

//for l in L3Underlay_Edge return l

//FOR n in LSPrefix FILTER n.prefix == "10.0.0.6" RETURN n.prefix_attr_tlvs.ls_prefix_sid[*].prefix_sid

//for l in LSLink filter l.protocol_id == 7 return l
//for p in UnicastPrefixV4 filter p.peer_ip == "10.71.0.1" return p

//for l in epe_link return { key: l._key, latency: l.link_latency }

//FOR p in epe_link FILTER p._key == "7_0_0_100000_10.0.0.7_10.71.1.0_10.0.0.71_10.71.1.1" UPDATE p with { link_latency: 60 } in epe_link

//FOR p in epe_link UPDATE p with { link_latency: 20 } in epe_link

//for l in UnicastPrefixV4
  //  filter l.peer_ip == "10.2.2.3" 
    //return { UnicastPrefixV4: l,
      //  LSLink: (for s in LSLink
        //    filter s.remote_link_ip == "10.2.2.3"
          //  return s)
    //    }
        
//for u in UnicastPrefixV4 for l in LSLink filter u.peer_ip == l.remote_link_ip filter u.peer_ip == "10.71.1.1"  return {u: u._id, l: l.remote_link_ip}

//FOR d IN peer filter d.remote_ip == "10.71.0.1" FOR l in ls_link  filter d.remote_ip == l.remote_link_ip  return d
    
//for l in LSNode_Edge return l//filter l._from == "Peer/10.0.0.72_10.72.0.1" return l

//FOR d IN Peer filter d.remote_ip == "10.71.1.1" FOR l in LSLink  filter d.remote_ip == l.remote_link_ip  return { d, l }

//for l in UnicastPrefixV4 //filter l.peer_ip == "10.71.1.1" return l
//for s in LSLink filter l.peer_ip == s.remote_link_ip
//return {l, s}

//for l in UnicastPrefixV4 filter l._key == "10.0.0.35_32_10.2.2.3" return l

//FOR d IN UnicastPrefixV4 FOR l in LSLink  filter d.peer_ip == l.remote_link_ip  filter d.peer_ip == "10.72.0.1" return d._key

//for d in LSLink for l in UnicastPrefixV4 filter l.prefix == "10.0.0.35" filter d.remote_link_ip == l.peer_ip return d._key
 
//for l in LSLink filter l.protocol_id ==7 return l 

//FOR d IN LSNode filter d.router_id == "10.0.0.7" filter d.domain_id == 0 return d
    
//RETURN LENGTH(FOR v IN OUTBOUND SHORTEST_PATH 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge Return v)

//FOR v, e, p IN 4..5 ANY 'LSNode/2_0_0_0000.0000.0001' LSNode_Edge FILTER v._id == 'UnicastPrefixV4/72.72.1.0_24_10.71.0.1' RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.edges[*].epe_peer])

//FOR v, e, p IN 1..3 ANY "LSNode/2_0_0_0000.0000.0004" LSNode_Edge 
//    FILTER v._id == "LSNode/2_0_0_0000.0000.0002" 
//    FILTER e.mt_id_tlv.mt_id == null 
//    RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.vertices[*].router_id, p.edges[*]._key])

//for u in unicast_prefix_v4 return u._key
//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' ls_node_edge return [v._key, e._key]

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' lsv4_edge return [v._key, e._key]

//for v, e in outbound 'ls_node/2_0_0_0000.0000.0002' GRAPH ls_node return [v._key, e._key]

//for d in ls_link filter d.mt_id_tlv.mt_id != 2 return d._key


//FOR p in LSNode_Edge filter p._key == "2_0_0_0_0000.0000.0004_10.1.1.21_0000.0000.0003_10.1.1.20" UPDATE p with { link_latency: 20 } in LSNode_Edge

//for l in LSNode_Edge return l //filter l._key == "2_0_0_0_0000.0000.0004_10.1.1.65_0000.0000.0007_10.1.1.64" return l

//for v, e in outbound shortest_path 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge filter e.mt_id != 2 return e

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' ls_node_edge filter e.mt_id != 2 return {node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn  } 

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' ls_node_edge OPTIONS {weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency }

//for l in ls_link filter l.protocol_id == 7 && l.peer_asn == 100000 && l.remote_link_ip == "10.73.0.1" return { epe_sid: l.peer_node_sid.prefix_sid } 

//for l in ls_link filter l.protocol_id == 7 && l.peer_asn == 100000 && l.remote_link_ip == "10.73.0.1" return { epe_sid: l.peer_node_sid.sid } 
//for l in ls_link filter l.protocol_id == 7 return l.peer_node_sid.sid//&&  l.remote_link_ip == "10.73.0.1" return l
//for l in ls_link filter l._key == "7_0_0_46489_10.0.0.43_10.73.0.0_10.0.0.73_10.73.0.1" return l
//for l in ls_link filter l.protocol_id == 7 return [l._key, l.remote_link_ip, l.peer_node_sid.sid]

//FOR p in lsv4_edge UPDATE p with { link_latency: 10 } in lsv4_edge
//for l in lsv4_edge return [l._key, l.link_latency]
//FOR p in lsv4_edge filter p._key == "2_0_0_0_0000.0000.0004_10.1.1.65_0000.0000.0007_10.1.1.64" UPDATE p with { link_latency: 50 } in lsv4_edge 

//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' lsv4_edge filter e.mt_id != 2 return {node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn  } 


//for v, e in outbound shortest_path 'ls_node/2_0_0_0000.0000.0017' TO 'unicast_prefix_v4/10.71.2.0_24_10.71.0.1' lsv4_edge OPTIONS {weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency }


//for l in l3vpn_v4_prefix return l.base_attrs.ext_community_list
for l in l3vpn_v4 SEARCH  l.base_attrs.ext_community_list == "rt=100:100" filter l.base_attrs.local_pref != null return l//for d in l3vpn_v4 filter d.nexthop == "10.0.0.9" return d._key//filter l.nexthop == "10.0.0.9" return l


//for l in topology_nodes_Edge return l

//for l in LSNode_Edge filter l.link_latency == 50 return l._key

//for l in L3Underlay_Edge return l

//FOR n in LSPrefix FILTER n.prefix == "10.0.0.6" RETURN n.prefix_attr_tlvs.ls_prefix_sid[*].prefix_sid

//for l in LSLink filter l.protocol_id == 7 return l
//for p in UnicastPrefixV4 filter p.peer_ip == "10.71.0.1" return p

//for l in EPELink return { key: l._key, latency: l.link_latency }

//FOR p in EPELink FILTER p._key == "7_0_0_100000_10.0.0.7_10.71.1.0_10.0.0.71_10.71.1.1" UPDATE p with { link_latency: 20 } in EPELink


//for l in UnicastPrefixV4
  //  filter l.peer_ip == "10.2.2.3" 
    //return { UnicastPrefixV4: l,
      //  LSLink: (for s in LSLink
        //    filter s.remote_link_ip == "10.2.2.3"
          //  return s)
    //    }
        
//for u in UnicastPrefixV4 for l in LSLink filter u.peer_ip == l.remote_link_ip filter u.peer_ip == "10.71.1.1"  return {u: u._id, l: l.remote_link_ip}
    
//for l in LSNode_Edge return l//filter l._from == "Peer/10.0.0.72_10.72.0.1" return l

//FOR d IN Peer filter d.remote_ip == "10.71.1.1" FOR l in LSLink  filter d.remote_ip == l.remote_link_ip  return { d, l }

//for l in UnicastPrefixV4 //filter l.peer_ip == "10.71.1.1" return l
//for s in LSLink filter l.peer_ip == s.remote_link_ip
//return {l, s}

//for l in UnicastPrefixV4 filter l._key == "10.0.0.35_32_10.2.2.3" return l

//FOR d IN UnicastPrefixV4 FOR l in LSLink  filter d.peer_ip == l.remote_link_ip  filter d.peer_ip == "10.72.0.1" return d._key

//for d in LSLink for l in UnicastPrefixV4 filter l.prefix == "10.0.0.35" filter d.remote_link_ip == l.peer_ip return d._key
 
//for l in LSLink filter l.protocol_id ==7 return l 

//FOR d IN LSNode filter d.router_id == "10.0.0.7" filter d.domain_id == 0 return d
    
//RETURN LENGTH(FOR v IN OUTBOUND SHORTEST_PATH 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge Return v)

//FOR v, e, p IN 4..5 ANY 'LSNode/2_0_0_0000.0000.0001' LSNode_Edge FILTER v._id == 'UnicastPrefixV4/72.72.1.0_24_10.71.0.1' RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.edges[*].epe_peer])

//FOR v, e, p IN 1..3 ANY "LSNode/2_0_0_0000.0000.0004" LSNode_Edge 
//    FILTER v._id == "LSNode/2_0_0_0000.0000.0002" 
//    FILTER e.mt_id_tlv.mt_id == null 
//    RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.vertices[*].router_id, p.edges[*]._key])

//for v, e in outbound shortest_path 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge return [v._key, e._key]

for v, e in outbound 'LSNode/2_0_0_0000.0000.0017' GRAPH Peer return [v._key, e._key]



//FOR p in LSNode_Edge filter p._key == "2_0_0_0_0000.0000.0004_10.1.1.21_0000.0000.0003_10.1.1.20" UPDATE p with { link_latency: 20 } in LSNode_Edge

//for l in LSNode_Edge return l //filter l._key == "2_0_0_0_0000.0000.0004_10.1.1.65_0000.0000.0007_10.1.1.64" return l

//for v, e in outbound shortest_path 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge filter e.mt_id != 2 return e

//for v, e in outbound shortest_path 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge filter e.mt_id != 2 return {node: v._key, link: e._key, latency: e.link_latency, asn: v.asn, local_asn: v.local_asn, remote_asn: v.remote_asn  } 

//for v, e in outbound shortest_path 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1' LSNode_Edge OPTIONS {weightAttribute: 'link_latency' } filter e.mt_id != 2 return { node: v._key, link: e._key, latency: e.link_latency }





//FOR p IN OUTBOUND K_SHORTEST_PATHS 'LSNode/2_0_0_0000.0000.0017' TO 'UnicastPrefixV4/10.71.2.0_24_10.71.0.1'
  //LSNode_Edge
    //  LIMIT 4
      //RETURN {
        //  nodes: p.vertices[*]._key,
          //latencies: p.edges[*].link_latency,
          //totalLatency: SUM(p.edges[*].link_latency)
      //}
      



//for l in topology_nodes_Edge return l

//for l in LSNode_Edge return l

//for l in LSNode return l

//for l in UnicastPrefixV4 return l //filter l.router_ip == "10.71.1.0" return l

//for l in L3Underlay_Edge return l

//FOR n in LSPrefix FILTER n.prefix == "10.0.0.6" RETURN n.prefix_attr_tlvs.ls_prefix_sid[*].prefix_sid

//for l in LSLink filter l.protocol_id == 7 return l
//for p in UnicastPrefixV4 filter p.peer_ip == "10.71.0.1" return p

//for l in EPELink return { key: l._key, latency: l.link_latency }

//FOR p in EPELink FILTER p._key == "7_0_0_100000_10.0.0.7_10.71.1.0_10.0.0.71_10.71.1.1" UPDATE p with { link_latency: 20 } in EPELink

//for l in UnicastPrefixV4
  //  filter l.peer_ip == "10.2.2.3" 
    //return { UnicastPrefixV4: l,
      //  LSLink: (for s in LSLink
        //    filter s.remote_link_ip == "10.2.2.3"
          //  return s)
    //    }
        
//for u in UnicastPrefixV4 for l in LSLink filter u.peer_ip == l.remote_link_ip filter u.peer_ip == "10.71.1.1"  return {u: u._id, l: l.remote_link_ip}
    
//for l in LSNode_Edge return l//filter l._from == "Peer/10.0.0.72_10.72.0.1" return l

//FOR d IN Peer filter d.remote_ip == "10.71.1.1" FOR l in LSLink  filter d.remote_ip == l.remote_link_ip  return { d, l }
 
FOR d IN UnicastPrefixV4 FOR l in LSLink  filter d.peer_ip == l.remote_link_ip  filter d.peer_ip == "10.72.0.1" return d._key

Queries
for e in L3VPN_Topology filter e.RD == "100:100" return e

for e in L3VPN_Topology filter e.RD == "101:101" and e.RouterID == e.SrcIP return e

FOR L3VPN_Topology IN L3VPN_Topology
  RETURN L3VPN_Topology

FOR LSLink IN LSLink
  RETURN LSLink

RETURN LENGTH(
FOR v IN OUTBOUND 
SHORTEST_PATH 'LSNode/2_0_0_0000.0000.0007' TO 'LSNode/2_0_0_0000.0000.0019' LSNode_Edge
 Return v
)

Clear a collection:
FOR u IN EPEExternalPrefix
  REMOVE u IN EPEExternalPrefix

RETURN LENGTH(
FOR v IN OUTBOUND 
SHORTEST_PATH 'LSNode/172.31.101.1' TO 'LSNode/172.31.101.6' LS_Topology
 Return v
)


Wow
FOR prefix in LSPrefix
FILTER prefix.mt_id_tlv.mt_id == 2
FILTER LENGTH(prefix.srv6_locator)
SORT prefix.prefix, prefix.igp_router_id, prefix.protocol_id
RETURN { proto: prefix.protocol_id, router: prefix.igp_router_id, loc: prefix.srv6_locator, attrs: prefix.prefix_attr_flags, prefix: CONCAT(prefix.prefix, "/", prefix.prefix_len)}


FOR v, e, p IN 1..6 OUTBOUND "LSNode/172.31.101.1" LS_Topology
     FILTER v._id == "LSNode/172.31.101.6"
       RETURN { "RouterID": p.vertices[*].RouterID, "PrefixSID": p.edges[*].RemotePrefixSID, "AdjSID": p.edges[*].AdjacencySID, "Util": p.edges[*].Out_Octets }

FOR v, e, p IN 3..3 OUTBOUND "LSNode/172.31.101.1" LS_Topology
     FILTER v._id == "LSNode/172.31.101.6"
       RETURN CONCAT_SEPARATOR(" -> ", p.vertices[*].RouterID)

FOR v, e, p IN 1..6 OUTBOUND "LSNode/172.31.101.1" LS_Topology
     FILTER v._id == "LSNode/172.31.101.6"
       RETURN { "RouterID": p.vertices[*].RouterID, "PrefixSID": p.edges[*].RemotePrefixSID, "Via": e.ToInterfaceIP, "Tx_bytes": e.Out_Octets }

FOR v, e IN OUTBOUND SHORTEST_PATH 'LSNode/172.31.101.1' TO 'LSNode/172.31.101.6' LS_Topology
    OPTIONS {weightAttribute: 'PercentUtilOutbound'}
    RETURN {v, e}

FOR v, e, p IN 3..3 OUTBOUND "LSNode/172.31.101.1" LS_Topology
     FILTER v._id == "LSNode/172.31.101.6"
       RETURN { "RouterID": p.vertices[*].RouterID, "Via": p.edges[*].ToInterfaceIP }


Steering queries:

// Full Topology
//for d in LSNode_Edge filter d._key == "2_0_0_0_0000.0000.0018_10.1.1.2_0000.0000.0020_10.1.1.3"  return d
//for d in LSNode_Edge filter d._key == "2_0_0_0_0000.0000.0020_10.1.1.3_0000.0000.0018_10.1.1.2"  return d
//for d in LSNode_Edge return d

//for d in LSPrefix_Edge return d
//for d in LSLink return [d._key, d.igp_router_id]
//for d in Node return d
//FOR d IN LSLink filter d.igp_router_id == null filter d.domain_id == 0 filter d.protocol_id == 7 return d
//for d in UnicastPrefixV6 filter d.prefix == "2001:19d0:600::" return d
//for d in UnicastPrefixV4 filter d.prefix == "1.0.4.0" return d
//for d in UnicastPrefixV6 filter d.prefix == "2001:420:ffff::116" return d

//FOR d IN LSLink filter d.protocol_id != 7 return d._key

//for d in LSPrefix filter d.prefix == "2001:420:ffff:1013::2" return d

Hop Count
//RETURN LENGTH (FOR v IN ANY SHORTEST_PATH "LSNode/2_0_0_0000.0000.0007" TO "LSNode/2_0_0_0000.0000.0019" LSNode_Edge RETURN v)

// All Paths
//FOR v, e, p IN 5..5 ANY "Node/10.2.1.0" LSNode_Edge FILTER v._id == "Node/198.62.154.19" RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.vertices[*].router_id])

//FOR v, e, p IN 4..4 ANY "LSNode/2_0_0_0000.0000.0007" LSNode_Edge FILTER e.mt_id_tlv != null FILTER v._id == "LSNode/2_0_0_0000.0000.0019" RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.vertices[*].router_id, p.edges[*]._key])

//FOR v, e, p IN 1..3 ANY "LSNode/2_0_0_0000.0000.0007" LSNode_Edge 
    //FILTER v._id == "LSNode/2_0_0_0000.0000.0019" 
   // FILTER e.mt_id_tlv.mt_id == null 
    //RETURN CONCAT_SEPARATOR(" -> ", [p.vertices[*]._key, p.vertices[*].router_id, p.edges[*]._key])

//FOR v, e, p IN 1..3 ANY "LSNode/2_0_0_0000.0000.0007" LSNode_Edge 
//    FILTER v._id == "LSNode/2_0_0_0000.0000.0019" 
//    FILTER p.edges[*].mtid ALL == 2
//    RETURN { vertices: p.vertices[*]._key, edges: p.edges[*]._key, mtid: p.edges[*].mtid }
   
FOR p in LSPrefix 
    filter p.igp_router_id == "0000.0000.0009" and p.prefix_metric == null 
    return [p._key, p.igp_router_id, p.prefix, p.ls_prefix_sid]



// Prefix SIDs
//FOR d in LSPrefix filter d.igp_router_id == "0000.0000.0003" filter d.prefix == "10.0.0.3" return [d.ls_prefix_sid]

//FOR d in LSPrefix filter d.igp_router_id == "0000.0000.0018" filter d.prefix == "10.0.0.12" return [d.ls_prefix_sid]


//FOR d IN LSLink filter d.protocol_id == 7 RETURN d
//FOR d in Node filter d.remote_bgp_id == "10.0.0.15" filter d.remote_ip == "198.62.154.17" return d

//for l in LSv4_Topology return l


//RETURN LENGTH(
//FOR v IN OUTBOUND 
//SHORTEST_PATH 'LSNode/0000.0000.0007' to 'LSNode/0000.0000.0020' LSv4_Topology
//  Return v
//  )

//for e in LSv4_Topology filter e._key == "0000.0000.0022_0.0.0.0_6_0000.0000.0003_0.0.0.0_11"
//return e

//for l in LSSRv6SID return l

for p in EPEPrefix filter p.Prefix == "152.89.60.0"
  return p

//for n in LSNode filter n.igp_router_id == "0000.0000.0008" and n.protocol_id == 2
//  return n


//for l in LSNode_Edge return l

//for l in LSNode return l._key

//for l in LSNode return l

//for l in LSSRv6SID_Edge return l

//for l in LSPrefix_Edge return l

//for l in LSLink filter l._key == "7_0_0_64032_10.0.0.32_10.0.32.1_10.0.0.1_10.0.32.0" return l

for l in LSLink filter l.protocol_id == 7 return l

//for l in lldp_nodes_Edge return l

//for l in L3VPNV4_Prefix filter l._key == "10.0.0.7:1_172.16.7.0_32" return l

//for l in LSPrefix filter l.prefix == "10.0.0.7" return l.prefix_attr_tlvs.ls_prefix_sid

//for l in LSNode filter l.router_id == "10.0.0.7" return l.ls_sr_capabilities

//for u in UnicastPrefixV4 filter u.nexthop == "10.73.1.1" return u._key

//for u in UnicastPrefixV4 filter u.nexthop == "10.73.0.1" return {prefix: u.prefix, length: u.prefix_len }

//for l in LSLink filter l.protocol_id == 7 && l.area_id == "46489" return  { key: l._key, local_link_ip: l.local_link_ip, epe_sid: l.peer_node_sid.prefix_sid }

old

Graph traversal
* lowest latency 
    * list the lowest cost path(s)
    * is one of these paths lowest latency? 
        * if so, use prefix-sid
        * if not, use prefix-sid of connecting node on lowest-latency path (ie, node on lowest latency path with lowest cost to destination
* least utilized
    * source SEA, dest NYC
    * is least utilized also one of the lowest cost paths?
        * yes: use prefix-sid of connecting node on least-utilized path + dest prefix-sid
        * no: use prefix-sid of connecting node with lowest cost to dest and on L-U-path + dest prefix-sid 

multiple databases
collections are equivalent to sql tables

Queries
See what interface = what label:
for v in LinkEdgesV4
filter v.Label == "24005"
return v
Shortest latency:
RETURN FLATTEN(
    FOR v,e IN OUTBOUND SHORTEST_PATH 'Routers/10.1.1.0' TO 'Prefixes/10.11.0.0_24'
    GRAPH "topology"
    OPTIONS {weightAttribute: 'Latency', defaultWeight: 60}
    RETURN [e.Label, v.Label]
)
Shortest latency:
FOR v,e IN OUTBOUND SHORTEST_PATH 'Routers/10.1.1.1' TO 'Prefixes/10.11.0.0_24' GRAPH 'topology' OPTIONS {weightAttribute: "Latency", defaultWeight: 1000} return e.Label
Get all prefix edges:
for r in PrefixEdges
filter r.Latency != null
and r._to == "Prefixes/10.11.0.0_24"
return r


Lots of example queries



//for l in LSv4_Topology return l

//RETURN LENGTH(
//FOR v IN OUTBOUND 
//SHORTEST_PATH 'LSNode/2_0__0000.0000.0004' to 'LSNode/2_0__0000.0000.0020' LSv4_Topology
//  Return v
//  )

//for e in LSv4_Topology filter e._key == "0000.0000.0022_0.0.0.0_6_0000.0000.0003_0.0.0.0_11"
//return e

//for l in LSSRv6SID return l

//for p in UnicastPrefix filter p.prefix == "1.0.4.0"
//  return p

//for n in LSNode filter n.igp_router_id == "0000.0000.0008" and n.protocol_id == 2
//  return n

//for l in LSv6_Topology return l
//for l in LSNode return l.srgb_start
//for l in L3VPNPrefix return l
//for l in LSLinkEdge return l._key
//for l in LSLinkEdge filter l.local_igp_id =="0000.0000.0008" return l

//FOR l in LSNode filter l._key == "2_0_0000.0000.0007" return l.srgb_start 

//FOR n in LSNode RETURN n.name



//FOR l in LSv4_Topology return l._key
//FOR l in LSLinkV1 filter l._key not in @lsv4_topology_keys return l._key
//FOR l in LSv4_Topology filter l._key == "2_0_0000.0000.0022_10.1.1.4_0_0000.0000.0019_10.1.1.5_0" RETURN { key: l._key }

//FOR l in LSPrefixV1 filter l.igp_router_id == "0000.0000.0022" and l.prefix_sid != null for z in l.prefix_sid filter z.algo == 0 return {"prefix": l.prefix, "length": l.length, "flags": z.flags, "sid_index":  z.prefix_sid}

//FOR l in LSPrefixV1 filter l.igp_router_id == "0000.0000.0022" return {"prefix": l.prefix, "length": l.length, "flags": l.flags, "sid_index":  l.prefix_sid}

FOR l in LSv4_Topology filter l._key == "2_0_0000.0000.0022_0.0.0.0_6_0000.0000.0003_0.0.0.0_11" return {"local_igp_id": l.local_igp_id}



//for l in LSv4_Topology return l

//RETURN LENGTH(
//FOR v IN OUTBOUND 
//SHORTEST_PATH 'LSNodeDemo/0000.0000.0006' TO 'LSNodeDemo/0000.0000.0019' LSv4_Topology
// Return v
//)

//FOR v, e, p IN 3..3 OUTBOUND "LSNodeDemo/0000.0000.0006" LSv4_Topology
//     FILTER v._id == "LSNodeDemo/0000.0000.0019"
//       RETURN CONCAT_SEPARATOR(" -> ", p.vertices[*].router_id)

FOR v, e IN OUTBOUND SHORTEST_PATH 'LSNodeDemo/0000.0000.0007' TO 'LSNodeDemo/0000.0000.0019' LSv4_Topology
    OPTIONS {weightAttribute: 'Percent_Util_Outbound'}
    FILTER e != null
    RETURN [v.router_id, e.remote_prefix_sid]


//FOR p IN OUTBOUND K_SHORTEST_PATHS 'LSNodeDemo/0000.0000.0007' TO 'LSNodeDemo/0000.0000.0019' LSv4_Topology
//LIMIT 3
//RETURN p

//FOR v, e IN OUTBOUND 
//SHORTEST_PATH 'LSNodeDemo/0000.0000.0006' TO 'LSNodeDemo/0000.0000.0019' LSv4_Topology
// FILTER v.router_id != "10.0.0.9"
// Return [v.router_id, e.remote_prefix_sid]


//FOR v,e IN
//  OUTBOUND SHORTEST_PATH "LSNodeDemo/0000.0000.0006" TO "LSNodeDemo/0000.0000.0019" GRAPH "LSv4"
//  RETURN [v.router_id, e.remote_prefix_sid] 
  
  
//FOR path IN
//  OUTBOUND K_SHORTEST_PATHS "LSNodeDemo/0000.0000.0006" TO "LSNodeDemo/0000.0000.0019" GRAPH "LSv4"
//LIMIT 1
//RETURN path
