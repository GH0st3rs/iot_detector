package main

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// searchPatterns - Check if the current text contains string pattern or regex
func searchPatterns(text, search string) bool {
	re, err := regexp.Compile(search)
	if err != nil {
		return strings.Contains(string(text), search)
	}
	return re.MatchString(text)
}

// portCheck - Simple host:port checker
func portCheck(target, port string) bool {
	address := fmt.Sprintf("%s:%s", target, port)
	timeout := 1 * time.Second
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// parsePorts - Function to parse ports like nmap styled
func parsePorts(portArg string) (ports []int) {
	if portArg == "" {
		return ports
	}

	ranges := strings.Split(portArg, ",")

	for _, r := range ranges {
		if strings.Contains(r, "-") {
			bounds := strings.Split(r, "-")
			start, _ := strconv.Atoi(bounds[0])
			end, _ := strconv.Atoi(bounds[1])

			for port := start; port <= end; port++ {
				ports = append(ports, port)
			}
		} else {
			port, _ := strconv.Atoi(r)
			ports = append(ports, port)
		}
	}
	return ports
}
