var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .object('subscriptions')
  .collection('subscriptions')
  .properties(function (obj) {
    return extend(
      {
        plan_id: obj.plan.id,
        start: new Date(obj.start*1000),
        status: obj.status,
        customer_id: obj.customer,
        cancel_at_period_end: obj.cancel_at_period_end,
        current_period_start: new Date(obj.current_period_start*1000),
        current_period_end: new Date(obj.current_period_end*1000),
        ended_at: new Date(obj.ended_at*1000),
        trial_start: new Date(obj.trial_start*1000),
        trial_end: new Date(obj.trial_end*1000),
        canceled_at: new Date(obj.canceled_at*1000),
        quantity: obj.quantity,
        application_fee_percent: obj.application_fee_percent,
        discount: obj.discount,
        tax_percent: obj.tax_percent
      }, 
      flatten(obj, 'metadata')
    );
  });