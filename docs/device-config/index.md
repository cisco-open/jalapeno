# Device Configuration

This section will contain notes and configuration snippets for getting devices to report information to Jalapeno.

## General Notes

Jalapeno uses the following default ports:

- TCP/32400 for gRPC-based MDT
- TCP/30511 for BMP

Generally it is recommended to setup MDT on **all** routers, but only enable BMP **only** on route reflectors & any routers with external peering sessions.

## Device Specific Config

Currently, we have documentation for the following platforms:

<div class="grid cards" markdown>

- :simple-cisco: **[Cisco IOS-XR](xr-config.md)**

</div>

Missing a platform? Feel free to [submit an issue](https://github.com/cisco-open/jalapeno/issues) to request it!

<br/>
