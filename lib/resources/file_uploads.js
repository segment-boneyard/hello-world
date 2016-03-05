'use strict';

var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');

module.exports = Resource()
  // 8/26/2015 - stripe 500's this endpoint
  // .url('https://uploads.stripe.com/v1/files')
  .collection('file_uploads')
  .consume(function(client, obj) {
    return client.set('file_uploads', obj.id, extend({
      size: obj.size,
      purpose: obj.purpose,
      type: obj.type
    },
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    ));
  });
