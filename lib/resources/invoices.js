var Resource = require('../resource');
var flatten = require('../utils/flatten');
var extend = require('extend');
var discounts = require('./discounts');
var invoice_lines = require('./invoice_lines');
var dates = require('../utils/dates');

module.exports = Resource()
  .url('https://api.stripe.com/v1/invoices')
  .subpages(['lines'])
  .collection('invoices')
  .consume(function* (client, obj) {
    var properties = extend({
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
        charge_id: obj.charge,
        application_fee: obj.application_fee,
        subscription_id: obj.subscription,
        tax_percent: obj.tax_percent,
        tax: obj.tax,
        statement_descriptor: obj.statement_descriptor,
        description: obj.description,
        receipt_number: obj.receipt_number
      }, 
      flatten(obj, 'metadata'),
      dates(obj, { 
        date: 'date',
        period_start: 'period_start',
        period_end: 'period_end',
        next_payment_attempt: 'next_payment_attempt',
        webhooks_delivered_at: 'webhooks_delivered_at'
      })
    );

    var discount = obj.discount;
    if (discount) {
      discounts._consume(client, discount);
      properties.discount_id = discount.customer + '_' + discount.coupon.id;
    }

    yield client.set('invoices', obj.id, properties);

    for (var i = 0; i < obj.lines.length; i += 1) {
      var line = obj.lines[i];
      yield invoice_lines._consume(client, line);
    }
  });