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
  const tprefix = "Prefixes/" + req.pathParams.toPrefix;
  const keys = db._query(`
    let k = (
      For e in LinkEdges
        Filter e.FromIP == @fip
        Return e.ToIP
    )
    For e in PrefixEdges
      Filter e.InterfaceIP == k[0] AND e._to == @tprefix
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
