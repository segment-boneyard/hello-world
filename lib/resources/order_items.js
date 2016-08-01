'use strict';

/* commented out until there's found a reliable way of generating primary keys for items

var co = require('bluebird').coroutine;
var Resource = require('../resource');
var md5 = require('md5');

module.exports = Resource()
  .collection('order_items')
  .consume(co(function* (client, obj) {
    var id = md5([
      obj.order_id,
      obj.amount,
      obj.currency,
      obj.description,
      obj.parent,
      obj.quantity,
      obj.type
    ].filter(Boolean).join(', '));

    var properties = {
      order_id: obj.order_id,
      amount: obj.amount,
      currency: obj.currency,
      description: obj.description,
      parent_id: obj.parent,
      quantity: obj.quantity,
      type: obj.type
    };

    yield client.set('order_items', id, properties);
  }));
*/