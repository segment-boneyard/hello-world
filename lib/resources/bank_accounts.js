var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .collection('bank_accounts')
  .consume(function* (client, obj) {
    yield client.set('bank_accounts', obj.id, extend({
        country: obj.country,
        currency: obj.currnecy,
        status: obj.status,
        bank_name: obj.bank_name,
        default_for_currency: obj.default_for_currency
      }, 
      flatten(obj, 'metadata')
    ));
  });