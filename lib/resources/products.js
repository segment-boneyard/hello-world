'use strict';

var co = require('bluebird').coroutine;
var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/products')
  .collection('products')
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        active: obj.active,
        attributes: obj.attributes.join(','),
        caption: obj.caption,
        deactivate_on: obj.deactivate_on.join(','),
        description: obj.description,
        images: obj.images.join(','), // might want to use delimiter other than ','
        livemode: obj.livemode,
        name: obj.name,
        shippable: obj.shippable,
        url: obj.url
      },
      flatten(obj, 'metadata'),
      flatten(obj, 'package_dimensions'),
      dates(obj, {
        created: 'created',
        updated: 'updated'
      })
    );

    yield client.set('products', obj.id, properties);
  }));
