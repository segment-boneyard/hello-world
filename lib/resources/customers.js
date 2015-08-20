var Resource = require('../resource');
var flat = require('flat');
var extend = require('extend');

module.exports = Resource()
  .object('customers')
  .collection('customers')
  .properties(function (obj) {
    return extend(
      {
        "created": new Date(obj.created*1000),
        "description": obj.description,
        "email": obj.email,
        "delinquent": obj.delinquent,
        "discount": obj.discount,
        "account_balance": obj.account_balance,
        "currency": obj.currency
      }, 
      flat(obj.metadata, { delimiter: '_' })
    );
  });