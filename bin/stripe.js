var program = require('commander');
var Source = require('@segment/source');
var Stripe = require('../lib');
var conf = require('@segment/config');
var log = require('@segment/logger')(conf('logger'), { app: 'root/stripe' });
var co = require('co');

/**
 * Program.
 */

program
  .option('--writeKey <writekey>', 'set the segment write key')
  .option('--secret <secret>', 'set the stripe secret key')
  .option('--host <host>', 'set the segment API host')
  .parse(process.argv);

/**
 * Create the options
 * @type {Object}
 */

var options = { secret: secret };

// create the batch
var source = Source('stripe')
  .logger(log);

var run = source.run(program.writeKey)
  .host(program.host);

var stats = source.stats();
stats.memory('start'); // take an initial snapshot of memory

// create the data source
var stripe = Stripe()
  .logger(log);

// start the pull
co(function*() {
  return yield stripe.pull(run, options); 
}).then(function (value) {
  log.info('finished running data source successfully');
  snapshot();
}, function (err) {
  log.critical('data source failed.', { err: err.stack });
  snapshot();
});

var start = Date.now();
process.on('uncaughtException', function (err) {
  log.critical('uncaught exception', { stack: err.stack });
  snapshot();
});

function snapshot() {
  stats.memory('end'); // take the final snapshot of memory
  stats.count('total duration', Date.now() - start);
  log.info('stats snapshot', stats.snapshot());
}