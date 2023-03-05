package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
)

func findName(doc soup.Root) string {
	return doc.Find("h1", "class", "page-header").FullText()
}

func findPortions(ingr soup.Root) int32 {
	portionsDiv := ingr.Find("div", "class", "field-name-field-ilosc-porcji")
	if portionsDiv.Pointer == nil {
		return -1
	}

	split := strings.Fields(portionsDiv.FullText())
	if len(split) == 0 {
		return -1
	}

	portions, err := strconv.Atoi(split[0])
	if err != nil {
		portions = -1
	}

	return int32(portions)
}

func findIngredients(ingr soup.Root) []IngredientsGroup {
	res := []IngredientsGroup{}
	div := ingr.Find("div", "class", "field-name-field-skladniki")
	groupName := ""

	for _, child := range div.Children() {
		items := child.FindAll("li")
		if len(items) > 0 {
			group := []string{}
			for _, li := range items {
				group = append(group, strings.TrimSpace(li.FullText()))
			}
			res = append(res, IngredientsGroup{
				Name:        groupName,
				Ingredients: group,
			})
		} else {
			if strings.Contains(child.Attrs()["class"], "wyroznione") {
				groupName = strings.TrimSpace(child.FullText())
			}
		}
	}

	return res
}

func findSteps(rcp soup.Root) []string {
	res := []string{}
	items := rcp.Find("div", "class", "field-name-field-przygotowanie").FindAll("li")

	for _, li := range items {
		res = append(res, strings.TrimSpace(li.FullText()))
	}

	return res
}

func findImgLinks(view soup.Root) []string {
	res := []string{}

	for _, img := range view.FindAll("img") {
		res = append(res, img.Attrs()["src"])
	}

	return res
}

type KwestiaSmakuScraper struct{}

func (scraper *KwestiaSmakuScraper) Name() string {
	return scraper.Domain()
}

func (scraper *KwestiaSmakuScraper) Domain() string {
	return "kwestiasmaku.com"
}

func (scraper *KwestiaSmakuScraper) TryFind(url string) (*Recipe, string, error) {
	fmt.Printf("Scraping %s\n", url)

	resp, err := soup.Get(url)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to fetch %s", url)
	}

	doc := soup.HTMLParse(resp)
	ingr := doc.Find("div", "class", "group-skladniki")
	if ingr.Pointer == nil {
		return nil, resp, fmt.Errorf("group-skladniki not found; not a recipe: %s", url)
	}

	rcp := doc.Find("div", "class", "group-przepis")
	if rcp.Pointer == nil {
		return nil, resp, fmt.Errorf("group-przepis not found; not a recipe: %s", url)
	}

	view := doc.Find("div", "class", "view-content")
	if view.Pointer == nil {
		return nil, resp, fmt.Errorf("view-content not found; not a recipe: %s", url)
	}

	return &Recipe{
		URL:         url,
		Name:        findName(doc),
		Portions:    findPortions(ingr),
		Ingredients: findIngredients(ingr),
		Steps:       findSteps(rcp),
		Imgs:        findImgLinks(view),
	}, resp, nil
}
