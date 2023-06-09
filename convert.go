package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type BlockList struct {
	Description         string   `json:"description"`
	Name                string   `json:"name"`
	DeniedRemoteAddress []string `json:"denied-remote-addresses,omitempty"`
}

type Filter struct {
	CountryID string
}

func extractIPs(source string, file io.Reader, filter Filter) ([]string, error) {
	switch source {
	case "maxmind":
		r := csv.NewReader(file)

		return extractIPsFromCSV(r, filter, CSVStructure{
			IPCol:        0,
			CountryIDCol: 2,
		})
	case "ip2location_cidr":
		s := bufio.NewScanner(file)
		return extractIPsFromList(s)
	case "ipinfo":
		r := csv.NewReader(file)

		return extractIPsFromCSV(r, filter, CSVStructure{
			IPRange:      true,
			IPCol:        0,
			IPEndCol:     1,
			CountryIDCol: 2,
		})
	}

	return nil, fmt.Errorf("source not supported (options: maxmind, ip2location")
}

func main() {
	var inputFile string
	flag.StringVar(&inputFile, "input_csv", "", "path to maxmind csv file")
	var source string
	flag.StringVar(&source, "input_source", "maxmind", "source of the ip addresses")
	var dir string
	flag.StringVar(&dir, "output_dir", "./", "output directory for resulting files")
	var filename string
	flag.StringVar(&filename, "output_file", "ip_block_list", "filename of the resulting files")
	var name string
	flag.StringVar(&name, "list_name", "", "name used in blocklist")
	var desc string
	flag.StringVar(&desc, "list_desc", "", "description used in blocklist")
	var filterCountryID string
	flag.StringVar(&filterCountryID, "countryid", "", "filter by maxmind country id (supported by maxmind input source only")

	flag.Parse()

	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	addrs, err := extractIPs(source, file, Filter{CountryID: filterCountryID})
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
func buildBlockList(name, desc string, ips []string) []BlockList {
	sets := make([]BlockList, 0)

	maxSize := 200000
	var j int
	for i := 0; i < len(ips); i += maxSize {
		j += maxSize
		if j > len(ips) {
			j = len(ips)
		}

		batch := ips[i:j]

		sets = append(sets, BlockList{
			Description:         name,
			Name:                desc,
			DeniedRemoteAddress: batch,
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

type CSVStructure struct {
	IPRange      bool
	IPCol        int
	IPStartCol   int
	IPEndCol     int
	CountryIDCol int
}

// extractIPsFromCSV reads in a CSV file and extracts IPs for a given structure.
func extractIPsFromCSV(r *csv.Reader, filter Filter, structure CSVStructure) ([]string, error) {
	var ipNetworks []string

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

		if filter.CountryID != "" && record[structure.CountryIDCol] != filter.CountryID {
			continue
		}

		if structure.IPRange == true {
			start := net.ParseIP(record[structure.IPStartCol])
			end := net.ParseIP(record[structure.IPEndCol])

			if start.To4() == nil || end.To4() == nil {
				continue
			}

			r := fmt.Sprintf("%s-%s", start.String(), end.String())
			ipNetworks = append(ipNetworks, r)
			continue
		}

		ipStr := record[structure.IPCol]
		_, ipv4Net, err := net.ParseCIDR(ipStr)
		if err != nil {
			return nil, err
		}

		ipNetworks = append(ipNetworks, ipv4Net.String())
	}

	return ipNetworks, nil
}

// extractIPsFromList extracts and returns IPs from a single text list.
func extractIPsFromList(s *bufio.Scanner) ([]string, error) {
	var ipNetworks []string

	for s.Scan() {
		t := s.Text()
		if strings.Trim(t, " ")[0] == '#' {
			continue
		}

		_, ipv4Net, err := net.ParseCIDR(t)
		if err != nil {
			return nil, err
		}
		ipNetworks = append(ipNetworks, ipv4Net.String())
	}

	return ipNetworks, nil
}
