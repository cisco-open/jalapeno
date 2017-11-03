# coding: utf-8

"""
    Voltron Framework

    This outlines version 1 of the voltron framework API. 

    OpenAPI spec version: 1.0
    
    Generated by: https://github.com/swagger-api/swagger-codegen.git
"""


from pprint import pformat
from six import iteritems
import re


class EdgeScore(object):
    """
    NOTE: This class is auto generated by the swagger code generator program.
    Do not edit the class manually.
    """
    def __init__(self, key=None, value=None):
        """
        EdgeScore - a model defined in Swagger

        :param dict swaggerTypes: The key is attribute name
                                  and the value is attribute type.
        :param dict attributeMap: The key is attribute name
                                  and the value is json key in definition.
        """
        self.swagger_types = {
            'key': 'str',
            'value': 'float'
        }

        self.attribute_map = {
            'key': 'key',
            'value': 'value'
        }

        self._key = key
        self._value = value

    @property
    def key(self):
        """
        Gets the key of this EdgeScore.

        :return: The key of this EdgeScore.
        :rtype: str
        """
        return self._key

    @key.setter
    def key(self, key):
        """
        Sets the key of this EdgeScore.

        :param key: The key of this EdgeScore.
        :type: str
        """

        self._key = key

    @property
    def value(self):
        """
        Gets the value of this EdgeScore.

        :return: The value of this EdgeScore.
        :rtype: float
        """
        return self._value

    @value.setter
    def value(self, value):
        """
        Sets the value of this EdgeScore.

        :param value: The value of this EdgeScore.
        :type: float
        """

        self._value = value

    def to_dict(self):
        """
        Returns the model properties as a dict
        """
        result = {}

        for attr, _ in iteritems(self.swagger_types):
            value = getattr(self, attr)
            if isinstance(value, list):
                result[attr] = list(map(
                    lambda x: x.to_dict() if hasattr(x, "to_dict") else x,
                    value
                ))
            elif hasattr(value, "to_dict"):
                result[attr] = value.to_dict()
            elif isinstance(value, dict):
                result[attr] = dict(map(
                    lambda item: (item[0], item[1].to_dict())
                    if hasattr(item[1], "to_dict") else item,
                    value.items()
                ))
            else:
                result[attr] = value

        return result

    def to_str(self):
        """
        Returns the string representation of the model
        """
        return pformat(self.to_dict())

    def __repr__(self):
        """
        For `print` and `pprint`
        """
        return self.to_str()

    def __eq__(self, other):
        """
        Returns true if both objects are equal
        """
        if not isinstance(other, EdgeScore):
            return False

        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """
        Returns true if both objects are not equal
        """
        return not self == other
