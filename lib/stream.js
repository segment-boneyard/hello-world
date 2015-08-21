
var Readable = require('stream').Readable;
var Promise = require('bluebird');
var defaults = require('defaults');
var fmt = require('util').format;

/**
 * Export `ResourceStream`.
 * @type {ResourceStream}
 */

module.exports = ResourceStream;

/**
 * Create a new `ResourceStream`
 * @param {HarvestClient} client
 * @param {String} name 
 * @param {String} url  
 */

function ResourceStream(source, stripe, resource) {
  this._source = source;
  this._stripe = stripe;
  this._resource = resource;
  this.stats = this._source._stats;
  this.log = this._source.log;
  Readable.call(this, { objectMode: true });
}

// inherit from stream.prototype
ResourceStream.prototype = Object.create(
  Readable.prototype, { constructor: { value: ResourceStream }}
);

/**
 * Read into the object stream.
 */

ResourceStream.prototype._read = function () {
  var self = this;
  
  if (!this._page) {
    // if there's no cursor, let's start one at the beggining
    this._startingAfter = null;
    this._page = 1;
    this.log.debug('set first page cursor', { 
      object: this._resource._object,
      page: this._page 
    });
  }

  if (this._page && !this._promise) {
    var options = {};
    if (this._startingAfter) options.starting_after = this._startingAfter;

    this._promise = this._query(options)
      .delay(this._delay())
      .then(this._onres.bind(this))
      .catch(function (err) { self.emit('error', err); });
  }
};

/**
 * Calculate the delay before hitting the API, used for rate limiting.
 * @return {Number}
 */
ResourceStream.prototype._delay = function () {
  // TODO: add rate limiting
  return 0;
};

/**
 * Query Greenhouse API for a specific object list page.
 * @param  {Object} options 
 * @return {Promise}         
 */

ResourceStream.prototype._query = function (options) {
  options = defaults(options, { limit: 100 });
  var self = this;
  return new Promise(function (resolve, reject) {
    var object = self._resource._object;
    var opts = { query: options, object: object };
    var start = Date.now();

    self.log.debug('querying stripe ..', opts);

    self._stripe[object].list(options, function (err, res) {
      opts.duration = Date.now() - start;

      if (err) { 
        var msg = fmt('stripe request failed: %s', err.stack);
        self.stats.incr('stripe.query.' + object + '.failed', 1);
        self.log.error(msg, opts);
        return reject(err);
      } 

      self.stats.incr('stripe.query.' + object + '.success', 1);
      self.log.debug('query request successful', opts);
      resolve(res);
    });
  });
};

/**
 * On response handler to a request to one of their sources.
 * @param  {Response} res
 */

ResourceStream.prototype._onres = function onres(res) {
  var self = this;

  this.log.debug('received list response', {
    items: res.data.length,
    page: this._page,
    hasMore: res.has_more
  });

  if (res.has_more) {
    this._page += 1;
    this._promise = null;
    this._startingAfter = res.data[res.data.length - 1].id;
    this.log.debug('set next page cursor', { 
      page: this._page,
      startingAfter: this._startingAfter
    });
  }

  // this must come after the _page statement
  // to indicate to the stream that it should call read(..)
  // again whenever its ready
  res.data.forEach(function (object) { 
    self.push(object); 
  });

  if (!res.has_more) {
    // we're on the last page now, let's end the stream
    this.push(null);
    this.log.debug('last page, ending stream');
  }
};

