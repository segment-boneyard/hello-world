'use strict';

var Promise = require('bluebird');
var co = require('bluebird').coroutine;
var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');
// var orderItems = require('./order_items');
var orderShippingMethods = require('./order_shipping_methods');
var _ = require('lodash');

module.exports = Resource()
  .url('https://api.stripe.com/v1/orders')
  .collection('orders')
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        amount: obj.amount,
        amount_returned: obj.amount_returned,
        application: obj.application,
        application_fee: obj.application_fee,
        charge_id: obj.charge,
        currency: obj.currency,
        customer_id: obj.customer,
        email: obj.email,
        livemode: obj.livemode,
        selected_shipping_method: obj.selected_shipping_method,
        status: obj.status
      },
      flatten(obj, 'metadata'),
      flatten(obj, 'shipping'),
      flatten(obj, 'status_transitions'),
      dates(obj, {
        created: 'created',
        updated: 'updated'
      })
    );

    /* commented out until there's found a reliable way of generating primary keys for items

    if (_.get(obj, 'items.length')) {
      yield Promise.each(obj.items, function(item) {
        item.order_id = obj.id;
        return orderItems._consume(client, item);
      });
    }
    */

    if (_.get(obj, 'shipping_methods.length')) {
      yield Promise.each(obj.shipping_methods, function(method) {
        method.order_id = obj.id;
        return orderShippingMethods._consume(client, method);
      });
    }

    yield client.set('orders', obj.id, properties);
  }));
