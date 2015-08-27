var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .collection('discounts')
  .consume(function* (client, obj) {
    var id = obj.customer + '_' + obj.coupon.id;
    yield client.set('discounts', id, extend({
        customer_id: obj.customer,
        coupon_id: obj.coupon.id,
        start: new Date(obj.start*1000),
        end: new Date(obj.end*1000),
        subscription: obj.subscription
      }
    ));
  });