#!/usr/bin/env bash

set -e

COUNTRY_FILE=$1
COUNTRY_NAME=$2
COUNTRY_CODE=$3

go run convert.go \
  -input_csv "$COUNTRY_FILE" \
  -input_source "ipinfo" \
  -list_name "Block $COUNTRY_NAME IPs (IPInfo)" \
  -list_desc "Blocks requests to $COUNTRY_NAME IPs (Source: IPInfo)" \
  -countryid "$COUNTRY_CODE" \
  -output_dir blocklists_by_country/ipinfo \
  -output_file "ip_block_list_$COUNTRY_CODE"