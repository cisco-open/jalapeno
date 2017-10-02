'use strict';

const createRouter = require('@arangodb/foxx/router');
const router = createRouter();
module.context.use(router);

var joi = require("joi");
const db = require('@arangodb').db;
const aql = require('@arangodb').aql;

// Add Custom Queries here. path segments preficed by ':' are vars
// available in the req.pathParams object.
router.get('/:router/interfaces/ips', function (req, res) {
  var router = req.pathParams.router;
  const keys = db._query(`
      FOR e in LinkEdges
      FILTER e._from == @key
      RETURN DISTINCT e.FromIP`, {'key': "Routers/" + router}
    );
  res.send(keys);
})
.response(joi.array().items(
  joi.string().required()
).required(), 'List of router interface IPs.')
.summary('List of router interface IPs')
.description('Assembles a list of router interface ips');
