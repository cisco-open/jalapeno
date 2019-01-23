import connexion
import six

from swagger_server.models.sr_label_stack import SRLabelStack  # noqa: E501
from swagger_server import util
from voltron.controllers import pathing_controller


def pathing_epe_bandwidth_get(dst_ip, min_bandwidth=None, peer_preference=None):  # noqa: E501
    """Optimize pathing to EPE based on bandwidth.

     # noqa: E501

    :param dst_ip: The destination IP.
    :type dst_ip: str
    :param min_bandwidth: Specification of a minimum allowable bandwidth, otherwise the greatest bandwidth path will be returned.
    :type min_bandwidth: int
    :param peer_preference: Specification of peer preference, either direct or transit
    :type peer_preference: str

    :rtype: SRLabelStack
    """
    return pathing_controller.pathing_epe_bandwidth_get(dst_ip, min_bandwidth, peer_preference)


def pathing_epe_latency_get(src_ip, src_transport_ip, dst_ip, max_latency=None, peer_preference=None):  # noqa: E501
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
    :param peer_preference: Specification of peer preference, either direct or transit
    :type peer_preference: str

    :rtype: SRLabelStack
    """
    return pathing_controller.pathing_epe_latency_get(src_ip, src_transport_ip, dst_ip, max_latency, peer_preference)


def pathing_epe_lossless_get(dst_ip, max_loss=None, peer_preference=None):  # noqa: E501
    """Optimize pathing to EPE based on loss-related statistics.

     # noqa: E501

    :param dst_ip: The destination IP.
    :type dst_ip: str
    :param max_loss: Specification of the maximum allowable loss, otherwise the lowest loss path will be returned.
    :type max_loss: int
    :param peer_preference: Specification of peer preference, either direct or transit
    :type peer_preference: str

    :rtype: SRLabelStack
    """
    return pathing_controller.pathing_epe_lossless_get(dst_ip, max_loss, peer_preference)


def pathing_epe_utilization_get(dst_ip, max_utilization=None, peer_preference=None):  # noqa: E501
    """Optimize pathing to EPE based on utilization percentages.

     # noqa: E501

    :param dst_ip: The destination IP.
    :type dst_ip: str
    :param max_utilization: Specification of the maximum allowable utilization percentage, otherwise the least utilized path will be returned.
    :type max_utilization: int
    :param peer_preference: Specification of peer preference, either direct or transit
    :type peer_preference: str

    :rtype: SRLabelStack
    """
    return pathing_controller.pathing_epe_utilization_get(dst_ip, max_utilization, peer_preference)
