'use strict';

const co = require('bluebird').coroutine;
var Resource = require('../resource');
var bank_accounts = require('./bank_accounts');
var dates = require('../utils/dates');
var extend = require('extend');
var flatten = require('../utils/flatten');
var reversals = require('./transfer_reversals');

module.exports = Resource()
  .url('https://api.stripe.com/v1/transfers')
  .subpages(['reversals'])
  .collection('transfers')
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        amount: obj.amount,
        amount_reversed: obj.amount_reversed,
        application_fee: obj.application_fee,
        balance_transaction_id: obj.balance_transaction,
        currency: obj.currency,
        description: obj.description,
        destination_id: obj.destination_id,
        failure_code: obj.failure_code,
        failure_message: obj.failure_message,
        reversed: obj.reversed,
        source_transaction: obj.source_transaction,
        statement_descriptor: obj.statement_descriptor,
        status: obj.status,
        type: obj.type
      },
      flatten(obj, 'metadata'),
      dates(obj, {
        created: 'created',
        date: 'date'
      })
    );

    var bankAccount = obj.bank_account;
    if (bankAccount) {
      // Add discount foreign key to invoice object
      properties.bank_account_id = bankAccount.id;

      // Consume as bank_account object
      yield bank_accounts._consume(client, bankAccount);
    }

    if (obj.reversals.length) {
      // TODO: Add foreign key to `properties` or to `reversal` object
      // Consume as reversal objects
      for (var reversal of obj.reversals) {
        yield reversals._consume(client, reversal);
      }
    }

    yield client.set('transfers', obj.id, properties);
  }));
