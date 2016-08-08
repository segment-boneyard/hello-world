'use strict';

const co = require('bluebird').coroutine;
var Resource = require('../resource');
var bankAccounts = require('./bank_accounts');
var cards = require('./cards');
var dates = require('../utils/dates');
var discounts = require('./discounts');
var extend = require('extend');
var flatten = require('../utils/flatten');
var map = require('lodash').map;

module.exports = Resource()
  .url('https://api.stripe.com/v1/customers')
  .collection('customers')
  .subpages(['sources', 'subscriptions'])
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        account_balance: obj.account_balance,
        currency: obj.currency,
        delinquent: obj.delinquent,
        description: obj.description,
        email: obj.email
      },
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    );

    var subs = obj.subscriptions;
    if (subs.length) {
      properties.subscription_ids = map(subs, (s) => s.id).join(',');
    }

    var sources = obj.sources;
    if (sources.length) {
      properties.source_ids = map(sources, (s) => s.id).join(',');

      // Consume all sources card or bank account objects
      for (var source of sources) {
        source.customer_id = obj.id;

        if (source.object === 'card') {
          yield cards._consume(client, source);
        } else if (source.object === 'bank_account') {
          yield bankAccounts._consume(client, source);
        }
      }
    }

    var discount = obj.discount;
    if (discount) {
      // Add discount foreign key to invoice object
      properties.discount_id = `${ discount.customer }_${ discount.coupon.id }`;

      // Consume as a discount object
      discounts._consume(client, discount);
    }

    yield client.set('customers', obj.id, properties);
  }));
