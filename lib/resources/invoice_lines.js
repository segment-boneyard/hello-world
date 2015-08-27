var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var plans = require('./plans');

module.exports = Resource()
  .collection('invoice_lines')
  .consume(function* (client, obj) {
    yield client.set('invoice_lines', obj.id, extend({
        amount: obj.amount,
        type: obj.type,
        currency: obj.currency,
        proration: obj.proration,
        period_start: new Date(obj.period.start*1000),
        period_end: new Date(obj.period.end*1000),
        subscription_id: obj.subscription,
        quantity: obj.quantity,
        description: obj.description,
        discountable: obj.discountable,
      }, 
      flatten(obj, 'metadata')
    ));

    var plan = obj.plan;
    if (plan) plans._consume(client, plan);
  });
