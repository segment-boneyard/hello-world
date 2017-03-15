'use strict';

const TokenBucket = require('token-bucket-promise');
const defaults = require('defaults');
const client = require('@segment/source');
const request = require('co-request');
const co = require('bluebird').coroutine;
const service = require('./kit');
const parseUrl = require('url').parse;
const clone = require('lodash').clone;
const get = require('lodash').get;
const Promise = require('bluebird');


const config = service.config;
const log = service.log;

// log every 10 seconds
const checkpointInterval = 10000;

const rps = Number(config.rps) || 80;
const tbf = new TokenBucket(rps, 1000);
log.info({ rps }, 'Limiting source reqs/sec');

// depaginate the entire array by following the links
const depaginate = co(function* (body, options, headers) {
  var data = body.data;

  while (body.has_more) {
    options.starting_after = body.data[body.data.length - 1].id;
    var url = 'https://api.stripe.com' + body.url;
    var response = yield executeWithRetry(url, options, headers);
    body = response.body;
    data = data.concat(body.data);
  }
  return data;
});

// pull a specific resource from stripe
exports.fetch = co(function* (resource, extraOptions) {
  var options = Object.assign({ limit: 100 }, extraOptions, resource._options);

  const start = Date.now();

  // log every 10 seconds
  let nextCheckpoint = Date.now() + checkpointInterval;

  // metrics
  const metrics = {
    requests: 0, // total outer collection list requests
    objects: 0, // total outer objects processed
    requestTime: 0, // total time querying collection
    subpages: {}
  };

  log.info({ collection: resource._collection }, 'starting collection query ..');
  do {
    var requestStart = Date.now();
    var response = yield executeWithRetry(resource._url, options, resource._headers);

    metrics.requests += 1; // assume no retries here
    metrics.requestTime += Date.now() - requestStart;

    var body = response.body;

    for (let i=0; i<body.data.length; i++) {
      const obj = body.data[i];
      // subpaginate the object's lists *before* consumption
      // that way obj[subpage] is populated when resource#consume is called
      yield Promise.each(resource._subpages, co(function* (subpage) {
        const subpageRequestStart = Date.now();

        obj[subpage] = yield depaginate(
          obj[subpage],
          clone(options),
          resource._headers
        );

        if (!metrics.subpages[subpage]) {
          metrics.subpages[subpage] = {
            objectsWithDeepSubpages: 0, // outer objects requiring subpage requests
            subpageObjectsProcessed: 0, // total outer objects processed
            maxSubpageObjectsPerObject: 0, // total outer objects processed
            subpageRequests: 0, // total amount of deep subpage requests
            subpageRequestTimeElapsed: 0 // total amount of time spent querying subpages
          };
        }
        metrics.subpages[subpage].subpageRequestTimeElapsed += Date.now() - subpageRequestStart;
        metrics.subpages[subpage].subpageObjectsProcessed += obj[subpage].length;
        if (obj[subpage].length > metrics.subpages[subpage].maxSubpageObjectsPerObject) {
          metrics.subpages[subpage].maxSubpageObjectsPerObject = obj[subpage].length;
        }

        const subpageRequests = Math.floor(obj[subpage].length / options.limit);
        if (subpageRequests > 0) {
          metrics.subpages[subpage].objectsWithDeepSubpages += 1;
          metrics.subpages[subpage].subpageRequests += subpageRequests;
        }
      }));

      metrics.objects += 1;

      // execute resource
      const shouldContinue = yield resource._consume(client, obj, log, options);
      if (shouldContinue === false) {
        return;
      }
    }

    if (body.data.length > 0) {
      options.starting_after = body.data[body.data.length - 1].id;
    }

    if (Date.now() > nextCheckpoint) {
      log.info({
        collection: resource._collection,
        elapsed: Date.now() - start,
        metrics: metrics
      }, 'collection checkpoint');
      nextCheckpoint = Date.now() + checkpointInterval;
    }
  } while (body.has_more);
  log.info({
    collection: resource._collection,
    elapsed: Date.now() - start,
    metrics: metrics
  }, 'collection finished');
});

// retry a function with properties
// based off co-retry but with logging
const retry = co(function* (fn, options) {
  var attempts = options.retries + 1;
  var interval = options.interval;
  /* eslint-disable no-constant-condition */
  while (true) {
    /* eslint-enable no-constant-condition */
    try {
      return yield fn();
    } catch (err) {
      log.error(err);
      attempts--;
      if (!attempts || err.response && err.response.statusCode === 401) {
        throw err;
      }
      yield Promise.delay(interval);
      interval = interval * options.factor;
    }
  }
});

// execute the request with error handling
const execute = co(function* (url, options, headers) {
  const current = new Date();
  yield tbf.take();
  const response = yield request(req(url, options, headers));
  const parsedUrl = parseUrl(url);
  const baseTags = [`path:${parsedUrl.pathname}`];
  const statusCodeBucket = `${String(response.statusCode).charAt(0)}xx`;

  yield Promise.all([
    client.statsHistogram('http.response.payload_size', response.headers['content-length'], baseTags),
    client.statsHistogram('http.request.duration', new Date() - current, baseTags),
    client.statsIncrement('http.request.duration', new Date() - current, baseTags),
    client.statsIncrement('http.request.total', 1, baseTags),
    client.statsIncrement(`http.response.status_code.${statusCodeBucket}`, 1, [
      `status_code:${response.statusCode}`,
      `status_code_bucket:${statusCodeBucket}`
    ].concat(baseTags))
  ]);

  if (response.statusCode !== 200) {
    log.error(response, 'Bad Stripe response');
    const err = new Error('Bad Stripe response');
    err.response = response;
    if (response.statusCode === 401) {
      err.message = 'Invalid credentials';
      err.showToUser = true;
    }

    if (response.statusCode >= 500) {
      err.message = 'Internal error happened in Stripe API';
      err.showToUser = true;
    }

    if (get(response.body, '.error.code') === 'rate_limit') {
      err.message = 'Stripe API rate limit exceeded';
      err.showToUser = true;
    }

    throw err;
  }
  return response;
});

// execute the request with a few retries
function executeWithRetry(url, options, headers) {
  return retry(execute.bind(null, url, options, headers), {
    retries: 5, // retry 5 times
    interval: 5000, // wait interval first before retrying
    factor: 2 // multiple interval by factor every response to get backoff
  });
}

// create a new stripe request
function req(url, options, headers) {
  return {
    url: url,
    method: 'GET',
    headers: headers || {},
    auth: { user: config.secret, pass: '' },
    qs: defaults(options || {}, { limit: 100 }),
    json: true
  };
}
