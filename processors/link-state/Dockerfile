FROM python:3.6
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /processors/ls
COPY . /processors/ls
RUN pip3 install -r /processors/ls/requirements.txt
CMD ["python3.6","create_ls_topology.py"]
