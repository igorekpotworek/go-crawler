package main

import (
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"strings"
)

var d = "https://cracovia.pl"
var retryClient = client()

type page struct {
	domain string
	links  []string
}

func main() {
	visited := make(map[string]bool)
	crawledPages := make(chan page)
	counter := 1
	go visit(d, crawledPages)
	for counter > 0 {
		println("counter=%s", counter)
		page := <-crawledPages
		visited[page.domain] = true
		println("visited=%s", len(visited))

		for _, url := range page.links {
			if !visited[url] {
				counter++
				go visit(url, crawledPages)
			}
		}
		counter--
	}
}

func visit(domain string, c chan page) {
	println(domain)
	links := crawl(domain)
	c <- page{domain: domain, links: links}
}

func crawl(domain string) []string {
	var result []string
	links := links(domain)
	for _, url := range links {
		isSameDomain := strings.Index(url, "/") == 0 && url != "/"
		if isSameDomain {
			result = append(result, d+url)
		}
	}
	return result
}

func links(address string) []string {
	resp, err := retryClient.Get(address)
	if err != nil {
		log.Fatal(err)
	}
	return parseLinks(resp.Body)
}

func client() *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 20
	retryClient.HTTPClient.Transport = &http.Transport{
		MaxConnsPerHost: 10,
	}
	return retryClient
}

func parseLinks(body io.Reader) []string {
	var links []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}

				}
			}

		}
	}
}
