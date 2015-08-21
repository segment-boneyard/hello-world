var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .object('refunds')
  .collection('refunds')
  .properties(function (obj) {
    return extend(
      {
        "amount": obj.amount,
        "currency": obj.currency,
        "created": new Date(obj.created*1000),
        "balance_transaction": obj.balance_transaction,
        "charge_id": obj.charge,
        "receipt_number": obj.receipt_number,
        "reason": obj.reason
      }, 
      flatten(obj, 'metadata')
    );
  });