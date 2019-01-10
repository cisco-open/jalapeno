# !/usr/bin/env python3
import os
import sys

#if sys.version > (3, 0):
#    import http.client
#else:
#    import httplib
import base64
import ssl

from flask import render_template
from flask import Flask, jsonify

app = Flask(__name__)


@app.route('/')
def login():
    return render_template('login.html')

@app.route('/overview')
def usecases():
    return render_template('overview.html')


@app.route('/demo')
def demo():
    return render_template('demo.html')


#@app.route('/docs')
#def docs():
#    return render_template('docs.html')


#@app.route('/usecases')
#def usecases():
#    return render_template('usecases.html')



#@app.route('/processing', methods=['GET'])
# Install the smartsheet sdk with the command: pip install smartsheet-python-sdk
#def processing():


if __name__ == '__main__':
    app.secret_key = os.urandom(12)
    port = int(os.environ.get('PORT', 5000))
    app.run(host='0.0.0.0', port=port, debug=True)
