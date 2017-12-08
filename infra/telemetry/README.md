# Telemetry Infrastructure

Contains the telemetry infrastructure related files and configurations to be applied to devices.

## Support
* IOS-XR

## Configuration

* [config_xr](telemetry/config_xr) contains IOS-XR gRPC and MDT configuration.
  * This includes what sensor paths will be pulled.
  * Notably, this does not utilize TLS in gRPC authentication for ease of use in development.

## [Pipeline](telemetry/pipeline/)

This folder contains the necessary files to run Pipeline with the specified configurations. This currently assumes that Pipeline will be located on the host router, authenticating on the host device as ```cisco/cisco```, and exporting to a Kafka bus on ```voltron.telemetry``` topic.

**```pipeline_rsa``` should not be used in production!**
