# Voltron Clients and Libraries

The thing that is asking Voltron for a path.

## Strategies

1. [Host Agent](#host-agent)
2. [Router Agent](#router-agent)
3. [BPF Program](#bpf-program)
4. [Language Library](#language-library)
5. [Kernel Module](#kernel-module)

We need to perform three fundamental things in order to accomplish Voltron as well as SR label imposition at the host.
1. Provide the application an interface to communicate with Voltron.  
e.g. "My intent is XYZ."
2. Provide the system information to Voltron for decision making.  
e.g. packets sent, retransmits, etc. 
3. Enable SR label imposition either at the host or application level.

The following are provided in order of ease of implementation and likelihood of success.

### Host Agent
This is foundationally important for Voltron. Voltron necessitates label imposition at the host as part of its charter. The host agent requires several components:
1. Kernel support for SR, [OVS](http://openvswitch.org/), or [VPP](https://fd.io/) running on the host.
2. An agent running on the host which has the ability to impose labels via commands to kernel/OVS/VPP.
3. A language library for the application to interface with and declare intent to Voltron services.
    1. Will the application communicate to the host agent what labels to impose? May become complex with containers.
    2. Will Voltron itself communicate to the host agent what labels to impose?  
    This solves the problem of containers needing to communicate to the actual host agent to manipulate the labels, because  from the container it is impossible(?) for it to change its own label stack (except for kernel support) due to needing to issue commands to VPP or OVS which it is segmented from. Or is this an incorrect understanding?

#### Challenges
1. No actual benchmarking in place on performance loss from VPP or OVS.
2. ~2-4% performance loss on hosts using OVS+DPDK.
3. ~? performance loss on hosts using VPP.
4. Potential loss of hardware optimizations for IP/TCP. This looks to be in the 70-80% range of throughput..

### Router Agent
An agent which interfaces with Voltron and uses policies to impose SR on a per flow basis. This is the easiest actual implementation, but is counter to Voltron's foundational charter of host label imposition.

#### Challenges
1. XTC+WAE already does router based SR. Voltron has a unique value proposition though.

### BPF Program
Experimental. Use case is mainly monitoring and tracing.  
Attempt to write a BPF implementation which modifies packets in flight. BPF runs in kernel, is pre-validated to never negatively impact kernel, and is JIT.   
Tools to facilitate: [BCC](https://github.com/iovisor/bcc)  
More information: http://www.brendangregg.com/ebpf.html

#### Challenges
1. BPF is a relatively new skillset.
2. Not validated. Uncertain if actual modification of network data in flight is possible.

### Language Library
We can implement a language library which modifies the behavior of the networking stack to enable communication with Voltron and SR label imposition at the host nearly transparently while affording Voltron APIs.
e.g. Python, Go, etc.

#### Challenges
1. Loss of optimizations.
2. Requirement to use raw sockets and software processing for serialization of packets.
3. Necessitates a userspace TCP/IP implementation?

### Kernel Module
We can implement a Linux kernel module which interfaces with Voltron and controls flows.

#### Challenges
1. Native Linux support for SR is currently [limited](http://www.segment-routing.net/open-software/linux/) in functionality. There is no reference for support on Windows or FreeBSD.
2. Exceedingly complex.
3. Requires flawless execution.
4. Requires knowledge of kernel modules.

