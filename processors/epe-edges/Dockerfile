FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /services/epe-edges
COPY . /services/epe-edges
RUN pip3 install -r /services/epe-edges/requirements.txt
CMD ["python3.6","create_epe_edges.py"]
