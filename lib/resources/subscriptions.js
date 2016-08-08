'use strict';

var discounts = require('../resources/discounts');
var plans = require('../resources/plans');
var flatten = require('../utils/flatten');
var Resource = require('../resource');
var dates = require('../utils/dates');
var extend = require('extend');

module.exports = Resource()
  .collection('subscriptions')
  .url('https://api.stripe.com/v1/subscriptions')
  .headers({ 'Stripe-Version': '2016-07-06' })
  .options({ status: 'all' })
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
        created: 'created',
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
