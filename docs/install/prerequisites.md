# Prerequisites

Jalapeno has been primarily developed, tested, and operated on:

- Ubuntu 18.04
- Ubuntu 20.04

With the following Kubernetes environments (Bare-metal, VM, or Cloud):

- [Kubernetes](./kubernetes.md#kubernetes-install)
- [MicroK8s](./kubernetes.md#microk8s-install)

Recommended VM sizing for a test lab:

- 4 vCPU
- 16GB memory
- 50G of disk.

!!! tip
    If deploying in production or a test environment with large table sizes (full Internet table, 250k + internal or vpn prefixes), then we recommend a bare metal K8s cluster with two or more nodes.
