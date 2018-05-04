FROM python:2
MAINTAINER Zia Syed <ziausyed@cisco.com>
WORKDIR /services/path
COPY . /services/path
RUN pip install -r /services/path/requirements.txt
CMD ["python","create_paths.py"]
