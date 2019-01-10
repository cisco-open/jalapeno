from voltron import queries


def topology_get():  # noqa: E501
    """Retrieve the topology representation in the database in d3-desired form.

     # noqa: E501


    :rtype: D3Topology
    """
    return queries.topology_get()
