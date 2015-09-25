var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');
var discounts = require('./discounts');

module.exports = Resource()
  .collection('subscriptions')
  .consume(function* (client, obj) {
    var properties = extend({
        status: obj.status,
        customer_id: obj.customer,
        cancel_at_period_end: obj.cancel_at_period_end,
        quantity: obj.quantity,
        application_fee_percent: obj.application_fee_percent,
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

    if (obj.plan) properties.plan_id = obj.plan.id;
    
    var discount = obj.discount;
    if (discount) {
      discounts._consume(client, discount);
      properties.discount_id = discount.customer + '_' + discount.coupon.id;
    }

    yield client.set('subscriptions', obj.id, properties);
  });