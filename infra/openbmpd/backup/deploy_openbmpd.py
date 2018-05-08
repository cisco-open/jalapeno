import shutil
#
# script to deploy openbmp
#

# move bin/openbmpd to usr/bin and install bin/libdl.so and bin/libdl.so.2 to usr/local/lib
#shutil.copy("./bin/openbmpd", "/usr/bin/openbmpd2")
#shutil.copy("./bin/libdl.so", "/usr/local/lib/libdl.so")
#shutil.copy("./bin/libdl.so.2", "/usr/local/lib/libdl.so.2")
# configure each device according to openbmpd_config
# set up the openbmpd systemctl service
# place openbmpd.service in etc/systemd/system
#shutil.copy("./openbmpd.service", "/etc/systemd/system/openbmpd.service")
# start the openbmpd service
# verify openbmpd is running using `sudo systemctl status openbmpd`


# sudo openbmpd -a v0-vm0 -p 5000 -l /var/log/openbmpd.log -pid /var/run/openbmpd.pid -k ie-snas1.cisco.com:30902 -debugDebug -dbgp -dbmp -dmsgbus
