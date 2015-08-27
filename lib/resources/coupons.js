var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .url('https://api.stripe.com/v1/coupons')
  .collection('coupons')
  .consume(function* (client, obj) {
    yield client.set('coupons', obj.id, extend({
        percent_off: obj.percent_off,
        amount_off: obj.amount_off,
        created: new Date(obj.created*1000),
        currency: obj.currency,
        duration: obj.duration,
        redeem_by: new Date(obj.redeem_by*1000),
        max_redemptions: obj.max_redemptions,
        times_redeemed: obj.times_redeemed,
        valid: obj.valid,
        duration_in_months: obj.duration_in_months
      }, 
      flatten(obj, 'metadata')
    ));
  });