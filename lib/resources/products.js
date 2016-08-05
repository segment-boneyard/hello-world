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
        attributes: obj.attributes,
        caption: obj.caption,
        deactivate_on: obj.deactivate_on,
        description: obj.description,
        images: obj.images, // might want to use delimiter other than ','
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

    properties.attributes = properties.attributes || properties.attributes.join(',');
    properties.deactivate_on = properties.deactivate_on || properties.deactivate_on.join(',');
    properties.images = properties.images || properties.images.join(',');

    yield client.set('products', obj.id, properties);
  }));
