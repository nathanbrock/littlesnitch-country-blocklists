package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/netip"
	"os"
	"strconv"
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

func extractIPs(source string, file io.Reader, filter Filter) ([]net.IPNet, error) {
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

type CSVStructure struct {
	IPRange      bool
	IPCol        int
	IPStartCol   int
	IPEndCol     int
	CountryIDCol int
}

// Thank you https://go.dev/play/p/Ynx1liLAGs2
func IpRangeToCIDR(cidr []string, start, end string) ([]string, error) {
	ips, err := netip.ParseAddr(start)
	if err != nil {
		return nil, err
	}
	ipe, err := netip.ParseAddr(end)
	if err != nil {
		return nil, err
	}

	isV4 := ips.Is4()
	if isV4 != ipe.Is4() {
		return nil, errors.New("start and end types are different")
	}
	if ips.Compare(ipe) > 0 {
		return nil, errors.New("start > end")
	}

	var (
		ipsInt = new(big.Int).SetBytes(ips.AsSlice())
		ipeInt = new(big.Int).SetBytes(ipe.AsSlice())
		tmpInt = new(big.Int)
		mask   = new(big.Int)
		one    = big.NewInt(1)
		buf    []byte

		bits, maxBit uint
	)
	if isV4 {
		maxBit = 32
		buf = make([]byte, 4)
	} else {
		maxBit = 128
		buf = make([]byte, 16)
	}

	for {
		bits = 1
		mask.SetUint64(1)
		for bits < maxBit {
			if (tmpInt.Or(ipsInt, mask).Cmp(ipeInt) > 0) ||
				(tmpInt.Lsh(tmpInt.Rsh(ipsInt, bits), bits).Cmp(ipsInt) != 0) {
				bits--
				mask.Rsh(mask, 1)
				break
			}
			bits++
			mask.Add(mask.Lsh(mask, 1), one)
		}

		addr, _ := netip.AddrFromSlice(ipsInt.FillBytes(buf))
		cidr = append(cidr, addr.String()+"/"+strconv.FormatUint(uint64(maxBit-bits), 10))

		if tmpInt.Or(ipsInt, mask); tmpInt.Cmp(ipeInt) >= 0 {
			break
		}
		ipsInt.Add(tmpInt, one)
	}
	return cidr, nil
}

// extractIPsFromCSV reads in a CSV file and extracts IPs for a given structure.
func extractIPsFromCSV(r *csv.Reader, filter Filter, structure CSVStructure) ([]net.IPNet, error) {
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

		if filter.CountryID != "" && record[structure.CountryIDCol] != filter.CountryID {
			continue
		}

		if structure.IPRange == true {

			start := record[structure.IPStartCol]
			end := record[structure.IPEndCol]

			if net.ParseIP(start).To4() == nil {
				continue
			}

			cidr, err := IpRangeToCIDR(nil, start, end)
			if err != nil {
				panic(err)
			}

			_, ipv4Net, err := net.ParseCIDR(cidr[0])
			if err != nil {
				return nil, err
			}

			ipNetworks = append(ipNetworks, *ipv4Net)
			continue
		}

		ipStr := record[structure.IPCol]
		_, ipv4Net, err := net.ParseCIDR(ipStr)
		if err != nil {
			return nil, err
		}

		ipNetworks = append(ipNetworks, *ipv4Net)
	}

	return ipNetworks, nil
}

// extractIPsFromList extracts and returns IPs from a single text list.
func extractIPsFromList(s *bufio.Scanner) ([]net.IPNet, error) {
	var ipNetworks []net.IPNet

	for s.Scan() {
		t := s.Text()
		if strings.Trim(t, " ")[0] == '#' {
			continue
		}

		_, ipv4Net, err := net.ParseCIDR(t)
		if err != nil {
			return nil, err
		}
		ipNetworks = append(ipNetworks, *ipv4Net)
	}

	return ipNetworks, nil
}
