// var Resource = require('../resource');
// var flatten = require('../utils/flatten');
// var extend = require('extend');

// module.exports = Resource()
//   .object('disputes')
//   .collection('disputes')
//   .properties(function (obj) {
//     return extend(
//       {
//         "created": new Date(obj.created*1000),
//         "charge_id": obj.charge,
//         "amount": obj.amount,
//         "status": obj.status,
//         "currency": obj.currency,
//         "reason": obj.reason,
//         "is_charge_refundable": obj.is_charge_refundable
//       }, 
//       flatten(obj, 'metadata')
//     );
//   });


//   "balance_transactions": [
//     {
//       "id": "txn_15RsQX2eZvKYlo2CUTLzmHcJ",
//       "object": "balance_transaction",
//       "amount": -195,
//       "currency": "usd",
//       "net": -1695,
//       "type": "adjustment",
//       "created": 1422915137,
//       "available_on": 1423440000,
//       "status": "available",
//       "fee": 1500,
//       "fee_details": [
//         {
//           "amount": 1500,
//           "currency": "usd",
//           "type": "stripe_fee",
//           "description": "Dispute fee",
//           "application": null
//         }
//       ],
//       "source": "dp_15RsQX2eZvKYlo2C0MFNUWJC",
//       "description": "Chargeback withdrawal for ch_15RsQR2eZvKYlo2CA8IfzCX0",
//       "sourced_transfers": {
//         "object": "list",
//         "total_count": 0,
//         "has_more": false,
//         "url": "/v1/transfers?source_transaction=ad_15RsQX2eZvKYlo2CYlUxjQ32",
//         "data": [

//         ]
//       }
//     }
//   ],
//   "evidence_details": {
//     "due_by": 1424303999,
//     "past_due": false,
//     "has_evidence": false,
//     "submission_count": 0
//   },
//   "evidence": {
//     "product_description": null,
//     "customer_name": null,
//     "customer_email_address": null,
//     "customer_purchase_ip": null,
//     "customer_signature": null,
//     "billing_address": null,
//     "receipt": null,
//     "shipping_address": null,
//     "shipping_date": null,
//     "shipping_carrier": null,
//     "shipping_tracking_number": null,
//     "shipping_documentation": null,
//     "access_activity_log": null,
//     "service_date": null,
//     "service_documentation": null,
//     "duplicate_charge_id": null,
//     "duplicate_charge_explanation": null,
//     "duplicate_charge_documentation": null,
//     "refund_policy": null,
//     "refund_policy_disclosure": null,
//     "refund_refusal_explanation": null,
//     "cancellation_policy": null,
//     "cancellation_policy_disclosure": null,
//     "cancellation_rebuttal": null,
//     "customer_communication": null,
//     "uncategorized_text": null,
//     "uncategorized_file": null
//   },
//   "metadata": {
//   }
// }