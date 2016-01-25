var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');

module.exports = Resource()
  .collection('discounts')
  .consume(function (client, obj) {
    var id = obj.customer + '_' + obj.coupon.id;
    return client.set('discounts', id, extend({
        customer_id: obj.customer,
        coupon_id: obj.coupon.id,
        subscription: obj.subscription
      },
      dates(obj, {
        start: 'start',
        end: 'end'
      })
    ));
  });