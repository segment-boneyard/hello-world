var Resource = require('../resource');
var flat = require('flat');
var extend = require('extend');

module.exports = Resource()
  .object('invoice_items')
  .collection('invoice_items')
  .properties(function (obj) {
    return extend(
      {
        date: new Date(obj.date*1000),
        amount: obj.amount,
        proration: obj.proration,
        currency: obj.currency,
        customer_id: obj.customer,
        discountable: obj.discountable,
        description: obj.description,
        invoice_id: obj.invoice,
        subscription_id: obj.subscription,
        quantity: obj.quantity,
        plan_id: obj.plan.id,
        period_start: new Date(obj.period.start),
        period_end: new Date(obj.period.end)
      },
      flatten(obj, 'metadata')
    );
  });