'use strict';

const createRouter = require('@arangodb/foxx/router');
const router = createRouter();
module.context.use(router);

var joi = require("joi");
const db = require('@arangodb').db;
const aql = require('@arangodb').aql;

// Get the lowest latency label stack
router.get("/:source/:destination/latency", function(req, res) {
  const source = req.pathParams.source;
  const destination = req.pathParams.destination;
  const keys = db._query(`
      FOR p IN EPEPaths_Latency
      FILTER p.Source == @source AND p.Destination == @destination
    SORT p.Latency
    LIMIT 1
    RETURN [p._key, p.Label_Path]`, {"source": source, "destination": destination}
    );
  res.send(keys);
}).summary("Get the lowest latency EPEPath and its label stack given a source and destination.")
.description("Get the lowest latency EPEPath and its label stack given a source and destination.");

// Get the lowest utilization (highest available bandwidth) label stack
router.get("/:destination/bandwidth", function(req, res) {
  const destination = req.pathParams.destination;  
  const keys = db._query(`
      FOR e in EPEPaths_Bandwidth
    FILTER e.Destination == @destination
    SORT e.Bandwidth
    LIMIT 1
    RETURN [e._key, e.Label_Path]`, {"destination": destination}
    );
  res.send(keys);
}).summary("Get the lowest utilization (highest available bandwidth) EPEPath and its label stack given a destination.")
.description("Get the lowest utilization (highest available bandwidth) EPEPath and its label stack given a destination.");

// Get the lowest utilization (highest available bandwidth based on openconfig telemetry) label stack
router.get("/:destination/bandwidth_openconfig", function(req, res) {
  const destination = req.pathParams.destination;  
  const keys = db._query(`
      FOR e in EPEPaths_Bandwidth_OpenConfig
    FILTER e.Destination == @destination
    SORT e.Bandwidth
    LIMIT 1
    RETURN [e._key, e.Label_Path]`, {"destination": destination}
    );
  res.send(keys);
}).summary("Get the lowest utilization (highest available bandwidth based on OpenConfig telemetry) EPEPath and its label stack given a destination.")
.description("Get the lowest utilization (highest available bandwidth based on OpenConfig telemetry) EPEPath and its label stack given a destination.");







