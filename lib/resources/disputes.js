var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .url('https://api.stripe.com/v1/disputes')
  .collection('disputes')
  .consume(function* (client, obj) {
    yield client.set('disputes', obj.id, extend({
        created: new Date(obj.created*1000),
        charge_id: obj.charge,
        amount: obj.amount,
        status: obj.status,
        currency: obj.currency,
        reason: obj.reason,
        is_charge_refundable: obj.is_charge_refundable
      }, 
      flatten(obj, 'evidence_details'),
      flatten(obj, 'evidence'),
      flatten(obj, 'metadata')
    ));
  });
