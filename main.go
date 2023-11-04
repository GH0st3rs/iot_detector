// iot_detector project main.go
package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type ListItem struct {
	ip   string
	port string
}

type OutputStruct struct {
	thread_id int
	ip        string
	port      string
	answer    string
}

var (
	THREADS_COUNT  int
	VERBOSE        bool
	GLOBAL_REQUEST Request
	AUTOSCHEME     bool
	scheme_array   = []string{"http://", "https://"}
	PORTS          string
)

func request(target string, port string, client *http.Client) (string, bool) {
	var t = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	if strings.Contains("https", target) {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		t.MaxIdleConns = 10
		t.IdleConnTimeout = 30 * time.Second
	}
	client.Transport = t
	// build a new request, but not doing the POST yet
	url := fmt.Sprintf("%s:%s%s", target, port, GLOBAL_REQUEST.Path)
	// Check POST and GET requests
	var req *http.Request
	var err error
	if GLOBAL_REQUEST.Method == "POST" {
		req, err = http.NewRequest(GLOBAL_REQUEST.Method, url, bytes.NewBuffer([]byte(GLOBAL_REQUEST.Data)))
	} else if GLOBAL_REQUEST.Method == "GET" {
		if len(GLOBAL_REQUEST.Data) > 0 {
			url += "?" + GLOBAL_REQUEST.Data
		}
		req, err = http.NewRequest(GLOBAL_REQUEST.Method, url, nil)
	} else {
		return "[WRONG METHOD]", false
	}
	if err != nil {
		return "[ERROR CONNECT]", false
	}
	// Insert header items
	for header := range GLOBAL_REQUEST.Headers {
		req.Header.Add(header, GLOBAL_REQUEST.Headers[header])
	}
	// now POST it
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("[NO RESPONSE] => %s", err.Error()), false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "[WRONG RESPONSE]", false
	}
	answer, err := ioutil.ReadAll(resp.Body)
	_, _ = io.Copy(io.Discard, resp.Body) // Discard the body
	if searchPatterns(string(answer), GLOBAL_REQUEST.Search) {
		return "[SUCCESS]", true
	}
	return "[NOT DETECTED]", false
}

func worker(thread_num int, jobs chan ListItem, wg *sync.WaitGroup, output chan OutputStruct, c *http.Client) {
	var answer, host string
	var status bool
	var result OutputStruct
	for v := range jobs {
		status = false
		answer = ""
		if portCheck(v.ip, v.port) {
			if AUTOSCHEME {
				for _, scheme := range scheme_array {
					host = scheme + v.ip
					answer, status = request(host, v.port, c)
					if status {
						break
					}
				}
			} else {
				answer, status = request(v.ip, v.port, c)
			}
		}
		if status || VERBOSE {
			result.answer = answer
			result.ip = v.ip
			result.port = v.port
			result.thread_id = thread_num
			output <- result
		}
	}
	wg.Done()
}

func main() {
	var iplist string
	var request_file string
	flag.IntVar(&THREADS_COUNT, "t", 1000, "Thread count")
	flag.StringVar(&iplist, "l", "", "List of ip or ip,port")
	flag.StringVar(&request_file, "r", "", "Json request file")
	flag.BoolVar(&VERBOSE, "v", true, "Verbose")
	flag.BoolVar(&AUTOSCHEME, "a", false, "Auto URL scheme")
	flag.StringVar(&PORTS, "p", "", "Ports to scan (e.g. 22,80,443,1000-2000)")
	flag.Parse()
	// Use maximum numbers of CPU
	numcpu := runtime.NumCPU()
	fmt.Println("NumCPU", numcpu)
	runtime.GOMAXPROCS(1)

	// Parse input request file
	if err := GLOBAL_REQUEST.loadRequestFromFile(request_file); err != nil {
		fmt.Println("Error:", err)
		return
	}

	jobs := make(chan ListItem, 100)
	wg := sync.WaitGroup{}
	output := make(chan OutputStruct)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 1; i < THREADS_COUNT; i++ {
		go worker(i, jobs, &wg, output, client)
		wg.Add(1)
	}

	file, err := os.Open(iplist)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// print output
	go func() {
		for line := range output {
			fmt.Printf("{%d}\t%s\t%s\t%s\n", line.thread_id, line.ip, line.port, line.answer)
		}
	}()

	scanner := bufio.NewScanner(file)
	// Parser for input ports
	ports := parsePorts(PORTS)
	// Input list scanner
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, ",") {
			// Split line from file by host:port and create a task
			host_port := strings.Split(line, ",")
			item := ListItem{ip: host_port[0], port: host_port[1]}
			jobs <- item
		} else {
			// Generate tasks for single host and multiple ports
			for _, port := range ports {
				item := ListItem{ip: line, port: fmt.Sprint(port)}
				jobs <- item
			}
		}
	}

	close(jobs)
	wg.Wait()
	close(output)
}
