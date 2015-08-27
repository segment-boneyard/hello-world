var Resource = require('../resource');
var extend = require('extend');
var flatten = require('../utils/flatten');
var cards = require('./cards');
var bank_accounts = require('./bank_accounts');
var disputes = require('./disputes');

module.exports = Resource()
  .url('https://api.stripe.com/v1/charges')
  .collection('charges')
  .consume(function* (client, obj) {
    yield client.set('charges', obj.id, extend({
        created: new Date(obj.created*1000),
        paid: obj.paid,
        status: obj.status,
        amount: obj.amount,
        currency: obj.currency,
        refunded: obj.refunded,
        captured: obj.captured,
        balance_transaction: obj.balance_transaction,
        failure_message: obj.failure_message,
        failure_code: obj.failure_code,
        amount_refunded: obj.amount_refunded,
        customer_id: obj.customer,
        invoice_id: obj.invoice,
        description: obj.description,
        statement_descriptor: obj.statement_descriptor,
        receipt_email: obj.receipt_email,
        receipt_number: obj.receipt_number,
        destination: obj.destination,
        application_fee: obj.application_fee
      }, 
      flatten(obj, 'metadata'),
      flatten(obj, 'fraud_details'),
      flatten(obj, 'shipping')
    ));

    var source = obj.source;
    if (source.object === 'card') yield cards._consume(client, source);
    else if (source.object === 'bank_account') yield bank_accounts._consume(client, source);

    var dispute = obj.dispute;
    if (dispute) disputes._consume(client, dispute);
  });