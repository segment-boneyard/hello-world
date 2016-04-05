'use strict';

/**
 * Map the dates in the `obj` to a new object using the `schema`.
 * For example:
 *   map({ modifiedDate: 1441314078006 }, { modifiedDate: 'modified_date' })
 *   > { "modified_date": ISODate("Thu Sep 03 2015 14:01:06 GMT-0700") }
 * @param  {Object} obj source object
 * @param  {Object} schema schema used in the mapping
 * @return {Function} toDate funciton used to map the dates
 */

module.exports = function dates(obj, schema) {
  var res = {};
  // we want to map only specific attributes
  Object.keys(schema).forEach(function(from) {
    var to = schema[from];

    if (obj[from]) {
      var date = toDate(obj[from]);
      res[to] = date;
    }
  });

  return res;
};

// Stripe date format
function toDate(secs) {
  return new Date(secs*1000);
}
