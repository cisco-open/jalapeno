# IOS-XR

This section contains notes on IOS-XR router configs and network design.

## Segment Routing

1. SRGB: we use a custom label block in our lab.  The default is 16000 - 23999.

    ```text
    segment-routing
     global-block 100000 163999   
    ```

2. Enable SR in ISIS

    ```text
    router isis 100
     is-type level-2-only
     net 49.0901.0000.0000.0006.00
     address-family ipv4 unicast
      metric-style wide
      advertise link attributes
      mpls traffic-eng level-2-only
      mpls traffic-eng router-id Loopback0
      maximum-paths 32
      segment-routing mpls
     !
     interface Loopback0
      passive
      address-family ipv4 unicast
       prefix-sid absolute 100006    // this can also be an SRGB index value
      !
    ```

3. Enable distribution of SR Prefix-SID information in BGP-LU (Optional, but very valuable for multi-domain networks)

    ```text
    route-policy SID($SID)
      set label-index $SID
    end-policy
    !
    router bgp 100000
    !
     address-family ipv4 unicast
      network 10.0.0.1/32 route-policy SID(1)    // BGP Prefix-SID index 1, results in SR label value 100001 based on our SRGB
      allocate-label all
     !
     neighbor 10.1.1.0   // ASBR
      remote-as 65000
      !       
      address-family ipv4 labeled-unicast
       route-policy pass in
       route-policy pass out
    !
    ```

### Setting up BMP and BGP-LS

#### Diagram

![Getting BMP data to Jalapeno](../img/BGP-LS-and-BMP.png)

### BGP-LS

1. Setup route-reflectors to receive LS messages from clients, but to not pass LS messages back out:

    ```text
    BGP-LS specific information shown:

    router bgp 100000
     !
     address-family link-state link-state
     !
     neighbor <neighbor IP>
      address-family link-state link-state
       route-policy pass in
       route-policy drop out
    ```

!!! note
    The RR's technically only need a copy of BGP-LS/LSDB from one router.

2. Configure one or more RR clients to pass LS messages to the RR's.  

    BGP-LS specific information shown:

    ```text
    
    router isis 100
     distribute link-state level 2  // Distribute my LSDB into local BGP-LS
     !
     address-family ipv4 unicast
      advertise link attributes   // all ISIS nodes should have this line as it adds their SR TLVs into the domain's LSDB
     !
    router bgp 100000
     !
     address-family link-state link-state
     !
     neighbor 192.0.2.50   // Route Reflector
      !
      address-family link-state link-state
       route-policy drop in
       route-policy pass out
    ```

3. Add Egress Peer Engineering data to BGP-LS feed (Optional)

    On an ASBR node, or Internet peering node add the following (it is assumed v4/v6 AFIs are already enabled):

    ```text
    router bgp 100000
     !
     address-family link-state link-state
     !
     neighbor 192.0.2.50   // Route Reflector
      !       
      address-family link-state link-state
       route-policy pass out
       !
      !
     neighbor 203.0.113.50    // External Peer
      remote-as 64496 
      egress-engineering
     !
    ```

### BMP (BGP Monitoring Protocol)

We collect topology data with BMP as it provides data from multiple AFI/SAFI combinations (not just BGP-LS)
While we anticipate most operators to run a BGP-free core, we'll generally want to collect BMP messages from route-reflectors and all BGP speakers with external facing peering sessions (ASBRs, Peering, etc.):

1. Configure BMP Server on RR's and on ASBRs and peering routers:

    ```text
    bmp server 1
     host 10.0.250.2 port 30511
     description Jalapeno GoBMP 
     update-source Loopback0  // or MgmtEth0/RP0/CPU0/0
     flapping-delay 60
     initial-delay 5
     stats-reporting-period 60
     initial-refresh delay 30 spread 2
    ```

2. Configure export of BMP data for BGP messages/advertisements from specific peers:

    ```text
     // Route reflector:
     
     router bgp 100000
     neighbor <neighbor IP>    // Assuming clients are vpnv4/v6 PE's this also captures vpnv4/6 messages
      bmp-activate server 1
      
    // ASBR/Peering
    
     router bgp 100000
     !
     neighbor 192.0.2.100
      remote-as 64500
      description External Peer 72
      egress-engineering
      bmp-activate server 1
    ```

### Streaming Telemetry (Model Driven Telemetry)

The following configuration snippet provides a reference for sending streaming telemetry to Jalapeno.

    ```
    telemetry model-driven
    destination-group jalapeno
     vrf <name> // optional 
     address-family ipv4 192.0.2.10 port 32400
      encoding self-describing-gpb
      protocol grpc no-tls
     !
    !
    sensor-group cisco_models 
     sensor-path Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface  // Interface statistics
     sensor-path Cisco-IOS-XR-infra-tc-oper:traffic-collector/afs/af/counters/prefixes/prefix // SR traffic collector statistics
     sensor-path Cisco-IOS-XR-fib-common-oper:mpls-forwarding/nodes/node/label-fib/forwarding-details/forwarding-detail // Per-MPLS label forwarding statistics
    !
    sensor-group openconfig_interfaces
     sensor-path openconfig-interfaces:interfaces/interface  // Openconfig interface statistics
    !
    subscription base_metrics
     sensor-group-id cisco_models sample-interval 10000
     sensor-group-id openconfig_interfaces sample-interval 10000
     destination-id jalapeno
     source-interface MgmtEth0/RP0/CPU0/0
    !
    ```
