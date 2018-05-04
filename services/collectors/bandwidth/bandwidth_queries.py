link_utilization_query = """SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-t0\' AND \"interface-name\" = \'GigabitEthernet0/0/0/1\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-t0\' AND \"interface-name\" = \'GigabitEthernet0/0/0/2\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-l1\' AND \"interface-name\" = \'GigabitEthernet0/0/0/1\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-l1\' AND \"interface-name\" = \'GigabitEthernet0/0/0/2\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-l2\' AND \"interface-name\" = \'GigabitEthernet0/0/0/1\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-l2\' AND \"interface-name\" = \'GigabitEthernet0/0/0/2\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-p3\' AND \"interface-name\" = \'GigabitEthernet0/0/0/2\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-p3\' AND \"interface-name\" = \'GigabitEthernet0/0/0/3\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-p3\' AND \"interface-name\" = \'GigabitEthernet0/0/0/4\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-p4\' AND \"interface-name\" = \'GigabitEthernet0/0/0/2\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-p4\' AND \"interface-name\" = \'GigabitEthernet0/0/0/3\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);

          SELECT moving_average(last(\"bytes-sent\"), 5)
          FROM \"Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters\"
          WHERE (\"Producer\" = \'v0-p4\' AND \"interface-name\" = \'GigabitEthernet0/0/0/4\')
          AND time >= now() - 5m GROUP BY time(200ms), \"Producer\", \"interface-name\" fill(null);"""
