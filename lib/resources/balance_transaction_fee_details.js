'use strict';

var co = require('bluebird').coroutine;
var Resource = require('../resource');
var md5 = require('md5');

module.exports = Resource()
  .collection('balance_transaction_fee_details')
  .consume(co(function* (client, obj) {
    var id = md5([
      obj.amount,
      obj.currency,
      obj.application,
      obj.type,
      obj.balance_transaction_id
    ].filter(Boolean).join(', '));

    yield client.set('balance_transaction_fee_details', id, obj);
  }));
