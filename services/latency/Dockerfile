FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/latency
WORKDIR /services/latency
RUN pip install -r /services/latency/requirements.txt
ENTRYPOINT ["python","network_latency.py"]
