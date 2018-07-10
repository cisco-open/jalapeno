class PlatformNotSupported(Exception):
    """Exception raised when the Voltron client is run on an unsupported platform.
    """
    def __init__(self, expression, message):
        self.expression = expression
        self.message = message

class MPLSNotEnabled(Exception):
    """Exception raised when MPLS is not enabled on the host.
    """
    def __init__(self, expression, message):
        self.expression = expression
        self.message = message

class MPLSNotConfigured(Exception):
    """Exception raised when MPLS is not configured correctly.
    """
    def __init__(self, expression, message):
        self.expression = expression
        self.message = message

class FlowRouteException(Exception):
    """Exception raised when there is an issue with a flow route.
    """
    def __init__(self, expression, message):
        self.expression = expression
        self.message = message
