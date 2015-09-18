var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var application_fee_refunds = require('./application_fee_refunds');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/application_fees')
  .subpages(['refunds'])
  .collection('application_fees')
  .consume(function* (client, obj) {
    yield client.set('application_fees', obj.id, extend({
        amount: obj.amount,
        currency: obj.currency,
        refunded: obj.refunded,
        amount_refunded: obj.amount_refunded,
        balance_transaction_id: obj.balance_transaction,
        account_id: obj.account,
        application_id: obj.application,
        charge_id: obj.charge,
        originating_transaction: obj.originating_transaction
      }, 
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    ));

    for (var i = 0; i < obj.refunds.length; i += 1) {
      var refund = obj.refunds[i];
      yield application_fee_refunds._consume(client, refund);
    }
  });