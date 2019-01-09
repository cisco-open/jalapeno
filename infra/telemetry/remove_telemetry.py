#!/usr/bin/python3.6
"""Remove Telemetry streaming on host network devices.
Unconfigures devices' currently telemetry configurations.
"""
import configure_telemetry

def main():
    print("Unconfiguring telemetry on devices")
    configure_telemetry.remove_telemetry_config()

if __name__ == '__main__':
    main()
