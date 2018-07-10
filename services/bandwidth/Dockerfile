FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/bandwidth
WORKDIR /services/bandwidth
RUN pip install -r /services/bandwidth/requirements.txt
ENTRYPOINT ["python","bandwidth_service.py"]
