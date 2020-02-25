FROM python:2
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /services/epe-edges
COPY . /services/epe-edges
RUN pip install -r /services/epe-edges/requirements.txt
CMD ["python","create_epe_edges.py"]
