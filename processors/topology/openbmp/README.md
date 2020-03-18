BGP Monitoring Protocol (BMP) is defined in RFC 7854:

https://tools.ietf.org/html/rfc7854

Per the RFC:

   BMP provides access to the Adj-RIB-In of a peer on an ongoing basis
   and a periodic dump of certain statistics the monitoring station can
   use for further analysis.  From a high level, BMP can be thought of
   as the result of multiplexing together the messages received on the
   various monitored BGP sessions.
   
 From the perspective of the Jalapeno project BMP represents a programmatic means of collecting all BGP topology data:
 
 * BGP-LS for internal topology
 * eBGP and iBGP IPv4, IPv6, and labeled-unicast for external/Internet or Inter-AS topology
 * MP-BGP for VPNv4, VPNv6, EVPN, etc. for VPN overlay topology

Jalapeno leverages the open source OpenBMP (snas.io) collector to capture BMP data from the network.

https://www.snas.io/

The following configuration is used to establish a BMP session from XR router to OpenBMP:
```
 bmp server 1
 host 10.0.250.2 port 5000
 description jalapeno OpenBMP  
 update-source Loopback0
 flapping-delay 60
 initial-delay 5
 stats-reporting-period 60
 initial-refresh delay 30 spread 2
``` 
 And enable BMP monitoring for each peer you wish to collect BGP data from:
 
 ```
 router bgp 65000
  neighbor 10.1.1.1
  remote-as 100000
  egress-engineering
  description eBGP to r1  
  bmp-activate server 1
  address-family ipv4 unicast
   route-policy pass in
   route-policy pass out
 ```
 
OpenBMP collector logs may be found/monitored here on the OpenBMP container 
```
tail -f /var/log/openbmpd.log
```

 
