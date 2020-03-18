FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /processors/l3vpn
COPY . /processors/l3vpn
RUN pip3 install -r /processors/l3vpn/requirements.txt
CMD ["python3.6","create_l3vpn_topology.py"]
