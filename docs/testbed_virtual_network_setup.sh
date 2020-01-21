#/bin/bash

echo adding ovs bridges
ovs-vsctl add-br vlt_br
ovs-vsctl add-br vlt_mgt_br
ovs-vsctl add-br vlt_outside_br
sleep 1

echo bringing up vlt interfaces
ifconfig vlt_br up
ifconfig vlt_mgt_br up
ifconfig vlt_outside_br up

ip addr add 10.251.251.1/24 dev vlt_mgt_br
ip addr add 10.0.254.1/24 dev vlt_outside_br
sleep 1

echo adding static routes
ip route add 10.0.250.0/23 via 10.0.254.71 dev vlt_outside_br
ip route add 10.0.130.0/24 via 10.0.254.71 dev vlt_outside_br
ip route add 10.0.131.0/24 via 10.0.254.76 dev vlt_outside_br
sleep 1

echo adding iptables NAT entries so jalapeno k8s containers may be reached from the outside world
# openshift UI
iptables -t nat -A PREROUTING -p tcp -m tcp --dport 8443 -j DNAT --to-destination 10.251.251.250:8443

# kafka
iptables -t nat -A PREROUTING -p tcp -m tcp --dport 30300 -j DNAT --to-destination 10.251.251.250:30300

# arangoDB
iptables -t nat -A PREROUTING -p tcp -m tcp --dport 30852 -j DNAT --to-destination 10.251.251.250:30852

iptables -t nat -A PREROUTING -p tcp -m tcp --dport 30902 -j DNAT --to-destination 10.251.251.250:30902
iptables -t nat -A PREROUTING -p tcp -m tcp --dport 30308 -j DNAT --to-destination 10.251.251.250:30308
iptables -t nat -A PREROUTING -p tcp -m tcp --dport 30880 -j DNAT --to-destination 10.251.251.250:30880
iptables -t nat -A PREROUTING -p tcp -m tcp --dport 30881 -j DNAT --to-destination 10.251.251.250:30881

# source NAT for external traffic inbound to VM management interfaces
iptables -t nat -A POSTROUTING -o vlt_mgt_br -j MASQUERADE

# source NAT for external traffic inbound to virtual topology
iptables -t nat -A POSTROUTING -o vlt_outside_br -j MASQUERADE

# optional: NAT all outbound traffic from jalapeno/the server itself.  Uncomment if needed
# iptables -t nat -A POSTROUTING -o <server_outside_nic> -j MASQUERADE

