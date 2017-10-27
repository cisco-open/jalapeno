FROM alpine:3.4
MAINTAINER Steve Louie <stelouie@cisco.com>

# create working directory
RUN mkdir -p /voltron

# set the working directory
WORKDIR /voltron

# add binary
COPY bin/framework /bin

# set the entrypoint
ENTRYPOINT ["/bin/framework"]
