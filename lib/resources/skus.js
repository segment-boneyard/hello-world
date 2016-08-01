'use strict';

var co = require('bluebird').coroutine;
var Resource = require('../resource');
var extend = require('extend');
var flatten = require('../utils/flatten');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/skus')
  .collection('skus')
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        product_id: obj.product,
        active: obj.active,
        currency: obj.currency,
        image: obj.image,
        livemode: obj.livemode,
        price: obj.price
      },
      flatten(obj, 'attributes'),
      flatten(obj, 'inventory'),
      flatten(obj, 'package_dimensions'),
      flatten(obj, 'metadata'),
      dates(obj, {
        created: 'created',
        updated: 'updated'
      })
    );

    yield client.set('skus', obj.id, properties);
  }));
