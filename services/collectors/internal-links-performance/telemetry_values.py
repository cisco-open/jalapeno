"""This maps telemetry values in Influx with their logical represenations"""
telemetry_value_mapper = {}
telemetry_value_mapper['state__counters__in-unicast-pkts'] = 'in-unicast-pkts'
telemetry_value_mapper['state__counters__out-unicast-pkts'] = 'out-unicast-pkts'

telemetry_value_mapper['state__counters__in-multicast-pkts'] = 'in-multicast-pkts'
telemetry_value_mapper['state__counters__out-multicast-pkts'] = 'out-multicast-pkts'

telemetry_value_mapper['state__counters__in-broadcast-pkts'] = 'in-broadcast-pkts'
telemetry_value_mapper['state__counters__out-broadcast-pkts'] = 'out-broadcast-pkts'

telemetry_value_mapper['state__counters__in-discards'] = 'in-discards'
telemetry_value_mapper['state__counters__out-discards'] = 'out-discards'

telemetry_value_mapper['state__counters__in-errors'] = 'in-errors'
telemetry_value_mapper['state__counters__out-errors'] = 'out-errors'

telemetry_value_mapper['state__counters__in-octets'] = 'in-octets'
telemetry_value_mapper['state__counters__out-octets'] = 'out-octets'
