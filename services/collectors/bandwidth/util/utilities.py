#! /usr/bin/env python
"""This class has utility functions that multiple services may use."""

def uniqify(list, id=None):
    """Remove duplicate values from list"""
    if id is None:
        def id(x): return x
    counts = {}
    newList = []
    for item in list:
        current = id(item)
        if current in counts: continue
        counts[current] = 1
        newList.append(item)
    return newList

