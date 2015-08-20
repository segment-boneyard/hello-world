
/**
 * Export `HarvestClient`.
 * @type {HarvestClient}
 */

module.exports = Resource;

/**
 * Create a new `Resource`.
 * @param {string} object
 */

function Resource() {
  if (!(this instanceof Resource)) return new Resource();
  // default to the Stripe pattern
  this.id(function (obj) { return obj.id; });
}

/**
 * Set the resource `object`.
 * @param  {String} object 
 * @return {Resource}     
 */

Resource.prototype.object = function (object) {
  this._object = object;
  this.collection(object);
  return this;
};

/**
 * Set the resource `collection`.
 * @param  {String} collection 
 * @return {Resource}     
 */

Resource.prototype.collection = function (collection) {
  this._collection = collection;
  return this;
};

/**
 * Set the id mapping function.
 * @param  {Function} mapId
 * @return {Resource}     
 */

Resource.prototype.mapId = function (fn) {
  this._mapId = fn;
  return this;
};

/**
 * Set the property mapping function.
 * @param  {Function} fn
 * @return {Resource}     
 */

Resource.prototype.mapProperties = function (fn) {
  this._mapProperties = fn;
  return this;
};
