FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/internal-links-performance-collector
WORKDIR /services/internal-links-performance-collector
RUN pip install -r /services/internal-links-performance-collector/requirements.txt
ENTRYPOINT ["python3.6","calculate_internal_links_performance.py"]
