'use strict';

const co = require('bluebird').coroutine;
var Resource = require('../resource');
var application_fee_refunds = require('./application_fee_refunds');
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
      // TODO: Add foreign key to `properties` or to `refund` object
      // Consume refunds as refund objects
      for (var refund of refunds) {
        yield application_fee_refunds._consume(client, refund);
      }
    }

    yield client.set('application_fees', obj.id, properties);
  }));
