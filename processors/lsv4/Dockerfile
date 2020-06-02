FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /processors/lsv4
COPY . /processors/lsv4
RUN pip3 install -r /processors/lsv4/requirements.txt
CMD ["python3.6","create_lsv4_topology.py"]
