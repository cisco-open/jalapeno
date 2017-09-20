FROM alpine:3.6
MAINTAINER Steve Louie <stelouie@cisco.com>

# create working directory
RUN mkdir -p /vojltorb

# set the working directory
WORKDIR /vojltorb

# add binary
COPY bin/arangodb /bin

# set the command
ENTRYPOINT ["/bin/arangodb"]
