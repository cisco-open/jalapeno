#!/usr/bin/python3.6
"""Deploy Telemetry streaming on host network devices.
Configures devices for telemetry streaming, enables guest
shell, and then provisions and starts Pipeline on devices.
"""
import configure_telemetry, enable_guestshell, manage_pipeline

def main():
    print("Stopping Pipeline on devices")
    manage_pipeline.main("stop")

    print("Removing Pipeline on devices")
    manage_pipeline.main("remove")

if __name__ == '__main__':
    main()
