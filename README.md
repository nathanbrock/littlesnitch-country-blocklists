# LittleSnitch Country Block Lists

A collection of [LittleSnitch](https://www.obdev.at/products/littlesnitch/index.html) rule sets to block outbound connections to a given country. 

## Update block lists from MaxMind

1. Install [GoLang](https://go.dev/dl/).
2. Download "GeoLite2 Country" from https://www.maxmind.com/en/accounts/867497/geoip/downloads and unzip.
3. Run the following command when converting GeoLite2 CSV files
`./maxmind_convert_all.sh ~/path/to/GeoLite2-Country-Locations-en.csv ~/path/to/GeoLite2-Country-Blocks-IPv4.csv`

## Update block lists from IP2Location

1. Install [GoLang](https://go.dev/dl/).
2. Download a country list in CIDR format from https://www.ip2location.com/free/visitor-blocker.
3. Run the following command when converting GeoLite2 CSV files. The country name and code is used in the output filename and LS Rules description.
`./ip2location_cidr_list_convert.sh ~/path/to/firewall.txt "{COUNTRY_NAME}" "{COUNTRY_CODE}"`