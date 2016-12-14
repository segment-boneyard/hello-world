'use strict';

const kit = require('@segment/kit');
const pkg = require('../../package.json');

module.exports = kit.create({
  name: pkg.name,
  version: pkg.version,
  config: {
    secret: {
      description: 'The Stripe secret key used to sync data',
      required: true
    },
    resources: {
      description: 'Override for only running some resources',
      type: (str) => str.split(',')
    },
    'datadog.addr': {
      description: 'Address used for Datadog/statsd metrics'
    },
    'set-transfer-id': {
      description: 'Set transfer_id on balance_transactions collection',
      type: 'Boolean'
    },
    rps: {
      description: 'reqs/second granted to Stripe API'
    }
  }
});
