#!/usr/bin/python3.6
"""Deploy Telemetry streaming on host network devices.
Configures devices for telemetry streaming.
"""
import configure_telemetry, enable_guestshell, manage_pipeline


def main():
    print("Configuring telemetry on devices")
    configure_telemetry.main()


if __name__ == '__main__':
    main()
