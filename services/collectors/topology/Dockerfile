FROM alpine:3.4
MAINTAINER Zia Syed <ziausyed@cisco.com>

# create working directory
RUN mkdir -p /voltron

# set the working directory
WORKDIR /voltron

# add binary
COPY bin/voltront /bin

# set the entrypoint
ENTRYPOINT ["/bin/voltront"]
