# LittleSnitch Country Block Lists

A collection of [LittleSnitch](https://www.obdev.at/products/littlesnitch/index.html) rule sets to block outbound connections to a given country. 

These lists are not updated regularly and are just for experimentation. If you wish to maintain your own lists, please fork the repo.

## Subscribing to lists

You can subscribe to a block list via the LittleSnitch rules UI by following the steps below.

1. Open the Little Snitch Rules either by opening the application or selecting 'Little Snitch Rules...' from the menu bar icon.
2. Click the plus symbol found at the bottom left side of the window and select "New Rule Group Subscription...".
3. From the `./blocklists_by_country` directory, choose the lists you want to subscribe to and view as "raw".
4. Use the `raw.githubusercontent.com` url as the subscription URL in the Little Snitch UI.
5. Follow the instructions presented. Note that IP lists can be sizable and can cause LS to slow down for a bit as it loads the blocklist.
6. After subscribing, make sure to approve all the new rules via the "Unapproved" tab.

## Updating lists

### Update block lists from MaxMind

1. Install [GoLang](https://go.dev/dl/).
2. Download "GeoLite2 Country" from https://www.maxmind.com/en/accounts/867497/geoip/downloads and unzip.
3. Run the following command when converting GeoLite2 CSV files
`./maxmind_convert_all.sh ~/path/to/GeoLite2-Country-Locations-en.csv ~/path/to/GeoLite2-Country-Blocks-IPv4.csv`

## Update block lists from IP2Location

1. Install [GoLang](https://go.dev/dl/).
2. Download a country list in CIDR format from https://www.ip2location.com/free/visitor-blocker.
3. Run the following command when converting IP2Location single country CIDR list. The country name and code is used in the output filename and LS Rules description.
`./ip2location_cidr_list_convert.sh ~/path/to/firewall.txt "{COUNTRY_NAME}" "{COUNTRY_CODE}"`

## Update block lists from IPInfo

1. Install [GoLang](https://go.dev/dl/).
2. Sign up for an IPInfo account, download the Free Country CSV GZ file from https://ipinfo.io/account/data-downloads and extract.
3. Run the following command to convert the IPInfo DB into a LittleSnitch blocklist for a single country. The country name and code is used in the output filename and LS Rules description. The country code must match the country code used in the source file.
   `./ipinfo_convert_single.sh ~/path/to/country.csv "{COUNTRY_NAME}" "{COUNTRY_CODE}"`