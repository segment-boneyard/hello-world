'use strict';

var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');

module.exports = Resource()
  .collection('application_fee_refunds')
  .consume(function(client, obj) {
    return client.set('application_fee_refunds', obj.id, extend({
      amount: obj.amount,
      currency: obj.currency,
      balance_transaction_id: obj.balance_transaction,
      fee_id: obj.fee
    },
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    ));
  });
