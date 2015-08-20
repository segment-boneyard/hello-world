
var Promise = require('bluebird');
var changecase = require('change-case');
var parallel = require('co-parallel');

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
  
  var tasks = [];
  for (var i = 0; i < res.sobjects.length; i += 1) {
    var sobject = res.sobjects[i];
    var name = sobject.name;

    tasks.push(this._query(run, resource));
  }

  yield parallel(tasks, 1);
  
  this.log.info('finished stripe pull', {  public: true });

  return Promise.resolve();
};

/**
 * Query the Stripe resource.
 * @param {Run} run         
 * @param {Resource} resource   
 * @yield {Promise}
 */

Stripe.prototype._query = function(run, resource) {
  var self = this;
  return new Promise(function (resolve, reject) { 
    var recordsRead = 0;
    var stats = run.stats();

    var parser = parse({ columns: true });

    var segment = run.stream(collection)
      .id(mapId)
      .properties(mapProperties);

    var stream = request
      .get(req)
      .on('error', salesforceError)
      .on('response', onresponse) // rejects promise
      .pipe(parser)
      .on('data', onrecord)
      .on('error', onresponse) // rejects promise
      .on('end', onend)
      .pipe(segment)
      .on('error', segmentError); // swallows error

    function onrecord(record) { recordsRead += 1; };

    function onresponse(res) {
      self.log.info('received bulk query result response', { code: res.statusCode });
    }

    function onend() {
      self.log.debug('salesforce read stream completed', { 
        collection: collection, 
        recordsRead: recordsRead,
        public: true
      });
      resolve();
    };

    function salesforceError(err) {
      stats.incr('salesforce.' + collection + '.error', 1);
      self.log.warn('salesforce bulk query error', { 
        err: err.stack, 
        jobId: r.jobId,
        batchId: r.batchId,
        resultId: r.id,
        public: true 
      });
      reject(err);
    };

    function csvError(err) {
      stats.incr('salesforce.' + collection + '.csv_error', 1);
      self.log.warn('csv parsing error', { 
        err: err.stack, 
        jobId: r.jobId,
        batchId: r.batchId,
        resultId: r.id,
        public: true 
      });
      reject(err);
    };

    function segmentError(err) {
      stats.incr('segment.' + collection + '.request_error', 1);
      self.log.warn('segment request error', { 
        err: err.stack, 
        jobId: r.jobId,
        batchId: r.batchId,
        resultId: r.id,
        public: true 
      });
      // swallow error for now
    };
  });
};

/**
 * Get a Stripe's `record` object id.
 * @param  {Object} record
 * @return {String}        
 */

function mapId(record) {
  return record.Id;
}

/**
 * Get a Stripe `record` to Segment object properties.
 * @param  {Object} record
 * @return {Object}
 */

function mapProperties(record) { 
  var props = {};
  Object.keys(record).forEach(function (key) {
    if (key !== 'Id') {
      var val = record[key];
      if (typeof val === 'string' && val.length > 65536) val = val.substring(0, 65536);
      props[changecase.snakeCase(key)] = val;
    }
  });
  return props;
}
