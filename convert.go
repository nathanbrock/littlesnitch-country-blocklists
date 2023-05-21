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
	var inputFile string
	flag.StringVar(&inputFile, "input_csv", "", "path to maxmind csv file")
	var dir string
	flag.StringVar(&dir, "output_dir", "./", "output directory for resulting files")
	var filename string
	flag.StringVar(&filename, "output_file", "ip_block_list", "filename of the resulting files")
	var name string
	flag.StringVar(&name, "list_name", "", "name used in blocklist")
	var desc string
	flag.StringVar(&desc, "list_desc", "", "description used in blocklist")
	var filterCountryID string
	flag.StringVar(&filterCountryID, "countryid", "", "filter by maxmind country id")

	flag.Parse()

	file, err := os.Open(inputFile)
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

	lists := buildBlockList(name, desc, addrs)
	err = writeListsToFile(dir, filename, lists)
	if err != nil {
		log.Fatal(err)
	}
}

// writeListsToFile takes a slice of block lists and writes them to the given directory
func writeListsToFile(dir, filename string, lists []BlockList) error {
	for i, list := range lists {
		suf := ""
		if len(lists) > 1 {
			suf = fmt.Sprintf("_%d", i+1)
		}

		err := writeJSONToFile(list, fmt.Sprintf("%s/%s%s.lsrules", dir, filename, suf))
		if err != nil {
			return err
		}
	}

	return nil
}

// buildBlockList creates one or more LittleSnitch block list structs for the given slice of IP addresses
func buildBlockList(name, desc string, ips []net.IPNet) []BlockList {
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
			Description:         name,
			Name:                desc,
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
