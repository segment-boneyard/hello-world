'use strict';

var Resource = require('../resource');
var dates = require('../utils/dates');
var extend = require('extend');
var flatten = require('../utils/flatten');
var plans = require('./plans');

module.exports = Resource()
  .url('https://api.stripe.com/v1/invoiceitems')
  .collection('invoice_items')
  .consume(function(client, obj) {
    var properties = extend(
      {
        amount: obj.amount,
        currency: obj.currency,
        customer_id: obj.customer,
        description: obj.description,
        discountable: obj.discountable,
        invoice_id: obj.invoice,
        proration: obj.proration,
        quantity: obj.quantity,
        subscription_id: obj.subscription
      },
      flatten(obj, 'metadata'),
      dates(obj, { date: 'date' }),
      dates(obj.period, {
        start: 'period_start',
        end: 'period_end'
      })
    );

    var plan = obj.plan;
    if (plan) {
      // Attach plan foreign key to invoice item object
      properties.plan_id = plan.id;

      // Consume as plan object
      plans._consume(client, plan);
    }

    return client.set('invoice_items', obj.id, properties);
  });
