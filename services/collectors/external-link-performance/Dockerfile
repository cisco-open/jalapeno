FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/epe-paths-bandwidth-openconfig-collector
WORKDIR /services/epe-paths-bandwidth-openconfig-collector
RUN pip install -r /services/epe-paths-bandwidth-openconfig-collector/requirements.txt
ENTRYPOINT ["python","calculate_epe_paths_bandwidth_openconfig.py"]
