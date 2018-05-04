FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/latency-collector
WORKDIR /services/latency-collector
RUN pip install -r /services/latency-collector/requirements.txt
ENTRYPOINT ["python","network_latency.py"]
