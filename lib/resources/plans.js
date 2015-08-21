var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .object('plans')
  .collection('plans')
  .properties(function (obj) {
    return extend(
      {
        interval: obj.interval,
        name: obj.name 
        created: new Date(obj.created*1000),
        amount: obj.amount,
        currency: obj.currency,
        interval_count: obj.interval_count,
        trial_period_days: obj.trial_period_days
      }, 
      flatten(obj, 'metadata')
    );
  });


