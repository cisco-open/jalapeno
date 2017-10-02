'use strict';

const createRouter = require('@arangodb/foxx/router');
const router = createRouter();
module.context.use(router);

var joi = require("joi");
const db = require('@arangodb').db;
const aql = require('@arangodb').aql;

// This should probably be implemented as a POST? or ?& query?
router.get("/latency/:from/:to/:latency", function(req, res) {
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
}).description("Update an Edge (:from->:to) with a latency value.");

// This should probably be implemented as a POST? or ?& query?
router.get("/latency/:from/:to", function(req, res) {
  const fip = req.pathParams.from;
  const tip = req.pathParams.to;
  const keys = db._query(`
    FOR e in LinkEdges
    FILTER e.FromIP == @fip AND e.ToIP == @tip
    RETURN e.Latency`, {"fip": fip, "tip": tip}
    );
  res.send(keys);
}).description("Get latency of edge");


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
