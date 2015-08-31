
var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .url('https://api.stripe.com/v1/bitcoin/receivers')
  .collection('bitcoin_receivers')
  .consume(function* (client, obj) {
    yield client.set('bitcoin_receivers', obj.id, extend({
        created: new Date(obj.created*1000),
        active: obj.active,
        amount: obj.amount,
        amount_received: obj.amount_received,
        bitcoin_amount: obj.bitcoin_amount,
        bitcoin_amount_received: obj.bitcoin_amount_received,
        currency: obj.currency,
        filled: obj.filled,
        uncaptured_funds: obj.uncaptured_funds,
        description: obj.description,
        email: obj.email,
        refund_address: obj.refund_address,
        used_for_payment: obj.used_for_payment
      }, 
      flatten(obj, 'metadata')
    ));
  });