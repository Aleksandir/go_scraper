package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
)

type Product struct {
	Name  string `json:"name"`
	Price string `json:"price"`
	URL   string `json:"url"`
}

func main() {
	// Read the source file
	data, err := ioutil.ReadFile("src/sources.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Split the file content into lines
	urls := strings.Split(string(data), "\n")

	// Create a new WebDriver instance for Firefox
	driver := agouti.GeckoDriver()
	if err := driver.Start(); err != nil {
		log.Fatal("Failed to start driver:", err)
	}
	defer driver.Stop()

	// Create a wait group
	var wg sync.WaitGroup

	// Create a channel to collect the results
	results := make(chan Product)

	// Launch a goroutine for each URL
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			scrapeAndSave(url, driver, results)
		}(url)
	}

	// Launch a goroutine to close the results channel after all other goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and save the results
	products := make(map[string]Product)
	for product := range results {
		products[product.Name] = product
	}
	saveData(products)
}

func scrapeAndSave(url string, driver *agouti.WebDriver, results chan<- Product) {
	// Create a new page
	page, err := driver.NewPage()
	if err != nil {
		log.Fatal("Failed to open page:", err)
	}

	// Navigate to the URL
	if err := page.Navigate(url); err != nil {
		log.Fatal("Failed to navigate:", err)
	}

	// Get the page content
	content, err := page.HTML()
	if err != nil {
		log.Fatal("Failed to get HTML:", err)
	}

	// Parse the page content with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		log.Fatal("Failed to parse HTML:", err)
	}

	// Scrape the product data
	doc.Find(".product-title").Each(func(i int, s *goquery.Selection) {
		name := s.Text()
		price := s.Parent().Find(".price-box .price").Text()
		url, _ := s.Attr("href")
		results <- Product{Name: name, Price: price, URL: url}
	})
}

func saveData(products map[string]Product) {
	// Convert the data to JSON
	data, err := json.MarshalIndent(products, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal data:", err)
	}

	// Write the data to a file
	if err := ioutil.WriteFile("products.json", data, 0644); err != nil {
		log.Fatal("Failed to write file:", err)
	}
}
