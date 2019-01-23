# Voltron Portal
A web interface for Voltron to present some pretty capabilities and easy insights into Voltron.

## Development
Voltron Portal requires Python, HTML/CSS, and JavaScript knowledge. It utilizes the [Voltron API](../api/) for displaying information about Voltron. Development code is located in `src/`.

### Running
Using Docker is the easiest way to run/test the Voltron Portal. It is important to note that not all files used for the Portal are included, namely the HTML/CSS styling framework. These are acquired on build.

```bash
docker build -t voltron/portal .
# Run Portal, exposing on port 5000
docker run -d -p 5000:80 voltron/portal
open http://localhost:5000/
```

### Libraries

#### Python
* [Flask](http://flask.pocoo.org/)
* [gevent](http://www.gevent.org/)

gevent is used as the production-oriented server when production-quality performance is desired. Using gevent allows greater scalability than just Flask alone. We disable debug mode for Flask when used with gevent.

#### HTML/CSS
The UI elements are based off of [creativetimofficial/argon-dashboard](https://github.com/creativetimofficial/argon-dashboard). The Dockerfile will acquire these files and copy them to `static/argon` for styling usage.
