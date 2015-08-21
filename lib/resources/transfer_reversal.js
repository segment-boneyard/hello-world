var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .object('transfer_reversal')
  .collection('transfer_reversal')
  .properties(function (obj) {
    return extend(
      {
        created: new Date(obj.created*1000),
        currency: obj.currency,
        refunded: obj.refunded,
        amount: obj.amount,
        balance_transaction_id: obj.balance_transaction,
        transfer_id: obj.transfer
      }, 
      flat(obj, 'metadata'),
    );
  });
