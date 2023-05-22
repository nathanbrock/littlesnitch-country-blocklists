#!/usr/bin/env bash

set -e

LOCATION_FILE=$1
COUNTRY_FILE=$2

{
  read # skip header row
  while IFS=, read -r geoname _ _ _ countrycode name _; do
    echo "Building block list for $name ($geoname - $countrycode)"
      go run convert.go \
        -input_csv "$COUNTRY_FILE" \
        -input_source "maxmind" \
        -countryid "$geoname" \
        -list_name "Block $name IPs" \
        -list_desc "Blocks outbound requests to $name IP ranges" \
        -output_dir blocklists_by_country/maxmind \
        -output_file "ip_block_list_$countrycode"
  done
} < "$LOCATION_FILE"
