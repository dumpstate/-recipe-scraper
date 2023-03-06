package main

type IngredientsGroup struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
}

type Recipe struct {
	URL         string             `json:"url"`
	Name        string             `json:"name"`
	Portions    int32              `json:"portions"`
	Ingredients []IngredientsGroup `json:"ingredients"`
	Steps       []string           `json:"steps"`
	Imgs        []string           `json:"imgs"`
}

type Record[T any] struct {
	Id     int
	Record T
}
