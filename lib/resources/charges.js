var Resource = require('../resource');
var bank_accounts = require('./bank_accounts');
var cards = require('./cards');
var dates = require('../utils/dates');
var disputes = require('./disputes');
var extend = require('extend');
var flatten = require('../utils/flatten');

module.exports = Resource()
  .url('https://api.stripe.com/v1/charges')
  .collection('charges')
  .consume(function* (client, obj) {
    var properties = extend(
      {
        amount: obj.amount,
        amount_refunded: obj.amount_refunded,
        application_fee: obj.application_fee,
        balance_transaction: obj.balance_transaction,
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
        status: obj.status
      },
      flatten(obj, 'metadata'),
      flatten(obj, 'fraud_details'),
      flatten(obj, 'shipping'),
      dates(obj, { created: 'created' })
    );

    var source = obj.source;
    if (source) {
      // TODO: Add foreign key to `properties` or to `source` object
      // Consume all sources card or bank account objects
      if (source.object === 'card') {
        yield cards._consume(client, source);
      }
      if (source.object === 'bank_account') {
        yield bank_accounts._consume(client, source);
      }
    }


    var dispute = obj.dispute;
    if (dispute) {
      // Add dispute foreign key to charge object
      properties.dispute_id = dispute.id;
      // Consume as a dispute object
      disputes._consume(client, dispute);
    }

    yield client.set('charges', obj.id, properties);
  });
