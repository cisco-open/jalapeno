FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /infra/telemetry
COPY . /infra/telemetry
RUN pip install -r /infra/telemetry/requirements.txt
CMD ["python","deploy_telemetry.py"]
