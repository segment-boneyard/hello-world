var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .url('https://api.stripe.com/v1/accounts')
  .collection('accounts')
  .consume(function (client, obj) {
    var properties = extend({
        email: obj.email,
        statement_descriptor: obj.statement_descriptor,
        display_name: obj.display_name,
        timezone: obj.timezone,
        details_submitted: obj.details_submitted,
        charges_enabled: obj.charges_enabled,
        transfers_enabled: obj.transfers_enabled,
        default_currency: obj.default_currency,
        country: obj.country,
        business_name: obj.business_name,
        business_url: obj.business_url,
        support_phone: obj.support_phone,
        business_logo: obj.business_logo,
        support_url: obj.support_url,
        support_email: obj.support_email,
        managed: obj.managed,
        product_description: obj.product_description,
        debit_negative_balances: obj.debit_negative_balances
      },
      flatten(obj, 'metadata'),
      flatten(obj, 'support_address')
    );

    if (obj.currencies_supported)
      properties.currencies_supported = obj.currencies_supported.join(',');

    return client.set('accounts', obj.id, properties);
  });
