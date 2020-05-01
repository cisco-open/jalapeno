module topology

go 1.14

replace (
        github.com/cisco-ie/jalapeno/processors/topology/pkg/arangodb => ../../pkg/arangodb
        github.com/cisco-ie/jalapeno/processors/topology/pkg/database => ../../pkg/database
)

require (
        github.com/arangodb/go-driver v0.0.0-20200403100147-ca5dd87ffe93 // indirect
        github.com/cisco-ie/jalapeno/processors/gobmp-topology/pkg/arangodb v0.0.0-00010101000000-000000000000 // indirect
        github.com/cisco-ie/jalapeno/processors/gobmp-topology/pkg/database v0.0.0-00010101000000-000000000000 // indirect
        github.com/sbezverk/gobmp/pkg/base v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/bgp v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/bgpls v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/bmp v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/evpn v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/l3vpn v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/ls v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/message v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/prefixsid v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/pub v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/sr v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/srv6 v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/tools v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/topology/dbclient v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/topology/kafkamessenger v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/topology/messenger v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/topology/mockdb v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/topology/mockmessenger v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/topology/processor v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/sbezverk/gobmp/pkg/unicast v0.0.0-20200428015650-10acbc383044 // indirect
        github.com/segmentio/kafka-go v0.3.5 // indirect
)

