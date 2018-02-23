# coding: utf-8

"""
    Voltron Framework

    This outlines version 1 of the voltron framework API. 

    OpenAPI spec version: 1.0
    
    Generated by: https://github.com/swagger-api/swagger-codegen.git
"""


from __future__ import absolute_import

# import models into sdk package
from .models.collector import Collector
from .models.edge_score import EdgeScore

# import apis into sdk package
from .apis.default_api import DefaultApi

# import ApiClient
from .api_client import ApiClient

from .configuration import Configuration

configuration = Configuration()