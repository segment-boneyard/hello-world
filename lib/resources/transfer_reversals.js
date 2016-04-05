'use strict';

var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var dates = require('../utils/dates');

module.exports = Resource()
  .collection('transfer_reversals')
  .consume(function(client, obj) {
    return client.set('transfer_reversals', obj.id, extend({
      amount: obj.amount,
      currency: obj.currency,
      balance_transaction_id: obj.balance_transaction,
      transfer_id: obj.transfer
    },
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    ));
  });
