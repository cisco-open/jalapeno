FROM alpine:3.4
MAINTAINER Zia Syed <ziausyed@cisco.com>

# create working directory
RUN mkdir -p /jalapeno

# set the working directory
WORKDIR /jalapeno

# add binary
COPY bin/jalapeno /bin

# set the entrypoint
ENTRYPOINT ["/bin/jalapeno"]
