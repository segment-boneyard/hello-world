var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var plans = require('./plans');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/invoiceitems')
  .collection('invoice_items')
  .consume(function* (client, obj) {
    var properties = extend({
        amount: obj.amount,
        proration: obj.proration,
        currency: obj.currency,
        customer_id: obj.customer,
        discountable: obj.discountable,
        description: obj.description,
        invoice_id: obj.invoice,
        subscription_id: obj.subscription,
        quantity: obj.quantity
      },
      flatten(obj, 'metadata'),
      dates(obj, { date: 'date' }),
      dates(obj.period, { 
        start: 'period_start',
        end: 'period_end'
      })
    );

    if (obj.plan) properties['plan_id'] = obj.plan.id;
    
    yield client.set('invoice_items', obj.id, properties);

    var plan = obj.plan;
    if (plan) plans._consume(client, plan);
  });
