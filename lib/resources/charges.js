var Resource = require('../resource');
var flat = require('flat');
var extend = require('extend');

module.exports = Resource()
  .object('charges')
  .collection('charges')
  .properties(function (obj) {
    return extend(
      {
        "created": new Date(obj.created*1000),
        "paid": obj.paid,
        "status": obj.status,
        "amount": obj.amount,
        "currency": obj.currency,
        "refunded": obj.refunded,
        "source_id": obj.source.id,
        "captured": obj.captured,
        "balance_transaction": obj.balance_transaction,
        "failure_message": obj.failure_message,
        "failure_code": obj.failure_code,
        "amount_refunded": obj.amount_refunded,
        "customer": obj.customer,
        "invoice": obj.invoice,
        "description": obj.description,
        "dispute_id": obj.dispute.id,
        "statement_descriptor": obj.statement_descriptor,
        "receipt_email": obj.receipt_email,
        "receipt_number": obj.receipt_number,
        "destination": obj.destination,
        "application_fee": obj.application_fee
      }, 
      flat(obj.metadata, { delimiter: '_' }),
      flat(obj.fraud_details, { delimiter: '_' }),
      flat(obj.shipping, , { delimiter: '_' })
    );
  });