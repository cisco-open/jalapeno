# Installing OpenBMPd
You could install from source... but to avoid that, I included a binary that should work on paul's vlab setup. Install the openbmpd binary to /usr/bin and install libdl.so and libdl.so.2 into /usr/local/lib on v0-vm0.

# Router stuff
Make sure the router's OpenBMP is set to send to v0-vm0's address at port 5000. I do not know how that is done. Paul/Bruce/Jeff should fill in this section.

# Starting OpenBMPd
To start openbmpd collector, use the following command (replacing `osc01.rio.wwva.ciscolabs.com:30902` with the openshift kafka broker if different.)
`sudo openbmpd -a v0-vm0 -p 5000 -l /var/log/openbmpd.log -pid /var/run/openbmpd.pid -k osc01.rio.wwva.ciscolabs.com:30909  -debugDebug -dbgp -dbmp -dmsgbus`

# ???

# Profit
Verify it works by checking that openbmpd is running and/or checking the contents of /var/log/openbmpd.log
