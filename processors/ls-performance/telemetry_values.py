"""This maps telemetry values in Influx with their logical represenations"""
telemetry_value_mapper = {}

telemetry_value_mapper['state/counters/in_unicast_pkts'] = 'in-unicast-pkts'
telemetry_value_mapper['state/counters/out_unicast_pkts'] = 'out-unicast-pkts'

telemetry_value_mapper['state/counters/in_multicast_pkts'] = 'in-multicast-pkts'
telemetry_value_mapper['state/counters/out_multicast_pkts'] = 'out-multicast-pkts'

telemetry_value_mapper['state/counters/in_broadcast_pkts'] = 'in-broadcast-pkts'
telemetry_value_mapper['state/counters/out_broadcast_pkts'] = 'out-broadcast-pkts'

telemetry_value_mapper['state/counters/in_discards'] = 'in-discards'
telemetry_value_mapper['state/counters/out_discards'] = 'out-discards'

telemetry_value_mapper['state/counters/in_errors'] = 'in-errors'
telemetry_value_mapper['state/counters/out_errors'] = 'out-errors'

telemetry_value_mapper['state/counters/in_octets'] = 'in-octets'
telemetry_value_mapper['state/counters/out_octets'] = 'out-octets'
