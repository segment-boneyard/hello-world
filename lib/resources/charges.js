'use strict';

const co = require('bluebird').coroutine;
var Resource = require('../resource');
var dates = require('../utils/dates');
var extend = require('extend');
var flatten = require('../utils/flatten');

module.exports = Resource()
  .url('https://api.stripe.com/v1/charges')
  .collection('charges')
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        amount: obj.amount,
        amount_refunded: obj.amount_refunded,
        application_fee: obj.application_fee,
        balance_transaction_id: obj.balance_transaction,
        captured: obj.captured,
        currency: obj.currency,
        customer_id: obj.customer,
        description: obj.description,
        destination: obj.destination,
        failure_code: obj.failure_code,
        failure_message: obj.failure_message,
        invoice_id: obj.invoice,
        paid: obj.paid,
        receipt_email: obj.receipt_email,
        receipt_number: obj.receipt_number,
        refunded: obj.refunded,
        statement_descriptor: obj.statement_descriptor,
        status: obj.status,
        tokenization_method: obj.tokenization_method
      },
      flatten(obj, 'metadata'),
      flatten(obj, 'fraud_details'),
      flatten(obj, 'shipping'),
      dates(obj, { created: 'created' })
    );

    var source = obj.source;
    if (source) {
      if (source.object === 'card') {
        properties.card_id = source.id;
      } else if (source.object === 'bank_account') {
        properties.bank_account_id = source.id;
      }
    }

    var dispute = obj.dispute;
    if (dispute) {
      // Add dispute foreign key to charge object
      properties.dispute_id = dispute.id;
    }

    yield client.set('charges', obj.id, properties);
  }));
