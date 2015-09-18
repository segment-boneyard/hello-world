var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/plans')
  .collection('plans')
  .consume(function* (client, obj) {
    yield client.set('plans', obj.id, extend({
        interval: obj.interval,
        name: obj.name,
        amount: obj.amount,
        currency: obj.currency,
        interval_count: obj.interval_count,
        trial_period_days: obj.trial_period_days,
        statement_descriptor: obj.statement_descriptor
      }, 
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    ));
  });