'use strict';

var Resource = require('../resource');
var dates = require('../utils/dates');
var extend = require('extend');

module.exports = Resource()
  .collection('discounts')
  .consume(function(client, obj) {
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
