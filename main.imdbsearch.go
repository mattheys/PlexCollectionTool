package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// IMDbSearch iterates through items in an IMDb search list
func IMDbSearch(inputURL string, limit int) <-chan string {
	// Request the HTML page.
	count := 250

	if limit == 0 {
		limit = 5000
	} else if limit < count{
		count = limit
	}

	chnl := make(chan string)
	go func() {
		start := 1

		for start < limit {

			u, err := url.Parse(inputURL)
			if err != nil {
				log.Fatal(err)
			}

			q := u.Query()
			q.Set("view", "simple")
			q.Set("start", strconv.Itoa(start))
			q.Set("count", strconv.Itoa(count))
			u.RawQuery = q.Encode()

			res, err := http.Get(u.String())
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
			}

			// Load the HTML document
			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				log.Fatal(err)
			}

			// Find the review items
			selection := doc.Find(".ribbonize").Each(func(i int, s *goquery.Selection) {
				title, exists := s.Attr("data-tconst")
				if exists && start+i < limit+1 {
					chnl <- title
				}
			})

			if len(selection.Nodes) == 0 {
				break
			}

			start += count
		}
		close(chnl)
	}()
	return chnl
}

// XRange is an iterator over all the numbers from 0 to the limit.
func XRange(limit int) <-chan int {
	chnl := make(chan int)
	go func() {
		for i := 0; i < limit; i++ {
			chnl <- i
		}

		// Ensure that at the end of the loop we close the channel!
		close(chnl)
	}()
	return chnl
}
