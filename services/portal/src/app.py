# !/usr/bin/env python3
import os
from gevent.pywsgi import WSGIServer
from flask import Flask, render_template


topology_endpoint = os.environ.get('API_NETLOC')
app = Flask(__name__)

@app.route('/')
def index():
    return render_template('index.html', topology_endpoint=topology_endpoint)

def start_prod():
    http_server = WSGIServer(('', 80), app)
    http_server.serve_forever()

def start_dev():
    app.run(host='0.0.0.0', port=80, debug=True, threaded=True)

if __name__ == '__main__':
    if 'FLASK_ENV' not in os.environ:
        start_prod()
    elif os.environ['FLASK_ENV'] == 'development':
        start_dev()
    else:
        start_prod()
