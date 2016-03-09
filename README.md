# @segment/source-stripe

Stripe data source for Segment.

## Requirements

- Node >= 4.0.0

## Quickstart

```
node bin/stripe \
  --secret    [ STRIPE_SECRET ] \
  --write-key [ WRITE_KEY ]
```

## Metrics

The following metrics are collected by this source:

Name                | Type      | Description
----                | ----      | -----------
`http.response.payload_size` | Histogram | HTTP response content length from source API response.
`http.request.duration` | Histogram | How long an HTTP request took, in ms.
`http.request.total` | Counter | Total number of HTTP requests to source API.
`http.response.status_code` | Counter | HTTP response status code from source API response.
