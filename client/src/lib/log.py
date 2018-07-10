import logging

logger_name = 'voltron'

def setup_logging():
    get_logger().setLevel(logging.DEBUG)
    handler = logging.FileHandler('{}.log'.format(logger_name))
    handler.setLevel(logging.DEBUG)
    formatter = logging.Formatter('%(asctime)s:%(levelname)s:%(name)s:%(message)s')
    handler.setFormatter(formatter)
    get_logger().addHandler(handler)

def get_logger():
    return logging.getLogger(logger_name)

def info(message):
    get_logger().info(message)

def debug(message):
    get_logger().debug(message)

def error(message):
    get_logger().error(message)

setup_logging()