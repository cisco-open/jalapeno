#!/usr/bin/python
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

    print("Starting Pipeline on devices")
    manage_pipeline.main("restart")

    #print("Stopping any pre-existing Pipeline instances on devices")
    #manage_pipeline.main("stop")

    #print("Removing any pre-existing Pipeline files from devices")
    #manage_pipeline.main("remove")

    #print("Provisioning Pipeline on devices")
    #manage_pipeline.main("provision")

    #print("Starting Pipeline on devices")
    #manage_pipeline.main("start")


if __name__ == '__main__':
    main()
