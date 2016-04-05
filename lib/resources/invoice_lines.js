'use strict';

var Resource = require('../resource');
var dates = require('../utils/dates');
var extend = require('extend');
var flatten = require('../utils/flatten');
var plans = require('./plans');
var md5 = require('md5');
var startsWith = require('underscore.string/startsWith');

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

    if (startsWith(obj.id, 'sub_')) {
      properties.subscription_id = obj.id;
    } else if (startsWith(obj.id, 'ii_')) {
      properties.item_id = obj.id;
    }

    var hashedId = md5([
      obj.id,
      properties.type,
      properties.period_start,
      properties.period_end
    ].filter((v) => v).join(', '));

    return client.set('invoice_lines', hashedId, properties);
  });
