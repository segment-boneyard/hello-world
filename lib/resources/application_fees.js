'use strict';

const co = require('bluebird').coroutine;
var Resource = require('../resource');
var applicationFeeRefunds = require('./application_fee_refunds');
var dates = require('../utils/dates');
var extend = require('extend');
var flatten = require('../utils/flatten');

module.exports = Resource()
  .url('https://api.stripe.com/v1/application_fees')
  .subpages(['refunds'])
  .collection('application_fees')
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        account_id: obj.account,
        amount: obj.amount,
        amount_refunded: obj.amount_refunded,
        application_id: obj.application,
        balance_transaction_id: obj.balance_transaction,
        charge_id: obj.charge,
        currency: obj.currency,
        originating_transaction: obj.originating_transaction,
        refunded: obj.refunded
      },
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    );

    var refunds = obj.refunds;
    if (refunds.length) {
      properties.refund_ids = refunds.map((r) => r.id).join(',');

      // Consume refunds as refund objects
      for (var refund of refunds) {
        refund.application_fee_id = obj.id;
        yield applicationFeeRefunds._consume(client, refund);
      }
    }

    yield client.set('application_fees', obj.id, properties);
  }));
