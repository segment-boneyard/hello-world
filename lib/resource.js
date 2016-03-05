'use strict';

/**
 * Export `Resource`.
 * @type {Resource}
 */

module.exports = Resource;

/**
 * Create a new `Resource`.
 * @param {string} object
 */

function Resource() {
  if (!(this instanceof Resource)) return new Resource();
  this.subpages([]);
}

/**
 * Set the resource `url`.
 * @param  {String} url
 * @return {Resource}
 */

Resource.prototype.url = function(url) {
  this._url = url;
  return this;
};


/**
 * Set the resource `collection`.
 * @param  {String} collection
 * @return {Resource}
 */

Resource.prototype.collection = function(collection) {
  this._collection = collection;
  return this;
};

/**
 * Set the resource `subpages`.
 * @param  {Array|String} subpages
 * @return {Resource}
 */

Resource.prototype.subpages = function(subpages) {
  this._subpages = subpages;
  return this;
};

/**
 * Set the consumer function.
 * @param  {Function} fn
 * @return {Resource}
 */

Resource.prototype.consume = function(fn) {
  this._consume = fn;
  return this;
};
