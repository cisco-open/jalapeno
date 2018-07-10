### router.go defines how we handle "router" openbmp messages. 
currently, router.go parses each "router" openbmp message for three fields:
	* BGPID (our key)
	* Name
	* RouterIP

it then creates a new router object with those fields, and then upserts 
that router object as a document into the Router collection in Arango.

this process occurs for all "router" openbmp messages.

note: the upserted router documents in the "Router" collection will not have all their fields 
(defined in the router document schema in ../database/router.go) filled out through this handling. 
the rest of the fields will be filled out by the handling of other openbmp messages (i.e. peer).


### peer.go defines how we handle "peer" openbmp messages.
each "peer" openbmp message has pertinent fields for two router documents (local and peer),
and for the link-edge documents between them (one each way).

currently, peer.go parses each "peer" openbmp message for these fields:
        * local_bgp_id (used as local router document's BGPID (our key))
        * local_asn (local router document's ASN, also discerns local router document's "isLocal" field)
        * remote_bgp_id (used as peer router document's BGPID (our key))
        * remote_asn (peer router document's ASN, also discerns peer router document's "isLocal" field)
        * local_ip (used as "FromIP" field one way, and "ToIP" in the reverse way, for the two LinkEdge documents)
        * remote_ip (used as "FromIP" field one way, and "ToIP" in the reverse way, for the two LinkEdge documents)

peer.go upserts the ASN/isLocal values for the local router document, upserts the ASN/isLocal values for the peer
router document, and upserts the two link-edge documents between them.

this process occurs for all "peer" openbmp messages.
