NSE Client
1. NSE Connect to GW
2. Client sends L3vpn_Join Request for VRF route(s) to gateway
    1. NSE (or something) identifies its VPP IP info (VPP loopback IP and VPP physical int IP which acts as BGP-LU nexthop)
3. GW queries database for VRF route(s) corresponding to RD, RT
4. DB Query returns route(s) with corresponding nexthop(s) and transport+vpn label(s) - this is FIB
5. GW passes FIB to NSE
    1. NSE programs VPP route+label entries for outbound traffic (remote prefixes, nexthops, transport labels, vpn labels)
    2. NSE programs VPP MPLS pop and forward entry for inbound traffic (vpn label)
```    
mpls table add 0
set interface mpls GigabitEthernet0/7/0 enable
mpls local-label 2000 eos via ip4-lookup-in-table 0
```
6. GW programs GoBGP to advertise BGP-LU entry to TOR (TOR will propagate it) for return traffic, VPP nexthop
```
sudo gobgp global rib add -a ipv4-mpls 10.2.2.1/32 3 nexthop 10.0.130.5

# VPP loopback is 10.2.2.1/32
# VPP nexthop is gig0/7/0 10.0.130.5
# label 3 instructs upstream TOR to PHP
# Todo - gobgp policy to limit advertisement to TOR only
```

7. GW programs GoBGP to advertise vpnv4 entry for NSC prefix, nexthop, vpn_label, rd:rt
```
sudo gobgp global rib add -a vpnv4 10.10.1.0/24 label 2000 rd 100:100 nexthop 10.2.2.1 rt 100:100

# Advertise globally, like Inter-AS Option C
```

