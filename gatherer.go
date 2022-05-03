package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

func ReadFile(filename string) *os.File {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return file
}

func CheckHTTPStatusCode400(resp *http.Response) bool {
	return resp.StatusCode >= 401 && resp.StatusCode <= 403
}

func CheckHTTPStatusCode200(resp *http.Response) bool {
	return resp.StatusCode <= 399
}

func HeadRequest(url_t string) *http.Response {
	resp, err := http.Head(url_t)
	if err != nil {
		return nil
	}
	return resp
}

func main() {
	var wg sync.WaitGroup
	var filename string
	var delay float64
	var max_depth int
	conc := flag.Int("concurrency", 10, "concurrency level")
	url_t := flag.String("url", "", "url to target.")
	flag.StringVar(&filename, "wordlist", "", "wordlist to use.")
	flag.Float64Var(&delay, "delay", 1, "delay between each request.")
	flag.IntVar(&max_depth, "depth", 1, "max depth to request new pages when a forbidden is discovered.")
	flag.Parse()
	file := ReadFile(filename)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	depth := 1
	for i := 0; i < *conc; i++ {
		for scanner.Scan() {
			word := scanner.Text()
			path := *url_t + word
			wg.Add(1)
			go func() {
				resp := HeadRequest(path)
				time.Sleep(time.Duration(delay) * time.Second)
				if resp == nil {
					return
				}
				if CheckHTTPStatusCode200(resp) {
					fmt.Println(path)
				} else if CheckHTTPStatusCode400(resp) {
					for depth <= max_depth {
						for scanner.Scan() {
							word = scanner.Text()
							path += "/" + word
							go func() {
								resp = HeadRequest(path)
								time.Sleep(time.Duration(delay) * time.Second)
								if resp.StatusCode <= 399 {
									fmt.Println(path)
									depth += 1
								}
							}()
						}
						if depth == 1 {
							break
						}
					}

				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
	defer file.Close()
}
