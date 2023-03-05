# recipe-scraper

Cooking recipe website scraper.

## Usage

1. Build the scraper.

```sh
make
```

2. Run the scraper.

Crawl the whole website:

```sh
./bin/recipe-scraper crawl --url "https://kwestiasmaku.com" --out "./path/to/sqlite.db" --concurrency 10
```

Scrape a single page:

```sh
./bin/recipe-scraper single --url "...recipe url..."
```
