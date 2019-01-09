# Voltron Web UI
A web interface for Voltron to present some pretty capabilities and easy insights into Voltron.

## Development
> What are the steps to develop on this/get this working?

Instructions to run the application:

Run the following commands from the VoltronUI folder:

virtualenv -p python3 ./venv3

source ./venv3/bin/activate

pip install Flask

pip install pyaml

export FLASK_APP=flaskr.py

flask run


Then, go to 127.0.0.1:5000/demo in your browser to see it running.
