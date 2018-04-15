# kafka-http-proxy
To avoid having to import kafka-libs into the router code, we opted to use a kafka-http-proxy which uses REST and forwards to kafka. I recommend rewriting the collector to use an actual kafka library to avoid this, but here are the instructions to set it up regardless.

#Install
Somehow magically install it via https://docs.confluent.io/current/installation/installing_cp.html#installation

Really just rewrite the stupid ass collector.

# Configure
set the following in kafka-rest.properties:
```
zookeeper.connect=osc01.rio.wwva.ciscolabs.com:30218
bootstrap.servers=PLAINTEXT://osc01.rio.wwva.ciscolabs.com:30902
```
Using the public addresses of zookeeper and kafka respectively.


#Run
`kafka-rest-start kafka-rest.properties`
