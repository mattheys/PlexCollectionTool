package main

import (
	"fmt"

	"io/ioutil"
	"net/http"
	"net/url"
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

	req, _ := http.NewRequest("GET", url, nil)

	for k, v := range h {
		req.Header.Add(k, v)
	}

	res, resErr := http.DefaultClient.Do(req)

	if resErr == nil && res.StatusCode >= 200 && res.StatusCode <= 299 {

		defer res.Body.Close()

		body, _ = ioutil.ReadAll(res.Body)
	} else {
		panic(fmt.Sprintf("Response was %s from %s", res.Status, req.Host))
	}

	return body
}
