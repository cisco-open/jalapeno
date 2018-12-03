from ipaddress import ip_interface

def ip_network_bandaid(address, mask='24'):
    """Default /24 an IP address to work around not knowing destination prefix.
    """
    if '/' not in address:
        address = '%s/%s' % (address, mask)
    return str(ip_interface(address).network.with_prefixlen)
