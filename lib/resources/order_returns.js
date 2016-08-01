'use strict';

// var Promise = require('bluebird');
var co = require('bluebird').coroutine;
var Resource = require('../resource');
var extend = require('extend');
var dates = require('../utils/dates');
// var orderReturnItems = require('./order_return_items');
// var _ = require('lodash');

module.exports = Resource()
  .url('https://api.stripe.com/v1/order_returns')
  .collection('order_returns')
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        amount: obj.amount,
        currency: obj.currency,
        description: obj.description,
        parent_id: obj.parent,
        quantity: obj.quantity,
        type: obj.type,
        livemode: obj.livemode,
        order_id: obj.order,
        refund_id: obj.refund
      },
      dates(obj, { created: 'created' })
    );

    /* commented out until there's found a reliable way of generating primary keys for items

    if (_.get(obj, 'items.length')) {
      yield Promise.each(obj.items, function(item) {
        item.order_return_id = obj.id;
        return orderReturnItems._consume(client, item);
      });
    }
    */

    yield client.set('order_returns', obj.id, properties);
  }));
