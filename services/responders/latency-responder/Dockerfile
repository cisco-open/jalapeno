FROM alpine:3.4
MAINTAINER Steve Louie <stelouie@cisco.com>

# create working directory
RUN mkdir -p /responder

# set the working directory
WORKDIR /responder

# add binary
COPY bin/responder /bin/responder

# set the entrypoint
ENTRYPOINT ["/bin/responder"]
