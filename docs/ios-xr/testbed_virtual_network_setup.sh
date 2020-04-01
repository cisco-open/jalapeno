#/bin/bash

echo adding ovs bridges
ovs-vsctl add-br jalapeno_br
ovs-vsctl add-br jalapeno_mgt_br
sleep 1

echo bringing up vlt interfaces
ifconfig jalapeno_br up
ifconfig jalapeno_mgt_br up

ip addr add 10.251.251.1/24 dev jalapeno_mgt_br
sleep 1


