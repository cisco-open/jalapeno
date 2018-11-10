FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/egress-links-performance-collector
WORKDIR /services/egress-links-performance-collector
RUN pip install -r /services/egress-links-performance-collector/requirements.txt
ENTRYPOINT ["python","calculate_egress_links_performance.py"]
