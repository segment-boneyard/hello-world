var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .collection('application_fee_refunds')
  .consume(function* (client, obj) {
    yield client.set('application_fee_refunds', obj.id, extend({
        amount: obj.amount,
        currency: obj.currency,
        created: new Date(obj.created*1000),
        balance_transaction: obj.balance_transaction,
        fee_id: obj.fee_id
      }, 
      flatten(obj, 'metadata')
    ));
  });