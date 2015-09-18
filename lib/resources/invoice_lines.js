var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var plans = require('./plans');
var dates = require('../utils/dates');

module.exports = Resource()
  .collection('invoice_lines')
  .consume(function* (client, obj) {
    yield client.set('invoice_lines', obj.id, extend({
        amount: obj.amount,
        type: obj.type,
        currency: obj.currency,
        proration: obj.proration,
        subscription_id: obj.subscription,
        quantity: obj.quantity,
        description: obj.description,
        discountable: obj.discountable,
      }, 
      flatten(obj, 'metadata'),
      dates(obj.period, { 
        start: 'period_start',
        end: 'period_end'
      })
    ));

    var plan = obj.plan;
    if (plan) plans._consume(client, plan);
  });
