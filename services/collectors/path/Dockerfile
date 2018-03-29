FROM python:2.7.14-alpine
MAINTAINER Zia Syed <ziausyed@cisco.com>
COPY . /services/path
WORKDIR /services/path
RUN pip install -r /services/path/requirements.txt
ENTRYPOINT ["python","create_paths.py"]
