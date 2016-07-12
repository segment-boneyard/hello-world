'use strict';

var consume = require('../utils/subscriptions').consume;
var Resource = require('../resource');

module.exports = Resource()
  .collection('subscriptions')
  .url('https://api.stripe.com/v1/subscriptions')
  .headers({ 'Stripe-Version': '2016-07-06' })
  .options({ status: 'canceled' })
  .consume(consume);
