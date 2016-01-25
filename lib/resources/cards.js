var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .collection('cards')
  .consume(function (client, obj) {
    return client.set('cards', obj.id, extend(
      {
        address_city: obj.address_city,
        address_country: obj.address_country,
        address_line1: obj.address_line1,
        address_line1_check: obj.address_line1_check,
        address_line2: obj.address_line2,
        address_state: obj.address_state,
        address_zip: obj.address_zip,
        address_zip_check: obj.address_zip_check,
        brand: obj.brand,
        country: obj.country,
        cvc_check: obj.cvc_check,
        exp_month: obj.exp_month,
        exp_year: obj.exp_year,
        funding: obj.funding,
        name: obj.name,
        tokenization_method: obj.tokenization_method,
        type: obj.type
      },
      flatten(obj, 'metadata')
    ));
  });
