var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var plans = require('./plans');

module.exports = Resource()
  .url('https://api.stripe.com/v1/invoiceitems')
  .collection('invoice_items')
  .consume(function* (client, obj) {
    yield client.set('invoice_items', obj.id, extend({
        date: new Date(obj.date*1000),
        amount: obj.amount,
        proration: obj.proration,
        currency: obj.currency,
        customer_id: obj.customer,
        discountable: obj.discountable,
        description: obj.description,
        invoice_id: obj.invoice,
        subscription_id: obj.subscription,
        quantity: obj.quantity,
        plan_id: obj.plan.id,
        period_start: new Date(obj.period.start*1000),
        period_end: new Date(obj.period.end*1000)
      },
      flatten(obj, 'metadata')
    ));

    var plan = obj.plan;
    if (plan) plans._consume(client, plan);
  });