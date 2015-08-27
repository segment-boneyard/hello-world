var Resource = require('../resource');
var extend = require('extend');
var flatten = require('../utils/flatten');
var bank_accounts = require('./bank_accounts');
var reversals = require('./transfer_reversals');

module.exports = Resource()
  .url('https://api.stripe.com/v1/transfers')
  .subpages(['reversals'])
  .collection('transfers')
  .consume(function* (client, obj) {
    yield client.set('transfers', obj.id, extend({
        created: new Date(obj.created*1000),
        date: new Date(obj.date*1000),
        amount: obj.amount,
        currency: obj.currency,
        status: obj.status,
        reversed: obj.reversed,
        type: obj.type,
        balance_transaction_id: obj.balance_transaction,
        destination_id: obj.destination_id,
        description: obj.description,
        failure_message: obj.failure_message,
        failure_code: obj.failure_code,
        amount_reversed: obj.amount_reversed,
        statement_descriptor: obj.statement_descriptor,
        application_fee: obj.application_fee,
        source_transaction: obj.source_transaction
      }, 
      flatten(obj, 'metadata')
    ));

    var bank_account = obj.bank_account;
    if (bank_account) yield bank_accounts._consume(client, bank_account);

    for (var i = 0; i < obj.reversals.length; i += 1) {
      var reversal = obj.reversals[i];
      yield reversals._consume(client, reversal);
    }
  });