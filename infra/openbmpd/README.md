## Installing OpenBMPd
On v0-vm0, install [bin/openbmpd](bin/openbmpd) to `/usr/bin` and install [bin/libdl.so](bin/libdl.so) and [bin/libdl.so.2](bin/libdl.so.2) to `/usr/local/lib`.

## Router Configuration
Each router should send OpenBMP data to v0-vm0's address at port 5000. Configure each device accordingly (see [openbmpd_config](openbmpd_config)). 

## Starting OpenBMPd
To start openbmpd collector, set up the openbmpd systemctl service. Place [openbmpd.service](openbmpd.service) in `/etc/systemd/system/` on v0-vm0. Then enable and start the service:

```
sudo systemctl enable openbmpd
sudo systemctl start openbmpd
```

## Validation
Verify openbmpd is running using `sudo systemctl status openbmpd` and check the log at `/var/log/openbmpd.log`

## OpenBMP data should now be flowing freely from routers to v0-vm0 to Kafka. 
