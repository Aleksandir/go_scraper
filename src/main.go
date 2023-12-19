package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Product struct {
	Name  string `json:"name"`
	Price string `json:"price"`
	URL   string `json:"url"`
}

func getSearchPages(file string) []string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(string(data), "\n")
}

func getDocument(url string) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func saveData(data map[string]Product) {
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile("products.json", file, 0644)
}

func scrapeProductCategoryPage(url string) map[string]Product {
	doc := getDocument(url)
	products := make(map[string]Product)
	doc.Find(".product-title").Each(func(i int, s *goquery.Selection) {
		name := s.Text()
		price := s.Find(".price-box .price").Text()
		url, _ := s.Attr("href")
		products[name] = Product{Name: name, Price: price, URL: url}
	})
	return products
}

func main() {
	SOURCES_FILE := "src/sources.txt"
	fmt.Println("Scraping PCCaseGear...")

	siteUrls := getSearchPages(SOURCES_FILE)

	allProducts := make(map[string]Product)
	var wg sync.WaitGroup
	for _, url := range siteUrls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			products := scrapeProductCategoryPage(url)
			for k, v := range products {
				allProducts[k] = v
			}
		}(url)
	}
	wg.Wait()

	saveData(allProducts)
	fmt.Printf("\n%d products scraped and saved in total.\n", len(allProducts))
}
