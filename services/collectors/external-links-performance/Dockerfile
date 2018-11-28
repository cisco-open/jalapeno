FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/external-links-performance-collector
WORKDIR /services/external-links-performance-collector
RUN pip install -r /services/external-links-performance-collector/requirements.txt
ENTRYPOINT ["python","calculate_external_links_performance.py"]
