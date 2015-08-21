var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');

module.exports = Resource()
  .object('tokens')
  .collection('tokens')
  .properties(function (obj) {
    return extend(
      {
        created: new Date(obj.created*1000),
        used: obj.used,
        type: obj.card,
        card_id: obj.card.id,
        client_ip: obj.client_ip
      }, 
      flatten(obj, 'metadata')
    );
  });