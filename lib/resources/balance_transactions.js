var co = require('bluebird').coroutine;
var Resource = require('../resource');
var dates = require('../utils/dates');
var extend = require('extend');
var flatten = require('../utils/flatten');

module.exports = Resource()
  .url('https://api.stripe.com/v1/balance/history')
  .collection('balance_transactions')
  .consume(co(function* (client, obj) {
    var properties = extend(
      {
        amount: obj.amount,
        currency: obj.currency,
        description: obj.description,
        fee: obj.fee,
        net: obj.net,
        status: obj.status,
        type: obj.type
      },
      flatten(obj, 'metadata'),
      dates(obj, { created: 'created', available_on: 'available' })
    );

    yield client.set('balance_transactions', obj.id, properties);
  }));
