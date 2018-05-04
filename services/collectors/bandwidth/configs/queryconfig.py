#!/usr/bin/env python
"""Configuration file for path generation query.
Specify the collection to create or join, and the source of the paths.
Destinations should be listed line by line in prefixes.txt.
"""
collection = 'Paths'
upstream_source = '10.1.1.0'  # this is the upstream source (similar to a ToR)
vmsource = '10.1.2.1'  # this is the client source

