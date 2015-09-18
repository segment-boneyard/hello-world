var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/coupons')
  .collection('coupons')
  .consume(function* (client, obj) {
    yield client.set('coupons', obj.id, extend({
        percent_off: obj.percent_off,
        amount_off: obj.amount_off,
        currency: obj.currency,
        duration: obj.duration,
        max_redemptions: obj.max_redemptions,
        times_redeemed: obj.times_redeemed,
        valid: obj.valid,
        duration_in_months: obj.duration_in_months
      }, 
      flatten(obj, 'metadata'),
      dates(obj, { 
        created: 'created',
        redeem_by: 'redeem_by'
      })
    ));
  });