'use strict';

var Resource = require('../resource');
var dates = require('../utils/dates');
var extend = require('extend');
var flatten = require('../utils/flatten');
var plans = require('./plans');

module.exports = Resource()
  .collection('invoice_lines')
  .consume(function(client, obj) {
    var properties = extend(
      {
        amount: obj.amount,
        currency: obj.currency,
        description: obj.description,
        discountable: obj.discountable,
        proration: obj.proration,
        quantity: obj.quantity,
        subscription_id: obj.subscription,
        type: obj.type
      },
      flatten(obj, 'metadata'),
      dates(obj.period, {
        start: 'period_start',
        end: 'period_end'
      })
    );

    var plan = obj.plan;
    if (plan) {
      // Add plan foreign key to invoice line object
      properties.plan_id = plan.id;
      // Consume as a plan object
      plans._consume(client, plan);
    }

    return client.set('invoice_lines', obj.id, properties);
  });
