package main

import (
	"fmt"
)

type Scraper interface {
	Name() string
	Domain() string
	TryFind(path string) (*Recipe, string, error)
	SkipPrefixes() []string
}

type ScraperLocator struct {
	Scrapers []Scraper
}

func NewScraperLocator() *ScraperLocator {
	return &ScraperLocator{
		Scrapers: []Scraper{
			&KwestiaSmakuScraper{},
		},
	}
}

func (locator *ScraperLocator) FindScraper(domain string) (Scraper, error) {
	if locator.Scrapers == nil {
		return nil, fmt.Errorf("no registered scrapers")
	}

	for _, scraper := range locator.Scrapers {
		if scraper.Domain() == domain {
			return scraper, nil
		}
	}

	return nil, fmt.Errorf("no scaper for domain: %s", domain)
}
