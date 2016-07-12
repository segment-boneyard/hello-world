'use strict';

var consume = require('../utils/subscriptions').consume;
var Resource = require('../resource');

module.exports = Resource()
  .collection('subscriptions')
  .consume(consume);
