FROM python:2.7.14-alpine
MAINTAINER Aishwarya <aishsree@cisco.com>
COPY requirements.txt /tmp
RUN pip install -r /tmp/requirements.txt
RUN rm /tmp/requirements.txt
COPY . /services/french-press
WORKDIR /services/french-press
ENTRYPOINT ["python","xr-data-collector.py","--kafkaEndpoint","10.86.204.8:9092","--arangodbEndpoint","10.86.204.8:30852"]
