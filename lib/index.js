
var Promise = require('bluebird');
var parallel = require('co-parallel');
var Stripe = require('stripe');
var resources = require('./resources');

/**
 * Expose `Stripe`
 */

module.exports = Stripe;

/**
 * Create a new `Stripe` instance.
 */

function Stripe() {
   if (!(this instanceof Stripe)) return new Stripe();
};

/**
 * Set the stats.
 */

Stripe.prototype.stats = function(stats) {
  this._stats = stats;
  return this;
};

/**
 * Set the logger.
 */

Stripe.prototype.logger = function(logger) {
  this.log = logger;
  return this;
};

/**
 * Pull from the Stripe integration.
 *
 * @param {Run} run
 * @param {Object} options
 * @returns {Promise} promise
 */

Stripe.prototype.pull = function*(run, options) {
  this.log.info('starting stripe pull ..', { runId: run.id, timestamp: new Date() })
  
  var stripe = Stripe(options.secret);

  var tasks = [];
  var objects = Object.keys(resources);
  for (var i = 0; i < objects.length; i += 1) {
    var object = objects[i];
    var resource = resources[object];

    tasks.push(this._query(stripe, run, resource));
  }

  yield parallel(tasks, 10);
  
  this.log.info('finished stripe pull', { public: true });

  return Promise.resolve();
};

/**
 * Query the Stripe resource.
 * @param {Stripe} stripe         
 * @param {Run} run         
 * @param {Resource} resource   
 * @yield {Promise}
 */

Stripe.prototype._query = function(stripe, run, resource) {
  var self = this;
  return new Promise(function (resolve, reject) { 
    var objectsRead = 0;
    var collection = resource._collection;

    var segment = run.stream(collection)
      .id(resource._mapId)
      .properties(resource._mapProperties);

    resource.stream(self, stripe)
      .on('error', stripeError) // rejects promise
      .on('data', onObject)
      .pipe(segment)
      .on('flush error', segmentError) // swallows error
      .on('finish', onFinish);

    function onObject(obj) {
      objectsRead += 1; 
    };

    function onFinish() {
      self.log.debug('stripe read stream completed', { 
        collection: collection, 
        objectsRead: objectsRead,
        public: true
      });

      segment.flush()
        .then(resolve)
        .catch(resolve);
    };

    function stripeError(err) {
      self._stats.incr('stripe.' + collection + '.error', 1);
      self.log.warn('stripe query error', { 
        err: err.stack, collection: collection, public: true 
      });
      reject(err);
    };

    function segmentError(err) {
      self._stats.incr('segment.' + collection + '.request_error', 1);
      self.log.warn('segment request error', { 
        err: err.stack, collection: collection, public: true 
      });
    };
  });
};
