'use strict';

var Resource = require('../resource');
var dates = require('../utils/dates');
var discounts = require('./discounts');
var extend = require('extend');
var flatten = require('../utils/flatten');
var plans = require('./plans');

module.exports = Resource()
  .collection('subscriptions')
  .consume(function(client, obj) {
    var properties = extend(
      {
        application_fee_percent: obj.application_fee_percent,
        cancel_at_period_end: obj.cancel_at_period_end,
        customer_id: obj.customer,
        quantity: obj.quantity,
        status: obj.status,
        tax_percent: obj.tax_percent
      },
      flatten(obj, 'metadata'),
      dates(obj, {
        start: 'start',
        current_period_start: 'current_period_start',
        current_period_end: 'current_period_end',
        ended_at: 'ended_at',
        trial_start: 'trial_start',
        trial_end: 'trial_end',
        canceled_at: 'canceled_at'
      })
    );

    var plan = obj.plan;
    if (plan) {
      // Attach plan foreign key to subscription object
      properties.plan_id = plan.id;

      // Consume as plan object
      plans._consume(client, plan);
    }

    var discount = obj.discount;
    if (discount) {
      // Attach discount foreign key to subscription object
      properties.discount_id = `${ discount.customer }_${ discount.coupon.id }`;

      // Consume as plan object
      discounts._consume(client, discount);
    }

    return client.set('subscriptions', obj.id, properties);
  });
