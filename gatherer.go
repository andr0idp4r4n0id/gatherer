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
	return resp.StatusCode == 200
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
	var delay int
	conc := flag.Int("concurrency", 10, "concurrency level")
	url_t := flag.String("url", "", "url to target.")
	flag.StringVar(&filename, "wordlist", "", "wordlist to use.")
	flag.IntVar(&delay, "delay", 0, "delay between each request.")
	flag.Parse()
	file := ReadFile(filename)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	for i := 0; i < *conc; i++ {
		for scanner.Scan() {
			word := scanner.Text()
			path := *url_t + word
			wg.Add(1)
			go func() {
				resp := HeadRequest(path)
				time.Sleep(time.Duration(delay) * time.Second)
				fmt.Printf("Attempting: %s                   \r\r\r\r", path)
				if resp != nil {
					if CheckHTTPStatusCode200(resp) {
						fmt.Println(path)
					} else if CheckHTTPStatusCode400(resp) {
						for scanner.Scan() {
							second_word := scanner.Text()
							second_path := *url_t + word + "/" + second_word
							second_resp := HeadRequest(second_path)
							wg.Add(1)
							go func() {
								if second_resp.StatusCode == 200 {
									fmt.Println(second_path)
								}
								wg.Done()
							}()
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
