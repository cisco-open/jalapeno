# Installing OpenBMPd

**v0-vm0:**

Install [bin/openbmpd](bin/openbmpd) to `/usr/bin`

Install [bin/libdl.so](bin/libdl.so) and [bin/libdl.so.2](bin/libdl.so.2) to `/usr/local/lib`

# Router Configuration
Each router should send OpenBMP data to v0-vm0's address at port 5000.

Configure each device accordingly (see [openbmpd_config](openbmpd_config)). 

# Starting OpenBMPd
To start openbmpd collector, set up the openbmpd systemctl service.

Place [openbmpd.service](openbmpd.service) in `/etc/systemd/system/` on v0-vm0 and run:
`sudo systemctl enable openbmpd`

Start the service:
`sudo systemctl start openbmpd`

# Validation
Verify openbmpd is running: `sudo systemctl status openbmpd`

Check log: `/var/log/openbmpd.log`

OpenBMP data should now freely from routers to v0-vm0 to Kafka. 
