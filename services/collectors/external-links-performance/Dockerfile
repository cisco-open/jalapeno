FROM python:3.6-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/external-links-performance-collector
WORKDIR /services/external-links-performance-collector
RUN pip install -r /services/external-links-performance-collector/requirements.txt
ENTRYPOINT ["python3.6","calculate_external_links_performance.py"]
