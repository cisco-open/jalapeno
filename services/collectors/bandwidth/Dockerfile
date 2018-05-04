FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/bandwidth-collector
WORKDIR /services/bandwidth-collector
RUN pip install -r /services/bandwidth-collector/requirements.txt
ENTRYPOINT ["python","network-utilization.py"]
