'use strict';

var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/refunds')
  .collection('refunds')
  .consume(function(client, obj) {
    return client.set('refunds', obj.id, extend({
      amount: obj.amount,
      currency: obj.currency,
      balance_transaction_id: obj.balance_transaction,
      charge_id: obj.charge,
      receipt_number: obj.receipt_number,
      reason: obj.reason
    },
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    ));
  });
