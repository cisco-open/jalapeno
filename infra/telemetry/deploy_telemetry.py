#!/usr/bin/python3.6
"""Deploy Telemetry streaming on host network devices.
Configures devices for telemetry streaming, enables guest
shell, and then provisions and starts Pipeline on devices.
"""
import configure_telemetry, enable_guestshell, manage_pipeline

def main():
    print("Configuring telemetry on devices")
    configure_telemetry.main()

    print("Enabling guest shell on devices")
    enable_guestshell.main()

    print("Provisioning Pipeline on devices")
    manage_pipeline.main("provision")

    print("Starting Pipeline on devices")
    manage_pipeline.main("start")


if __name__ == '__main__':
    main()
