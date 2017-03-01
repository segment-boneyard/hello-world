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

  log.info({ collection: resource._collection }, 'starting collection query ..');
  do {
    var response = yield executeWithRetry(resource._url, options, resource._headers);
    var body = response.body;

    for (let i=0; i<body.data.length; i++) {
      const obj = body.data[i];
      // subpaginate the object's lists *before* consumption
      // that way obj[subpage] is populated when resource#consume is called
      yield Promise.each(resource._subpages, co(function* (subpage) {
        obj[subpage] = yield depaginate(
          obj[subpage],
          clone(options),
          resource._headers
        );
      }));
      // execute resource
      const shouldContinue = yield resource._consume(client, obj, log, options);
      if (shouldContinue === false) {
        return;
      }
    }

    if (body.data.length > 0) {
      options.starting_after = body.data[body.data.length - 1].id;
    }
  } while (body.has_more);
  log.info({ collection: resource._collection }, 'collection finished');
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
