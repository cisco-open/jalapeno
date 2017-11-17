# Client Agent
Only supports Linux-based host at the moment, not router.  
Only Python is currently supported as a language library.

## Dependencies
```python compatibility.py```
* MPLS support in Linux kernel.
* MPLS enabled on an interface.
* Latest ```ip``` util installed.  
```bash
git clone git://git.kernel.org/pub/scm/linux/kernel/git/shemminger/iproute2.git
cd iproute2
./configure
make
sudo make install
```
* User privileges to modify networking state (root, for instance).

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
