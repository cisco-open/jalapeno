FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/lsv6-performance-processor
WORKDIR /services/lsv6-performance-processor
RUN pip install -r /services/lsv6-performance-processor/requirements.txt
ENTRYPOINT ["python3.6","calculate_lsv6_performance.py"]
