
var flat = require('flat');
var changeCase = require('change-case');

/**
 * Map the `obj` `key` using the optional `schema`. If `schema`
 * is not provided, then map all the attributes.
 * For example:
 *   map({ source: { id: 'foo' }}, 'source', { id: 'id' })
 *   > { "metadata_id": "foo" }
 * @param  {Object} obj source object
 * @param  {String} key source object key
 * @param  {Object} schema optional schema used in the mapping
 * @return {Object} mapped object
 */

module.exports = function map (obj, key, schema) {
  if (obj[key]) {
    var res = {};
    if (schema) {
       // we want to map only specific attributes
      Object.keys(schema).forEach(function (schemaKey) {
        res[key + '_' + changeCase.snakeCase(schemaKey)] = obj[key][schema[schemaKey]];
      });
    } else {
      // we want to map all the attributes
      if (Object.keys(obj[key]).length > 0) {
        // there are keys in the object
        res[changeCase.snakeCase(key)] = obj[key];
        res = flat(res, { delimiter: '_' })
      }
    }
    return res;
  }
};