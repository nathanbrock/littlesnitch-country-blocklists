#!/usr/bin/env bash

set -e

CIDR_FILE=$1
COUNTRY_NAME=$2
COUNTRY_CODE=$3

go run convert.go \
  -input_csv "$CIDR_FILE" \
  -input_source "ip2location_cidr" \
  -list_name "Block $COUNTRY_NAME IPs" \
  -list_desc "Blocks outbound requests to $COUNTRY_NAME IP ranges" \
  -output_dir blocklists_by_country/ip2location \
  -output_file "ip_block_list_$COUNTRY_CODE"