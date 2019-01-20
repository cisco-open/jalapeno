# coding: utf-8

from __future__ import absolute_import

from flask import json
from six import BytesIO

from swagger_server.models.sr_label_stack import SRLabelStack  # noqa: E501
from swagger_server.test import BaseTestCase


class TestPathingController(BaseTestCase):
    """PathingController integration test stubs"""

    def test_pathing_epe_bandwidth_get(self):
        """Test case for pathing_epe_bandwidth_get

        Optimize pathing to EPE based on bandwidth.
        """
        query_string = [('dst_ip', 'dst_ip_example'),
                        ('min_bandwidth', 56),
                        ('peer_preference', 'peer_preference_example')]
        response = self.client.open(
            '/api/v1/pathing/epe/bandwidth',
            method='GET',
            content_type='application/json',
            query_string=query_string)
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))

    def test_pathing_epe_latency_get(self):
        """Test case for pathing_epe_latency_get

        Optimize pathing to EPE based on latency.
        """
        query_string = [('src_ip', 'src_ip_example'),
                        ('src_transport_ip', 'src_transport_ip_example'),
                        ('dst_ip', 'dst_ip_example'),
                        ('max_latency', 56),
                        ('peer_preference', 'peer_preference_example')]
        response = self.client.open(
            '/api/v1/pathing/epe/latency',
            method='GET',
            content_type='application/json',
            query_string=query_string)
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))

    def test_pathing_epe_lossless_get(self):
        """Test case for pathing_epe_lossless_get

        Optimize pathing to EPE based on loss-related statistics.
        """
        query_string = [('dst_ip', 'dst_ip_example'),
                        ('max_loss', 56),
                        ('peer_preference', 'peer_preference_example')]
        response = self.client.open(
            '/api/v1/pathing/epe/lossless',
            method='GET',
            content_type='application/json',
            query_string=query_string)
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))

    def test_pathing_epe_utilization_get(self):
        """Test case for pathing_epe_utilization_get

        Optimize pathing to EPE based on utilization percentages.
        """
        query_string = [('dst_ip', 'dst_ip_example'),
                        ('max_utilization', 56),
                        ('peer_preference', 'peer_preference_example')]
        response = self.client.open(
            '/api/v1/pathing/epe/utilization',
            method='GET',
            content_type='application/json',
            query_string=query_string)
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    import unittest
    unittest.main()
