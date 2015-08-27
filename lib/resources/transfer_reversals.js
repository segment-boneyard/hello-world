var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .collection('transfer_reversals')
  .consume(function* (client, obj) {
    var id = obj.customer + '_' + obj.coupon.id;
    yield client.set('transfer_reversals', id, extend({
        created: new Date(obj.created*1000),
        amount: obj.amount,
        currency: obj.currency,
        balance_transaction: obj.balance_transaction,
        transfer_id: obj.transfer
      }, 
      flatten(obj, 'metadata')
    ));
  });