var Resource = require('../resource');
var bank_accounts = require('./bank_accounts');
var cards = require('./cards');
var dates = require('../utils/dates');
var discounts = require('./discounts');
var extend = require('extend');
var flatten = require('../utils/flatten');
var subscriptions = require('./subscriptions');

module.exports = Resource()
  .url('https://api.stripe.com/v1/customers')
  .collection('customers')
  .subpages(['sources', 'subscriptions'])
  .consume(function* (client, obj) {
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
      // TODO: Add foreign key to `properties` or to `refund` object
      // Consume as subscription objects
      for (var subscription of subs) {
        yield subscriptions._consume(client, subscription);
      }
    }

    var sources = obj.sources;
    if (sources.length) {
      // TODO: Add foreign key to `properties` or to `source` object
      // Consume all sources card or bank account objects
      for (var source of sources) {
        if (source.object === 'card') {
          yield cards._consume(client, source);
        }
        if (source.object === 'bank_account') {
          yield bank_accounts._consume(client, source);
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
  });
