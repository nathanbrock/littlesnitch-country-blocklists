# LittleSnitch Country Block Lists

A collection of [LittleSnitch](https://www.obdev.at/products/littlesnitch/index.html) rule sets to block outbound connections to a given country. 

## Update block lists
1. Install [GoLang](https://go.dev/dl/).
2. Download "GeoLite2 Country" from https://www.maxmind.com/en/accounts/867497/geoip/downloads and unzip.
3. Run the following command
`./convert_all.sh ~/path/to/GeoLite2-Country-Locations-en.csv ~/path/to/GeoLite2-Country-Blocks-IPv4.csv`