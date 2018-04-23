FROM alpine:3.4
MAINTAINER Zia Syed <ziausyed@cisco.com>

# create working directory
RUN mkdir -p /voltron

# set the working directory
WORKDIR /voltron

# add binary
COPY bin/voltron /bin

# set the entrypoint
ENTRYPOINT ["/bin/voltron"]
