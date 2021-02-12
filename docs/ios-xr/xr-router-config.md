### Notes on IOS-XR router configs and network design

### Segment Routing

1. SRGB - we use a custom label block in our lab.  The default is 16000 - 23999
```
segment-routing
 global-block 100000 163999
```

2. Enable SR in ISIS
```
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
```
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
### Setting up BMP and BGP-LS:

![Getting BMP data to Jalapeno](https://github.com/jalapeno/jalapeno/blob/master/docs/diagrams/BGP-LS-and-BMP.png)

### BGP-LS

1. setup route-reflectors to receive LS messages from clients, but to not pass LS messages back out:
```
BGP-LS specific information shown:

router bgp 100000
 !
 address-family link-state link-state
 !
 neighbor-group ASN100000_clients
  !       
  address-family link-state link-state
   route-policy pass in
   route-policy drop out
```
The RR's technically only need a copy of BGP-LS/LSDB from one router.

2. Configure one or more RR clients to pass LS messages to the RR's.  
```
BGP-LS specific information shown:

router isis 100
 distribute link-state level 2  // distribute my LSDB into local BGP-LS
 !
 address-family ipv4 unicast
  advertise link attributes   // all ISIS nodes should have this line as it adds their SR TLVs into the domain's LSDB
 !
router bgp 100000
 !
 address-family link-state link-state
 !
 neighbor 10.0.0.10   // route-reflector
  !
  address-family link-state link-state
   route-policy drop in
   route-policy pass out
```

3.  Add Egress Peer Engineering data to BGP-LS feed (optional):
```
On an ASBR node, or Internet peering node add the following (it is assumed v4/v6 AFIs are already enabled):

router bgp 100000
 !
 address-family link-state link-state
 !
 neighbor 10.0.0.10   // Route Reflector
  !       
  address-family link-state link-state
   route-policy pass out
   !
  !
 neighbor 10.71.0.1    // external peer
  remote-as 7100
  egress-engineering
 !
```
### BMP (BGP Monitoring Protocol)
We collect topology data with BMP as it provides data from multiple AFI/SAFI combinations (not just BGP-LS)
While we anticipate most operators to run a BGP-free core, we'll generally want to collect BMP messages from route-reflectors and all BGP speakers with external facing peering sessions (ASBRs, Peering, etc.):

1.  Configure BMP Server on RR's and on ASBRs and peering routers:
```
bmp server 1
 host 10.0.250.2 port 5000
 description Jalapeno GoBMP 
 update-source Loopback0  // or MgmtEth0/RP0/CPU0/0
 flapping-delay 60
 initial-delay 5
 stats-reporting-period 60
 initial-refresh delay 30 spread 2
```

2. Configure export of BMP data for BGP messages/advertisements from specific peers:
```
 // Route reflector:
 
 router bgp 100000
 neighbor-group ASN100000_clients    // Assuming clients are vpnv4/v6 PE's this also captures vpnv4/6 messages
  remote-as 100000
  bmp-activate server 1
  
// ASBR/Peering

 router bgp 100000
 !
 neighbor 10.72.1.1
  remote-as 7200
  description External Peer 72
  egress-engineering
  bmp-activate server 1
 ```
 
