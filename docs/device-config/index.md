# Device Configuration

This section will contain notes and configuration snippets for getting devices to report information to Jalapeno.

Jalapeno's default MDT port is 32400 and the BMP port is 30511.

Generally we would setup MDT on **all** routers and BMP **only** on route reflectors and any routers with external peering sessions.

Currently, we have documentation for the following platforms:

- [Cisco IOS-XR](xr-config.md)
