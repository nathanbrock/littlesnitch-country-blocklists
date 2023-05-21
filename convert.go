package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type BlockList struct {
	Description         string   `json:"description"`
	Name                string   `json:"name"`
	DeniedRemoteAddress []string `json:"denied-remote-addresses,omitempty"`
}

type Filter struct {
	CountryID string
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "file", "", "path to maxmind csv file")
	var filterCountryID string
	flag.StringVar(&filterCountryID, "countryid", "", "filter by maxmind country id")

	flag.Parse()

	if filePath == "" {
		log.Fatal("-file={filepath} flag required")
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	r := csv.NewReader(file)

	addrs, err := extractIPsFromCSV(r, Filter{
		CountryID: filterCountryID,
	})
	if err != nil {
		log.Fatal(err)
	}

	lists := buildBlockList(addrs)

	for i, list := range lists {
		err := writeJSONToFile(list, fmt.Sprintf("ip_block_list_%s_%d.lsrules", filterCountryID, i))
		if err != nil {
			log.Fatal(err)
		}
	}
}

// buildBlockList creates one or more LittleSnitch block list structs for the given slice of IP addresses
func buildBlockList(ips []net.IPNet) []BlockList {
	sets := make([]BlockList, 0)

	maxSize := 200000
	var j int
	for i := 0; i < len(ips); i += maxSize {
		j += maxSize
		if j > len(ips) {
			j = len(ips)
		}

		batch := ips[i:j]

		addrs := make([]string, len(batch))
		for i, r := range batch {
			addrs[i] = r.String()
		}

		sets = append(sets, BlockList{
			Description:         "Block IPs",
			Name:                "Block IPs",
			DeniedRemoteAddress: addrs,
		})
	}

	return sets
}

// writeJSONToFile marshals given struct and dumps directly to given path
func writeJSONToFile(j interface{}, path string) error {
	contents, err := json.Marshal(j)
	if err != nil {
		return err
	}

	return os.WriteFile(path, contents, 0644)
}

// extractIPsFromCSV expects a GeoLite2 CSV file which it'll extract IP addresses from
func extractIPsFromCSV(r *csv.Reader, filter Filter) ([]net.IPNet, error) {
	var ipNetworks []net.IPNet

	firstRow := true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if firstRow {
			firstRow = false
			continue
		}

		if filter.CountryID != "" {
			if record[2] != filter.CountryID {
				continue
			}
		}

		ipStr := record[0]
		_, ipv4Net, err := net.ParseCIDR(ipStr)
		if err != nil {
			return nil, err
		}

		ipNetworks = append(ipNetworks, *ipv4Net)
	}

	return ipNetworks, nil
}
