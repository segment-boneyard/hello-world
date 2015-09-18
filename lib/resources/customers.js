var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var subscriptions = require('./subscriptions');
var cards = require('./cards');
var bank_accounts = require('./bank_accounts');
var discounts = require('./discounts');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/customers')
  .collection('customers')
  .subpages(['sources', 'subscriptions'])
  .consume(function* (client, obj) {
    yield client.set('customers', obj.id, extend({
        description: obj.description,
        email: obj.email,
        delinquent: obj.delinquent,
        account_balance: obj.account_balance,
        currency: obj.currency
      }, 
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created' })
    ));

    for (var i = 0; i < obj.subscriptions.length; i += 1) {
      var subscription = obj.subscriptions[i];
      yield subscriptions._consume(client, subscription);
    }

    for (var i = 0; i < obj.sources.length; i += 1) {
      var source = obj.sources[i];
      if (source.object === 'card') yield cards._consume(client, source);
      else if (source.object === 'bank_account') yield bank_accounts._consume(client, source);
    }

    var discount = obj.discount;
    if (discount) discounts._consume(client, discount);
  });