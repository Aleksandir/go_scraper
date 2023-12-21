package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/gocolly/colly"
)

type PokemonProduct struct {
	url, image, name, price string
}

func main() {
	// create a slice of PokemonProduct structs
	var pokemonProducts []PokemonProduct

	// create a new collector
	c := colly.NewCollector()

	// On every a element which has href attribute call callback
	c.OnHTML("li.product", func(e *colly.HTMLElement) {
		// create a new PokemonProduct struct
		pokemonProduct := PokemonProduct{}

		pokemonProduct.url = e.ChildAttr("a", "href")
		pokemonProduct.image = e.ChildAttr("img", "src")
		pokemonProduct.name = e.ChildText("h2")
		pokemonProduct.price = e.ChildText("span.price span.woocommerce-Price-amount.amount")

		pokemonProducts = append(pokemonProducts, pokemonProduct)
	})

	//when c finishes scraping, call the function to write the products to a CSV file
	c.OnScraped(func(r *colly.Response) {
		writeProductsToCSV(pokemonProducts)
	})

	// concurrently using goroutines and a WaitGroup
	var wg sync.WaitGroup

	// get the number of pages
	numPages := getNumPages("https://scrapeme.live/shop/page/1")

	// Print the number of pages
	println("Number of pages: ", numPages)

	// Create a new progress bar
	bar := pb.StartNew(numPages)

	// Visit each page
	for i := 1; i <= numPages; i++ {
		wg.Add(1)
		go func(i int) {
			// Decrement the WaitGroup counter when the goroutine completes
			defer wg.Done()
			c.Visit("https://scrapeme.live/shop/page/" + strconv.Itoa(i))

			// Increment the progress bar
			bar.Increment()
		}(i)
	}

	// Wait for all HTTP requests to finish
	wg.Wait()

	// Finish the progress bar
	bar.Finish()
}

// getNumPages is a function that retrieves the maximum number of pages from a given URL.
// It uses the colly library to scrape the HTML and extract the number of pages.
// The URL parameter specifies the URL to scrape.
// The function returns the maximum number of pages as an integer.
func getNumPages(url string) int {
	// Create a new collector for getting the max number of pages
	c1 := colly.NewCollector()

	var numPages int

	c1.OnHTML("ul.page-numbers li:nth-last-child(2) a", func(e *colly.HTMLElement) {
		numPages, _ = strconv.Atoi(e.Text)
	})

	c1.Visit(url)

	return numPages
}

// writeProductsToCSV writes the given Pokemon products to a CSV file.
// It takes a slice of PokemonProduct structs as input and creates a CSV file named "products.csv".
// The function writes the headers and then iterates over each PokemonProduct to write its URL, image, name, and price to the CSV file.
// Finally, it flushes the writer to ensure all data is written to the file.
func writeProductsToCSV(pokemonProducts []PokemonProduct) {
	file, err := os.Create("products.csv")
	if err != nil {
		log.Fatalln("Failed to create output CSV file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	headers := []string{
		"url",
		"image",
		"name",
		"price",
	}
	writer.Write(headers)

	for _, pokemonProduct := range pokemonProducts {
		record := []string{
			pokemonProduct.url,
			pokemonProduct.image,
			pokemonProduct.name,
			pokemonProduct.price,
		}

		writer.Write(record)
	}

	writer.Flush()
}
