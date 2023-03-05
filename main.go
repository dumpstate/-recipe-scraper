package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func domain(host string) string {
	if strings.HasPrefix(host, "www.") {
		return host[4:]
	}

	return host
}

func prettyPrint(data interface{}) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", string(b))
}

func main() {
	var inputUrl string
	var output string
	var concurrency int

	locator := NewScraperLocator()

	app := &cli.App{
		Name:  "recipe-scraper",
		Usage: "recipe scraper",
		Commands: []*cli.Command{
			{
				Name:    "single",
				Aliases: []string{"s"},
				Usage:   "scrap single recipe",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "url",
						Usage:       "target URL",
						Required:    true,
						Destination: &inputUrl,
					},
				},
				Action: func(ctx *cli.Context) error {
					parsedUrl, err := url.Parse(inputUrl)
					if err != nil {
						return err
					}

					scraper, err := locator.FindScraper(domain(parsedUrl.Host))
					if err != nil {
						return err
					}

					recipe, _, err := scraper.TryFind(parsedUrl.String())
					if err != nil {
						return err
					}

					prettyPrint(recipe)

					return nil
				},
			},
			{
				Name:    "crawl",
				Aliases: []string{"c"},
				Usage:   "crawl website",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "url",
						Usage:       "target URL",
						Required:    true,
						Destination: &inputUrl,
					},
					&cli.StringFlag{
						Name:        "out",
						Usage:       "target SQLite database path",
						Destination: &output,
						Required:    true,
					},
					&cli.IntFlag{
						Name:        "concurrency",
						Usage:       "crawler concurrency",
						Destination: &concurrency,
						Required:    true,
					},
				},
				Action: func(ctx *cli.Context) error {
					parsedUrl, err := url.Parse(inputUrl)
					if err != nil {
						return err
					}

					scraper, err := locator.FindScraper(domain(parsedUrl.Host))
					if err != nil {
						return err
					}

					store := NewStore(output)
					defer store.Close()
					crwlr := NewCrawler(scraper, store, concurrency)
					crwlr.Crawl(parsedUrl.String())

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
