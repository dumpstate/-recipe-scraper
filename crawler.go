package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
)

var TODO_TIMEOUT_SECONDS = 1
var TODO_IDLE_LOOP = 10

type Crawler struct {
	Scraper     Scraper
	Store       *Store
	Concurrency int
}

func NewCrawler(scraper Scraper, store *Store, concurrency int) *Crawler {
	return &Crawler{
		Scraper:     scraper,
		Store:       store,
		Concurrency: concurrency,
	}
}

type NextLinks struct {
	SourceUrl string
	Links     []string
}

func allLinks(resp string, currUrl string) []string {
	urls := []string{}
	anchors := soup.HTMLParse(resp).FindAll("a")
	for _, anchor := range anchors {
		href := anchor.Attrs()["href"]

		if strings.HasPrefix(href, "/") {
			parsedUrl, err := url.Parse(currUrl)
			if err == nil {
				urls = append(urls, fmt.Sprintf("%s://%s%s", parsedUrl.Scheme, parsedUrl.Host, href))
			}
		}
	}
	return urls
}

func worker(
	scraper Scraper,
	urls <-chan string,
	recipes chan<- *Recipe,
	nextUrls chan<- *NextLinks,
) {
	for url := range urls {
		recipe, resp, err := scraper.TryFind(url)

		if err == nil {
			recipes <- recipe
		} else {
			nextUrls <- &NextLinks{
				SourceUrl: url,
				Links:     allLinks(resp, url),
			}
		}
	}
}

func recipeReceiver(
	scraper Scraper,
	store *Store,
	recipes <-chan *Recipe,
) {
	for recipe := range recipes {
		store.SaveRecipe(scraper.Name(), recipe)
	}
}

func linksReceiver(scraper Scraper, store *Store, nextUrls <-chan *NextLinks) {
	for next := range nextUrls {
		store.AddCrawlerJobs(scraper.Name(), next.SourceUrl, next.Links)
	}
}

func (crawler *Crawler) Crawl(startUrl string) {
	urls := make(chan string, crawler.Concurrency)
	recipes := make(chan *Recipe)
	nextUrls := make(chan *NextLinks)
	emptyTodo := 0

	for w := 1; w <= crawler.Concurrency; w++ {
		go worker(crawler.Scraper, urls, recipes, nextUrls)
	}

	go recipeReceiver(crawler.Scraper, crawler.Store, recipes)
	go linksReceiver(crawler.Scraper, crawler.Store, nextUrls)

	crawler.Store.AddCrawlerJobs(crawler.Scraper.Name(), "", []string{startUrl})
	startUrls := crawler.Store.CrawlerJobByStatus(crawler.Scraper.Name(), CrawlerJobInProgress, -1)
	crawler.Store.UpdateCrawlerJobStatus(crawler.Scraper.Name(), startUrls, CrawlerJobToDo)

	for {
		if emptyTodo > TODO_IDLE_LOOP {
			fmt.Println("No work left.")
			break
		}

		time.Sleep(time.Duration(TODO_TIMEOUT_SECONDS) * time.Second)
		todo := crawler.Store.Todo(crawler.Scraper.Name(), crawler.Concurrency)

		if len(todo) == 0 {
			emptyTodo += 1
		} else {
			emptyTodo = 0
		}

		for _, url := range todo {
			urls <- url
		}
	}

	close(urls)
	close(recipes)
	close(nextUrls)

	os.Exit(0)
}
