FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/lsv4-performance-processor
WORKDIR /services/lsv4-performance-processor
RUN pip install -r /services/lsv4-performance-processor/requirements.txt
ENTRYPOINT ["python3.6","calculate_lsv4_performance.py"]
