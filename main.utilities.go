package main

import (
	"bufio"
	"fmt"

	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	//"strconv"
)

func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func escape(u string) string {
	return url.QueryEscape(u)
}

func get(url string, h map[string]string) []byte {

	var body []byte

	if _, err := os.Stat(path + escape(url)); !cache || os.IsNotExist(err) {

		req, _ := http.NewRequest("GET", url, nil)

		for k, v := range h {
			req.Header.Add(k, v)
		}

		res, resErr := http.DefaultClient.Do(req)

		if resErr == nil && res.StatusCode >= 200 && res.StatusCode <= 299 {

			defer res.Body.Close()

			if cache {
				rawresp, err := httputil.DumpResponse(res, true)
				if err == nil {
					ioutil.WriteFile(path+escape(url), rawresp, 0644)
				}
			}
			body, _ = ioutil.ReadAll(res.Body)
		} else {
			panic(fmt.Sprintf("Response was %s from %s", res.Status, req.Host))
		}

	} else {
		f, _ := os.Open(path + escape(url))
		r := bufio.NewReader(f)

		res, _ := http.ReadResponse(r, nil)
		body, err = ioutil.ReadAll(res.Body)

		if err != nil {
			log.Fatal("Can't read file " + path + escape(url))
		}

		//		if rate, ok := res.Header["X-Ratelimit-Remaining"]; ok {
		//			if irate, ierr := strconv.Atoi(rate[0]); ierr == nil {
		//				if irate < 10 {
		//					time.Sleep(1 * time.Second)
		//				}
		//			}
		//		}
	}

	return body
}
