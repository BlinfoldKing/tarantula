package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"os"
	"strings"
)

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	return ok, href
}

func crawlUrl(url string, chUrl chan string, chFinished chan bool) {
	resp, err := http.Get(url)
	defer func() {
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl " + url)
		return
	}

	b := resp.Body
	defer b.Close()
	document := html.NewTokenizer(b)
	for {
		tt := document.Next()
		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := document.Token()

			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			ok, url := getHref(t)
			if !ok {
				continue
			}

			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				fmt.Println(url)
				chUrl <- url
			}
		}
	}
}

func main() {
	urls := os.Args[1:]

	chFinished := make(chan bool)
	chUrl := make(chan string)
	for _, url := range urls {
		hasProto := strings.Index(url, "http") == 0
		if !hasProto {
			url = "http://" + url
		}

		go crawlUrl(url, chUrl, chFinished)
	}

	for c := 0; c < len(urls); {
		select {
		case <-chUrl:
			urls = append(urls, <-chUrl)
			go crawlUrl(<-chUrl, chUrl, chFinished)
		case <-chFinished:
			c++
		}
	}
}
