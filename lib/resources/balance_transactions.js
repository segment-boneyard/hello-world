'use strict';

var Promise = require('bluebird');
var Resource = require('../resource');
var co = require('bluebird').coroutine;
var dates = require('../utils/dates');
var extend = require('extend');
var feeDetails = require('./balance_transaction_fee_details');
var flatten = require('../utils/flatten');
var _ = require('lodash');

/*
  This resource is pulled in two modes:
  - all transactions (options.transfer == undefined):
      in this mode we stop after reaching a 10-day old transaction
  - transactions filtered by transfer
      paginates until the end
*/

const stopAfter = new Date().valueOf() / 1000 - 86400 * 10;

module.exports = Resource()
  .url('https://api.stripe.com/v1/balance/history')
  .collection('balance_transactions')
  .consume(co(function* (client, obj, log, options) {
    if (options.transfer === undefined && obj.created < stopAfter) {
      return false;
    }

    var properties = extend(
      {
        amount: obj.amount,
        currency: obj.currency,
        description: obj.description,
        fee: obj.fee,
        net: obj.net,
        status: obj.status,
        type: obj.type,
        source: obj.source
      },
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created', available_on: 'available' })
    );

    if (options.transfer !== undefined) {
      properties.transfer_id = options.transfer;
    }

    if (_.get(obj, 'sourced_transfers.data.length')) {
      properties.sourced_transfers = _.map(
        obj.sourced_transfers.data,
        (transfer) => transfer.id
      ).join(',');
    }

    if (_.get(obj, 'fee_details.length')) {
      yield Promise.each(obj.fee_details, function(detail) {
        detail.balance_transaction_id = obj.id;
        return feeDetails._consume(client, detail);
      });
    }

    yield client.set('balance_transactions', obj.id, properties);
  }));
