#! /usr/bin/env python
"""This class upserts performance metrics into the LSv6_Topology collection in ArangoDB"""
from configs import arangoconfig
from util import connections

def update_lsv6_performance(arango_client, key, in_unicast_pkts, out_unicast_pkts,
                          in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                          in_discards, out_discards, in_errors, out_errors, in_octets, out_octets,
                          speed, percent_util_inbound, percent_util_outbound):
    """Update specified collection document with new performance data (document specified by key)."""
    print("Updating existing LSv6_Topology record " + key + " with performance metrics")
    aql = """FOR p in LSv6_Topology
        FILTER p._key == @key
        UPDATE p with { In_Unicast_Pkts: @in_unicast_pkts, Out_Unicast_Pkts: @out_unicast_pkts,
                        In_Multicast_Pkts: @in_multicast_pkts, Out_Multicast_Pkts: @out_multicast_pkts,
                        In_Broadcast_Pkts: @in_broadcast_pkts, Out_Broadcast_Pkts: @out_broadcast_pkts,
                        In_Discards: @in_discards, Out_Discards: @out_discards,
                        In_Errors: @in_errors, Out_Errors: @out_errors,
                        In_Octets: @in_octets, Out_Octets: @out_octets,
                        Speed: @speed, Percent_Util_Inbound: @percent_util_inbound,
                        Percent_Util_Outbound: @percent_util_outbound } in LSv6_Topology"""
    bindVars = {'key': key,
                'in_unicast_pkts': in_unicast_pkts, 'out_unicast_pkts': out_unicast_pkts,
                'in_multicast_pkts': in_multicast_pkts, 'out_multicast_pkts': out_multicast_pkts,
                'in_broadcast_pkts': in_broadcast_pkts, 'out_broadcast_pkts': out_broadcast_pkts,
                'in_discards': in_discards, 'out_discards': out_discards,
                'in_errors': in_errors, 'out_errors': out_errors,
                'in_octets': in_octets, 'out_octets': out_octets,
                'speed': speed, 'percent_util_inbound': percent_util_inbound,
                'percent_util_outbound': percent_util_outbound}
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)


def update_lsv6_interface_name(arango_client, key, interface_name):
    print("Updating existing LSv6_Topology record " + key + " with source interface name: " + interface_name)
    aql = """FOR p in LSv6_Topology
        FILTER p._key == @key
        UPDATE p with { FromInterfaceName: @source_interface_name } in LSv6_Topology"""
    bindVars = {'key': key, 'source_interface_name': interface_name}
    arango_client.AQLQuery(aql, rawResults=True, bindVars=bindVars)

