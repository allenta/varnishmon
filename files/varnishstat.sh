#!/usr/bin/env bash

#
# Notes:
#
#   - All counters, gauges, etc. coming from 'varnishstat' are internally
#     modelled in Varnish as uint64.
#
#   - Possible 'flag' (i.e., field semantics) values (see 'include/vapi/vsc.h'):
#     - c: counter.
#     - g: gauge.
#     - b: uint64 bitmap.
#     - q: boolean (Varnish Enterprise only).
#     - ?: unknown.
#
#   - Possible 'format' (i.e., field display format) values  (see
#     'include/vapi/vsc.h'):
#     - i: integer.
#     - d: duration.
#     - B: bytes.
#     - b: uint64 bitmap.
#     - q: boolean (Varnish Enterprise only).
#     - ?: unknown.
#
#   - Beware of JSON numbers being limited to float64 & possible precision loss.
#

COUNTERS='
  "MGT.uptime": {
    "description": "Management process uptime",
    "flag": "c",
    "format": "d",
    "value": 6060502
  },
  "MAIN.n_backend": {
    "description": "Number of backends",
    "flag": "g",
    "format": "i",
    "value": 6
  },
  "VBE.boot.default_be.happy": {
    "description": "Happy health probes",
    "flag": "b",
    "format": "b",
    "value": 18446744073709551615
  },
  "VBE.boot.default_be.is_healthy": {
    "description": "Backend health status",
    "flag": "q",
    "format": "q",
    "value": 1
  },
  "MAIN.s_req_hdrbytes": {
    "description": "Request header bytes",
    "flag": "c",
    "format": "B",
    "value": 854488484
  }'

if [ $((RANDOM % 2)) -eq 0 ]; then
  # Old 'varnishstat' output.
  jq <<EOF
  {
    "timestamp": "2024-01-01T13:00:00",
    $COUNTERS
  }
EOF
else
  # New 'varnishstat' output, for Varnish Cache >= 6.5.0.
  jq <<EOF
  {
    "version": 1,
    "timestamp": "2024-01-01T13:00:00",
    "counters": {
      $COUNTERS
    }
  }
EOF
fi
