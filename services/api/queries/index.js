'use strict';

const createRouter = require('@arangodb/foxx/router');
const router = createRouter();
module.context.use(router);

var joi = require("joi");
const db = require('@arangodb').db;
const aql = require('@arangodb').aql;

// This should probably be implemented as a POST? or ?& query?
router.get("/linkedges/:from/:to/:latency", function(req, res) {
  const fip = req.pathParams.from;
  const tip = req.pathParams.to;
  const latency = parseInt(req.pathParams.latency);
  if (latency == NaN) {
    res.throw(400, 'Provided Latency is Not a Number');
    return
  }
  const keys = db._query(`
    FOR e in LinkEdges
    FILTER e.FromIP == @fip AND e.ToIP == @tip
      UPDATE {
        _key: e._key,
        Latency: @latency
      } in LinkEdges`, {"fip": fip, "tip": tip, "latency": latency}
    );
  res.send(keys);
}).description("Update a LinkEdge (:from->:to) with a latency value.");

router.get("/latency/:fromIP/:toPrefix/:latency", function(req, res) {
  const fip = req.pathParams.fromIP;
  const tprefix = "_Prefixes_" + req.pathParams.toPrefix.replace(/\//g, '_');
  const latency = parseInt(req.pathParams.latency);
  if (latency == NaN) {
    res.throw(400, 'Provided Latency is Not a Number');
    return
  }
  const keys = db._query(`
    For e in LinkEdges
        Filter e.FromIP == @fip
        let key = CONCAT(SUBSTITUTE(e._to, "/", "_"), @tprefix)
        Update {
            _key: key,
            Latency: @latency
          } in PrefixEdges
        Return key`, {"fip": fip, "tprefix": tprefix, "latency": latency});
  res.send(keys);
}).description("Update a PrefixEdge with a latency value given the IP of an internal router connected to the external BGP peer. ");

router.get("/latency/:fromIP/:toPrefix", function(req, res) {
  const fip = req.pathParams.fromIP;
  const tprefix = "_Prefixes_" + req.pathParams.toPrefix.replace(/\//g, '_');
  const keys = db._query(`
    let k = (
      For e in LinkEdges
        Filter e.FromIP == @fip
        let key = CONCAT(SUBSTITUTE(e._to, "/", "_"), @tprefix)
        Return key
    )
    For e in PrefixEdges
      Filter e._key == k[0]
      Return e.Latency
    `,{"fip": fip, "tprefix": tprefix});
  res.send(keys);
}).description("Get a latency value given the IP of an internal router connected to the external BGP peer. ");

// This should probably be implemented as a POST? or ?& query?
router.get("/prefixedges/:fromIP/:toPrefix/:latency", function(req, res) {
  const fip = req.pathParams.fromIP;
  const tprefix = "Prefixes/" + req.pathParams.toPrefix;
  const latency = parseInt(req.pathParams.latency);
  if (latency == NaN) {
    res.throw(400, 'Provided Latency is Not a Number');
    return
  }
  const keys = db._query(`
    FOR e in PrefixEdges
    FILTER e.InterfaceIP == @fip AND e._to == @tprefix
      UPDATE {
        _key: e._key,
        Latency: @latency
      } in PrefixEdges`, {"fip": fip, "tprefix": tprefix, "latency": latency}
    );
  res.send(keys);
}).description("Update a PrefixEdges (:fromIP->:toPrefix) with a latency value.");


// Add Latency to full Paths
router.get("/paths/:source/:labelStack/:destination/:latency", function(req, res) {
  const source = req.pathParams.source;
  const labelstack = req.pathParams.labelStack;
  const destination = req.pathParams.destination;
  const latency = parseInt(req.pathParams.latency);
  if (latency == NaN) {
    res.throw(400, 'Provided Latency is Not a Number');
    return
  }
  const keys = db._query(`
      FOR p in Paths
    FILTER p.Source == @source AND p.Label_Path == @labelstack AND p.Destination == @destination
      UPDATE p with { 
          Latency : @latency 
      } in Paths`, {"source": source, "labelstack": labelstack, "destination": destination, "latency": latency}
    );
  res.send(keys);
}).summary('Set the latency for a complete path given its source, destination, and label stack')
.description("Update a complete path from source to destination with a latency value. The label stack should be in the following format: 24004_24001_24011.");


// Get the lowest latency label stack
router.get("/:source/:destination/latency", function(req, res) {
  const source = req.pathParams.source;
  const destination = req.pathParams.destination;
  const keys = db._query(`
      FOR p IN Paths
      FILTER p.Source == @source AND p.Destination == @destination
    SORT p.Latency
    LIMIT 1
    RETURN [p.Path, p.Label_Path]`, {"source": source, "destination": destination}
    );
  res.send(keys);
}).summary("Get the lowest latency path and its label stack given a source and destination.")
.description("Get the lowest latency path and its label stack given a source and destination.");

// Get the lowest utilization (highest available bandwidth) label stack
router.get("/:source/:destination/bandwidth", function(req, res) {
  if(req.pathParams.source == '10.1.2.1')
      req.pathParams.source = '10.1.1.0'
  const source = "Routers/" + req.pathParams.source;
  const destination = "Prefixes/" + req.pathParams.destination;  
  const keys = db._query(`
      FOR v,e IN OUTBOUND SHORTEST_PATH @source TO @destination
    GRAPH "topology"
    OPTIONS {weightAttribute: 'Bandwidth', defaultWeight: 200000000000000}
    FILTER e.Label != null
    RETURN e.Label`, {"source": source, "destination": destination}
    );
  res.send(keys);
}).summary("Get the lowest utilization (highest available bandwidth) path and its label stack given a source and destination.")
.description("Get the lowest utilization (highest available bandwidth) path and its label stack given a source and destination.");


// This should probably be implemented as a POST? or ?& query?
router.get("/linkedges/:from/:to", function(req, res) {
  const fip = req.pathParams.from;
  const tip = req.pathParams.to;
  const keys = db._query(`
    FOR e in LinkEdges
    FILTER e.FromIP == @fip AND e.ToIP == @tip
    RETURN e.Latency`, {"fip": fip, "tip": tip}
    );
  res.send(keys);
}).description("Get latency of linkedge");

router.get("/prefixedges/:fromIP/:toPrefix", function(req, res) {
  const fip = req.pathParams.fromIP;
  const tprefix = "Prefixes/" + req.pathParams.toPrefix;
  const latency = parseInt(req.pathParams.latency);
  if (latency == NaN) {
    res.throw(400, 'Provided Latency is Not a Number');
    return
  }
  const keys = db._query(`
    FOR e in PrefixEdges
    FILTER e.InterfaceIP == @fip AND e._to == @tprefix
    RETURN e.Latency`, {"fip": fip, "tprefix": tprefix}
    );
  res.send(keys);
}).description("Get latency of prefixedge");



const host2PrefixMap = {"10.11.0.1": "10.11.0.0_24", "10.12.0.1": "10.12.0.0_24", "10.13.0.1": "10.13.0.0_24"};
// Add Custom Queries here. path segments preficed by ':' are vars
// available in the req.pathParams object.
router.get('/labels/:router/:host', function (req, res) {
  var router = req.pathParams.router;
  var host = req.pathParams.host;
  if (host in host2PrefixMap) {
    host = host2PrefixMap[host];
  }
  if (host.indexOf("_24") === -1) {
    host = host + "_24";
  }
  const keys = db._query(`
    FOR v,e
    IN OUTBOUND SHORTEST_PATH
    @router TO @host
GRAPH 'topology'
OPTIONS {weightAttribute: "Latency", defaultWeight: 1000}
return e.Label
`, {'router': "Routers/" + router, 'host': "Prefixes/" + host });
  var labels = [];
  const docs = keys["_documents"];
  for (var k in docs) {
    if (!(docs[k] == null || docs[k] == "null" || docs[k] == undefined || docs[k] == "")){
      labels.push(docs[k]);
    }
  }
  res.send(labels);
}).description('Gets label stack for router->host.');


// Add Custom Queries here. path segments preficed by ':' are vars
// available in the req.pathParams object.
router.get('/edges/:router/ips', function (req, res) {
  var router = req.pathParams.router;
  const keys = db._query(`
      FOR e in LinkEdges
      FILTER e._from_ip == @key
      RETURN DISTINCT e.FromIP`, {'key': "Routers/" + router}
    );
  res.send(keys);
})
.response(joi.array().items(
  joi.string().required()
).required(), 'List of router interface IPs.')
.summary('List of router interface IPs')
.description('Assembles a list of router interface ips');

