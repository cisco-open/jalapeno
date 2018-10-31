FROM python:2
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /services/epe-paths
COPY . /services/epe-paths
RUN pip install -r /services/epe-paths/requirements.txt
CMD ["python","create_paths.py"]
