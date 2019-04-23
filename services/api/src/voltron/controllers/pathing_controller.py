from ipaddress import ip_address
from voltron import queries


def pathing_epe_bandwidth_get(dst_ip, min_bandwidth=None, peer_preference=None, composite=None):  # noqa: E501
    """Optimize pathing to EPE based on bandwidth.

     # noqa: E501

    :param dst_ip: The destination IP.
    :type dst_ip: str
    :param min_bandwidth: Specification of a minimum allowable bandwidth, otherwise the greatest bandwidth path will be returned.
    :type min_bandwidth: int
    :param peer_preference: Specification of peer preference, either direct or transit.
    :type peer_preference: str
    
    :rtype: SRLabelStack
    """
    try:
        dst_ip = str(ip_address(dst_ip))
    except ValueError:
        return 'Invalid destination IP.', 400
    if min_bandwidth is not None:
        try:
            min_bandwidth = int(min_bandwidth)
        except TypeError:
            return 'Minimum bandwidth must be an integer.', 400
    if peer_preference is not None:
        if peer_preference not in ["direct", "transit"]:
            return 'Invalid peer preference, should be direct or transit', 400
    return queries.pathing_epe_bandwidth_get(dst_ip, min_bandwidth, peer_preference, composite)

def pathing_epe_latency_all_get(src_ip, src_transport_ip):  # noqa: E501
    """Returns all latencies for EPE use-case.
     # noqa: E501
    :param src_ip: The source IP.
    :type src_ip: str
    :param src_transport_ip: The upstream or gateway IP that identifies traversal through the network beyond the host.
    :type src_transport_ip: str
    :rtype: SRLabelStack
    """
    try:
        src_ip = str(ip_address(src_ip))
    except ValueError:
        return 'Invalid source IP.', 400
    try:
        src_transport_ip = str(ip_address(src_transport_ip))
    except ValueError:
        return 'Invalid source transport IP.', 400
    return queries.pathing_epe_latency_all_get(src_ip, src_transport_ip)


def pathing_epe_latency_get(src_ip, src_transport_ip, dst_ip, max_latency=None, peer_preference=None, composite=None):  # noqa: E501
    """Optimize pathing to EPE based on latency.

     # noqa: E501

    :param src_ip: The source IP.
    :type src_ip: str
    :param src_transport_ip: The upstream or gateway IP that identifies traversal through the network beyond the host.
    :type src_transport_ip: str
    :param dst_ip: The destination IP.
    :type dst_ip: str
    :param max_latency: Specification of a maximum allowable latency, otherwise the lowest latency path will be returned.
    :type max_latency: int
    :param peer_preference: Specification of peer preference, either direct or transit.
    :type peer_preference: str

    :rtype: SRLabelStack
    """
    try:
        src_ip = str(ip_address(src_ip))
    except ValueError:
        return 'Invalid source IP.', 400
    try:
        src_transport_ip = str(ip_address(src_transport_ip))
    except ValueError:
        return 'Invalid source transport IP.', 400
    try:
        dst_ip = str(ip_address(dst_ip))
    except ValueError:
        return 'Invalid destination IP.', 400
    if max_latency is not None:
        try:
            max_latency = int(max_latency)
        except TypeError:
            return 'Maximum latency must be an integer.', 400
    if peer_preference is not None:
        if peer_preference not in ["direct", "transit"]:
            return 'Invalid peer preference, should be direct or transit', 400
    return queries.pathing_epe_latency_get(src_ip, src_transport_ip, dst_ip, max_latency, peer_preference, composite)


def pathing_epe_lossless_get(dst_ip, max_loss=None, peer_preference=None, composite=None):  # noqa: E501
    """Optimize pathing to EPE based on loss-related statistics.

     # noqa: E501

    :param dst_ip: The destination IP.
    :type dst_ip: str
    :param max_loss: Specification of the maximum allowable loss, otherwise the lowest loss path will be returned.
    :type max_loss: int
    :param peer_preference: Specification of peer preference, either direct or transit.
    :type peer_preference: str

    :rtype: SRLabelStack
    """
    try:
        dst_ip = str(ip_address(dst_ip))
    except ValueError:
        return 'Invalid destination IP.', 400
    if max_loss is not None:
        try:
            max_loss = int(max_loss)
        except TypeError:
            return 'Maximum loss must be an integer.', 400
    if peer_preference is not None:
        if peer_preference not in ["direct", "transit"]:
            return 'Invalid peer preference, should be direct or transit', 400
    return queries.pathing_epe_lossless_get(dst_ip, max_loss, peer_preference, composite)


def pathing_epe_utilization_get(dst_ip, max_utilization=None, peer_preference=None, composite=None):  # noqa: E501
    """Optimize pathing to EPE based on utilization percentages.

     # noqa: E501

    :param dst_ip: The destination IP.
    :type dst_ip: str
    :param max_utilization: Specification of the maximum allowable utilization percentage, otherwise the least utilized path will be returned.
    :type max_utilization: int
    :param peer_preference: Specification of peer preference, either direct or transit.
    :type peer_preference: str

    :rtype: SRLabelStack
    """
    try:
        dst_ip = str(ip_address(dst_ip))
    except ValueError:
        return 'Invalid destination IP.', 400
    if max_utilization is not None:
        try:
            max_utilization = int(max_utilization)
        except TypeError:
            return 'Maximum utilization must be an integer.', 400
    if peer_preference is not None:
        if peer_preference not in ["direct", "transit"]:
            return 'Invalid peer preference, should be direct or transit', 400
    return queries.pathing_epe_utilization_get(dst_ip, max_utilization, peer_preference, composite)
