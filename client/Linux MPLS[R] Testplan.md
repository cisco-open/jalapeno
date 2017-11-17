# Linux MPLS[R] Testplan
This document serves as a guideline for validating SR at the host by using MPLS at the host and SR at the router.

## Foreword
There is very little data about the current state of SR at the host. There are kernel implementations, but MPLS is already supported in the kernel and thus if we can just use the MPLS support for basic LSR, and advanced SR functionality does not matter, we can utilize OOB behaviors making our lives easier.

## Scenarios

### Container <-> Container
http://www.samrussell.nz/2015/12/mpls-testbed-on-ubuntu-linux-with.html  
Configure 2 containers to speak directly to each other over MPLS interfaces.

#### Results
N/A

### Container <-> XRv <-> Container
Configure 2 containers to utilize MPLS, and an XRv to use SR, and attempt to pass traffic.

#### Results
N/A

### Machine <-> Machine
Configure 2 physical devices to utilize MPLS, and an XR device to use SR, and attempt to pass traffic.

#### Results
N/A
