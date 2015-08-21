var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .object('cards')
  .collection('cards')
  .properties(function (obj) {
    return extend(
      {
        last4: obj.last4,
        brand: obj.brand,
        funding: obj.funding,
        exp_month: obj.exp_month,
        exp_year: obj.exp_year,
        country: obj.country,
        name: obj.name,
        address_line1: obj.address_line1,
        address_line2: obj.address_line2,
        address_city: obj.address_city,
        address_state: obj.address_state,
        address_zip: obj.address_zip,
        address_country: obj.address_country,
        cvc_check: obj.cvc_check,
        address_line1_check: obj.address_line1_check,
        address_zip_check: obj.address_zip_check,
        tokenization_method: obj.tokenization_method,
        dynamic_last4: obj.dynamic_last4
      }, 
      flatten(obj, 'metadata')
    );
  });