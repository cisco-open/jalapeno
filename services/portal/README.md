# Jalapeno Portal
A web interface for Jalapeno to present some pretty capabilities and easy insights.

## Development
Jalapeno Portal requires Python, HTML/CSS, and JavaScript knowledge. It utilizes the [Jalapeno API](../api/) for displaying information. Development code is located in `src/`.

### Running
Using Docker is the easiest way to run/test the Jalapeno Portal. It is important to note that not all files used for the Portal are included, namely the HTML/CSS styling framework. These are acquired on build.

```bash
docker build -t jalapeno/portal .
# Run Portal, exposing on port 5000
docker run -d -p 5000:80 jalapeno/portal
open http://localhost:5000/
```

### Libraries

#### Python
* [Flask](http://flask.pocoo.org/)
* [gevent](http://www.gevent.org/)

gevent is used as the production-oriented server when production-quality performance is desired. Using gevent allows greater scalability than just Flask alone. We disable debug mode for Flask when used with gevent.

#### HTML/CSS
The UI elements are based off of [creativetimofficial/argon-dashboard](https://github.com/creativetimofficial/argon-dashboard). The Dockerfile will acquire these files and copy them to `static/argon` for styling usage.
