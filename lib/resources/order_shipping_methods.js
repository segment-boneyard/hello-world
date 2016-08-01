'use strict';

var co = require('bluebird').coroutine;
var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var md5 = require('md5');

module.exports = Resource()
  .collection('order_shipping_methods')
  .consume(co(function* (client, obj) {
    // obj.id is only unique within the order so we have to generate hash
    // for the primary key

    var id = md5([
      obj.id,
      obj.order_id
    ].filter(Boolean).join(', '));

    var properties = extend(
      {
        order_id: obj.order_id,
        shipping_id: obj.id,
        amount: obj.amount,
        currency: obj.currency,
        description: obj.description
      },
      flatten(obj, 'delivery_estimate')
    );

    yield client.set('order_shipping_methods', id, properties);
  }));
