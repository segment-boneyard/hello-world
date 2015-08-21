var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .object('discounts')
  .collection('discounts')
  .properties(function (obj) {
    return extend(
      {
        "coupon_id": obj.coupon.id,
        "start": new Date(obj.start*1000),
        "customer_id": obj.customer,
        "subscription_id": obj.subscription,
        "end": new Date(obj.end*1000)
      }
    );
  });

