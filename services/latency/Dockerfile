FROM python:2.7.14-alpine
MAINTAINER Mike Napolitano <mnapolit@cisco.com>
COPY . /services/latency
WORKDIR /services/latency
RUN pip install -r /services/latency/requirements.txt
ENTRYPOINT ["python","xr-data-collector.py"]
