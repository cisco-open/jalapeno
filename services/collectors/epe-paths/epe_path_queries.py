#! /usr/bin/env python
"""AQL Queries executed by the EPE-Path Collector.
"""
    
"""Collect external-prefixes from the existing ExternalPrefix collection.
The list of all external-prefixes is returned."""
def get_external_prefixes_query(db):
    aql = """ FOR i in ExternalPrefixes 
              RETURN i.Prefix """
    bindVars = {}
    external_prefixes = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return external_prefixes

"""Collect EPEEdge documents that connect to a specific destination.
A collection of paths is returned."""
def generate_paths_query(db, destination):
    """AQL Query to generate paths from Arango data."""
    aql = """FOR e in EPEEdges
        FILTER e.Destination == @destination
        RETURN {Source: e.Source, Destination: e.Destination, SRNodeSID: e.SourceSRNodeSID, Interface: e.SourceInterfaceIP, EPELabel: e.EPELabel} """
    bindVars = {'destination': destination}
    paths = db.AQLQuery(aql, rawResults=True, bindVars=bindVars)
    return paths

