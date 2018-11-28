FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/internal-links-performance-collector
WORKDIR /services/internal-links-performance-collector
RUN pip install -r /services/internal-links-performance-collector/requirements.txt
ENTRYPOINT ["python","calculate_internal_links_performance.py"]
