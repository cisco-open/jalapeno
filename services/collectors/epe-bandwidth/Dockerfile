FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/epe-paths-bandwidth-collector
WORKDIR /services/epe-paths-bandwidth-collector
RUN pip install -r /services/epe-paths-bandwidth-collector/requirements.txt
ENTRYPOINT ["python","calculate_epe_paths_bandwidth.py"]
