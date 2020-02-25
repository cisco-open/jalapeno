FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /services/epe-paths
COPY . /services/epe-paths
RUN pip install -r /services/epe-paths/requirements.txt
CMD ["python3.6","create_paths.py"]
