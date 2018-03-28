# Installing OpenBMPd
Install OpenBMPd from source, or install [bin/openbmpd](bin/openbmpd) to `/usr/bin` and install [bin/libdl.so](bin/libdl.so) and [bin/libdl.so.2](bin/libdl.so.2) into `/usr/local/lib` on v0-vm0.

# Router Configuration
Each router should send OpenBMP data to v0-vm0's address at port 5000. Configure each device accordingly (see [openbmpd/openbmpd_config](../openbmpd_config)). 

# Starting OpenBMPd
To start openbmpd collector, set up the openbmpd systemctl service.
Place [openbmpd/openbmpd.service](../openbmpd.service) in `/etc/systemd/system/` on v0-vm0 and run:
`sudo systemctl enable openbmpd`

Start the service:
`sudo systemctl start openbmpd`

# Validation
Verify openbmpd is running: `sudo systemctl status openbmpd`
Check log: `/var/log/openbmpd.log`

OpenBMP data should now freely from routers to v0-vm0 to Kafka. 
