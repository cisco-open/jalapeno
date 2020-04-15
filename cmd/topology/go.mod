module topology

go 1.14

replace (
	github.com/sbezverk/gobmp/pkg/base => github.com/sbezverk/gobmp/pkg/base v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/bgp => github.com/sbezverk/gobmp/pkg/bgp v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/bgpls => github.com/sbezverk/gobmp/pkg/bgpls v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/bmp => github.com/sbezverk/gobmp/pkg/bmp v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/evpn => github.com/sbezverk/gobmp/pkg/evpn v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/l3vpn => github.com/sbezverk/gobmp/pkg/l3vpn v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/ls => github.com/sbezverk/gobmp/pkg/ls v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/message => github.com/sbezverk/gobmp/pkg/message v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/pub => github.com/sbezverk/gobmp/pkg/pub v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/sr => github.com/sbezverk/gobmp/pkg/sr v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/srv6 => github.com/sbezverk/gobmp/pkg/srv6 v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/tools => github.com/sbezverk/gobmp/pkg/tools v0.0.0-20200415125537-4a28afd8bc12
	github.com/sbezverk/gobmp/pkg/topology/arangodb => ../../pkg/topology/arangodb
	github.com/sbezverk/gobmp/pkg/topology/database => ../../pkg/topology/database
	github.com/sbezverk/gobmp/pkg/topology/dbclient => ../../pkg/topology/dbclient
	github.com/sbezverk/gobmp/pkg/topology/kafkamessenger => ../../pkg/topology/kafkamessenger
	github.com/sbezverk/gobmp/pkg/topology/messenger => ../../pkg/topology/messenger
	github.com/sbezverk/gobmp/pkg/topology/mockdb => ../../pkg/topology/mockdb
	github.com/sbezverk/gobmp/pkg/topology/mockmessenger => ../../pkg/topology/mockmessenger
	github.com/sbezverk/gobmp/pkg/topology/processor => ../../pkg/topology/processor
)

require (
	github.com/arangodb/go-driver v0.0.0-20200403100147-ca5dd87ffe93 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/sbezverk/gobmp/pkg/bgp v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/bmp v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/message v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/pub v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/topology/arangodb v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/topology/database v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/topology/dbclient v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/topology/kafkamessenger v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/topology/messenger v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/topology/mockdb v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/topology/mockmessenger v0.0.0-00010101000000-000000000000 // indirect
	github.com/sbezverk/gobmp/pkg/topology/processor v0.0.0-00010101000000-000000000000 // indirect
	github.com/segmentio/kafka-go v0.3.5 // indirect
)
