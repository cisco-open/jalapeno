module github.com/iejalapeno/lsv4_edge

go 1.15

require (
	github.com/Shopify/sarama v1.27.0
	github.com/arangodb/go-driver v0.0.0-20200811070917-cc2b983cf602
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/sbezverk/gobmp v0.0.0-20200902195225-3800ef40ee66
	github.com/jalapeno-sdn/topology v0.0.0 
	github.com/segmentio/kafka-go v0.4.2
	go.uber.org/atomic v1.7.0
)

replace (
//	github.com/sbezverk/gobmp => ../gobmp
//	github.com/jalapeno-sdn/topology => ../topology
)
