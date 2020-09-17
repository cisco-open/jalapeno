module github.com/sbezverk/topology

go 1.15

require (
	github.com/arangodb/go-driver v0.0.0-20200403100147-ca5dd87ffe93
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/sbezverk/gobmp v0.0.0-20200902195225-3800ef40ee66
	github.com/segmentio/kafka-go v0.4.2
)

replace github.com/sbezverk/gobmp => ../gobmp
