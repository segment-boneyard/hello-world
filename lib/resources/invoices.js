var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var discounts = require('./discounts');
var invoice_lines = require('./invoice_lines');

module.exports = Resource()
  .url('https://api.stripe.com/v1/invoices')
  .subpages(['lines'])
  .collection('invoices')
  .consume(function* (client, obj) {
    yield client.set('invoices', obj.id, extend({
        date: new Date(obj.date*1000),
        period_start: new Date(obj.period_start*1000),
        period_end: new Date(obj.period_end*1000),
        subtotal: obj.subtotal,
        total: obj.total,
        customer_id: obj.customer,
        attempted: obj.attempted,
        closed: obj.closed,
        forgiven: obj.forgiven,
        paid: obj.paid,
        attempt_count: obj.attempt_count,
        amount_due: obj.amount_due,
        currency: obj.currency,
        starting_balance: obj.starting_balance,
        ending_balance: obj.ending_balance,
        next_payment_attempt: new Date(obj.next_payment_attempt*1000),
        webhooks_delivered_at: new Date(obj.webhooks_delivered_at*1000),
        charge_id: obj.charge,
        application_fee: obj.application_fee,
        subscription_id: obj.subscription,
        tax_percent: obj.tax_percent,
        tax: obj.tax,
        statement_descriptor: obj.statement_descriptor,
        description: obj.description,
        receipt_number: obj.receipt_number
      }, 
      flatten(obj, 'metadata')
    ));

    var discount = obj.discount;
    if (discount) discounts._consume(client, discount);

    for (var i = 0; i < obj.lines.length; i += 1) {
      var line = obj.lines[i];
      yield invoice_lines._consume(client, line);
    }
  });