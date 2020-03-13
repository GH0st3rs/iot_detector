// iot_detector project main.go
package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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

type Request struct {
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Search  string            `json:"search"`
	Data    string            `json:"data"`
}

var THREADS_COUNT int
var VERBOSE bool
var GLOBAL_REQUEST Request

func load_request_from_file(filename string) error {
	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	// Parse Json file
	err = json.Unmarshal(byteValue, &GLOBAL_REQUEST)
	if err != nil {
		return err
	}
	return nil
}

func request(target string, port string, shema string) (string, bool) {
	if shema == "https" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	// build a new request, but not doing the POST yet
	url := fmt.Sprintf("%s://%s:%s%s", shema, target, port, GLOBAL_REQUEST.Path)
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
	if err != nil || resp.StatusCode != 200 {
		return "[NO RESPONSE]", false
	}
	defer resp.Body.Close()
	answer, err := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(answer), GLOBAL_REQUEST.Search) {
		return "SUCCESS", true
	}
	return "[NOT DETECTED]", false
}

func worker(thread_num int, jobs chan ListItem, wg *sync.WaitGroup) {
	for v := range jobs {
		status := false
		answer := ""
		for _, shema := range []string{"https", "http"} {
			answer, status = request(v.ip, v.port, shema)
			if status == true {
				break
			}
		}
		if status == true || VERBOSE == true {
			fmt.Printf("{%d}\t%s\t%s\t%s\n", thread_num, v.ip, v.port, answer)
		}
	}
	wg.Done()
}

func main() {
	var iplist string
	var request_file string
	flag.IntVar(&THREADS_COUNT, "t", 1000, "Thread count")
	flag.StringVar(&iplist, "l", "", "List of ip,port")
	flag.StringVar(&request_file, "r", "", "Json request file")
	flag.BoolVar(&VERBOSE, "v", false, "Verbose")
	flag.Parse()

	numcpu := runtime.NumCPU()
	fmt.Println("NumCPU", numcpu)
	runtime.GOMAXPROCS(numcpu)

	err := load_request_from_file(request_file)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	jobs := make(chan ListItem, 100)
	wg := sync.WaitGroup{}
	for i := 1; i < THREADS_COUNT; i++ {
		go worker(i, jobs, &wg)
		wg.Add(1)
	}

	file, err := os.Open(iplist)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitted_data := strings.Split(line, ",")
		item := ListItem{ip: splitted_data[0], port: splitted_data[1]}
		jobs <- item
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	close(jobs)
	wg.Wait()
}
