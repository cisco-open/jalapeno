FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /processors/lsv6
COPY . /processors/lsv6
RUN pip3 install -r /processors/lsv6/requirements.txt
CMD ["python3.6","create_lsv6_topology.py"]
