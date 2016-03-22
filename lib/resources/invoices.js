'use strict';

const co = require('bluebird').coroutine;
var Resource = require('../resource');
var dates = require('../utils/dates');
var discounts = require('./discounts');
var extend = require('extend');
var flatten = require('../utils/flatten');
var invoice_lines = require('./invoice_lines');

module.exports = Resource()
  .url('https://api.stripe.com/v1/invoices')
  .subpages(['lines'])
  .collection('invoices')
  .consume(co(function* (client, obj, log) {
    log.debug({ invoice_id: obj.id }, 'Received invoice from Stripe API');

    var properties = extend(
      {
        amount_due: obj.amount_due,
        application_fee: obj.application_fee,
        attempt_count: obj.attempt_count,
        attempted: obj.attempted,
        charge_id: obj.charge,
        closed: obj.closed,
        currency: obj.currency,
        customer_id: obj.customer,
        description: obj.description,
        ending_balance: obj.ending_balance,
        forgiven: obj.forgiven,
        paid: obj.paid,
        receipt_number: obj.receipt_number,
        starting_balance: obj.starting_balance,
        statement_descriptor: obj.statement_descriptor,
        subscription_id: obj.subscription,
        subtotal: obj.subtotal,
        tax: obj.tax,
        tax_percent: obj.tax_percent,
        total: obj.total
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
      // Add discount foreign key to invoice object
      properties.discount_id = `${ discount.customer }_${ discount.coupon.id }`;

      // Consume as a discount object
      discounts._consume(client, discount);
    }

    if (obj.lines.length) {
      // TODO: Add foreign key to `properties` or to `line` object
      // Consume lines as invoice_line objects
      for (var line of obj.lines) {
        yield invoice_lines._consume(client, line);
      }
    }

    yield client.set('invoices', obj.id, properties);
  }));
