# Client Agent
Only supports Linux-based host at the moment, not router.  
Only Python is currently supported as a language library.

## Dependencies
Mostly derived from Sam Russel's [MPLS testbed on Ubuntu Linux with kernel 4.3](http://www.samrussell.nz/2015/12/mpls-testbed-on-ubuntu-linux-with.html).

* MPLS support in Linux kernel.  
Recommend Linux kernel 4.11+. Earlier kernel versions maxed at 2 MPLS labels, 4.11 bumped this number to 30. 
```bash
# Check kernel version
uname -r
# Upgrade if desired, out of doc scope
# Enable MPLS
sudo modprobe mpls_router
sudo modprobe mpls_gso
sudo modprobe mpls_iptunnel
```
* MPLS enabled on an interface.
```bash
# https://www.kernel.org/doc/Documentation/networking/mpls-sysctl.txt
sudo sysctl -w net.mpls.platform_labels=1048575
# Example, replace interfaces with actual
sudo sysctl -w net.mpls.conf.enp0s9.input=1
sudo sysctl -w net.mpls.conf.lo.input=1
```
* Latest ```ip``` util installed.  
```bash
git clone git://git.kernel.org/pub/scm/linux/kernel/git/shemminger/iproute2.git
cd iproute2
./configure
make
sudo make install
```
* User privileges to modify networking state (root, for instance).

To verify correctness, run `python compatibility.py`.

## Installation
```bash
# If pipenv is not installed...
pip install pipenv
# Otherwise...
pipenv --three install
```

## Running Demo
The demo will initialize "optimization" for a flow via Voltron, and ping for 30 seconds while transparently "optimizing" the flow in the background. The flow label stack is reassessed every 3 seconds.
```bash
# Use sudo if you do not otherwise have privilege to change network routes.
sudo pipenv shell
python demo_agent.py --help
# Example, v0-vm0 -> v0-c9
python demo_agent.py 10.1.2.1 2.2.2.1 10.1.1.0 10.11.0.1 127.0.0.1:50000 latency
```
