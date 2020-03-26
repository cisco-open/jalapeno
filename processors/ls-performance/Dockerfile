FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/ls-performance-processor
WORKDIR /services/ls-performance-processor
RUN pip install -r /services/ls-performance-processor/requirements.txt
ENTRYPOINT ["python3.6","calculate_ls_performance.py"]
