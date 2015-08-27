var program = require('commander');
var client = require('@segment/source')();
var conf = require('@segment/config');
var log = require('@segment/logger')(conf('logger'), { app: 'source/stripe' });
var co = require('co');
var resources = require('../lib/resources');
var Promise = require('bluebird');
var parallel = require('co-parallel');
var request = require('co-request');
var defaults = require('defaults');
var extend = require('extend');

// command line
program
  .option('--secret <secret>', 'set the stripe secret key')
  .parse(process.argv);

// pull from stripe
function* pull() {
  var tasks = [];
  var keys = Object.keys(resources);
  for (var i = 0; i < keys.length; i += 1) {
    var resource = resources[keys[i]];
    if (resource._url) {
      // if its a top level resource (customers, plans, etc)
      // nested resources (cards, subscriptions) should not be listed
      tasks.push(fetch(resource));
    }
  }
  yield parallel(tasks, 10);
  return Promise.resolve();
}

// pull a specific resource from stripe
function* fetch(resource) {
  var options = { limit: 100 };
  log.info('starting collection query ..', { collection: resource._collection });
  do {
    var response = yield request(req(resource._url, options));
    check(resource._url, response);
    var body = response.body;

    for (var i = 0; i < body.data.length; i += 1) {
      var obj = body.data[i];
      // subpaginate the object's lists
      for (var j = 0; j < resource._subpages.length; j += 1) {
        var subpage = resource._subpages[j]; 
        obj[subpage] = yield depaginate(obj[subpage]);
      }
      yield resource._consume(client, obj);
    }

    if (body.data.length > 0)
      options.starting_after = body.data[body.data.length - 1].id;
  } while (body.has_more);
  log.info('collection finished', { collection: resource._collection });
}

// depaginate the entire array by following the links
function* depaginate(body) {
  var data = body.data;
  var options = { limit: 100 };
  while (body.has_more) {
    options.starting_after = body.data[body.data.length - 1].id;
    var url = 'https://api.stripe.com' + body.url;
    var response = yield request(req(url, options));
    check(url, response);
    body = response.body;
    data = data.concat(body.data);
  }
  return data;
}

// create a new stripe request
function req(url, options) {
  return { 
    url: url, 
    method: 'GET',
    auth: { user: program.secret, pass: '' },
    qs: defaults(options || {}, { limit: 100 }),
    json: true
  };
}

// check the request status
function check(url, response) {
  if (response.statusCode !== 200) {
    throw new Error('Bad Stripe response [' + response.statusCode + '] [' 
      + url + ']: ' + JSON.stringify(response.body));
  }
}

// start the program
co(start).then(onSuccess).catch(onError);

// start the program
function* start() {
   try {
    return yield pull(); 
  } catch (e) {
    log.critical('application crashed', { err: e.stack });
    throw e;
  }
}

// on success
function onSuccess() {
  log.info('finished running data source successfully');
}

// on error
function onError() {
  log.critical('data source failed.', { err: err.stack });
}

// global error logger
process.on('uncaughtException', function (err) {
  log.critical('uncaught exception', { stack: err.stack });
});