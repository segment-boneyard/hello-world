var Resource = require('../resource');
var extend = require('extend');
var flatten = require('../utils/flatten');

module.exports = Resource()
  .object('charges')
  .collection('charges')
  .properties(function (obj) {
    return extend(
      {
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
      flatten(obj, 'shipping'),
      flatten(obj, 'source', { id: 'id', object: 'object'}),
      flatten(obj, 'dispute', { id: 'id' })
    );
  });