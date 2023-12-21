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
	var pokemonProducts []PokemonProduct

	c := colly.NewCollector()

	c.OnHTML("li.product", func(e *colly.HTMLElement) {
		// create a new PokemonProduct struct
		pokemonProduct := PokemonProduct{}

		pokemonProduct.url = e.ChildAttr("a", "href")
		pokemonProduct.image = e.ChildAttr("img", "src")
		pokemonProduct.name = e.ChildText("h2")
		pokemonProduct.price = e.ChildText("span.price span.woocommerce-Price-amount.amount")

		pokemonProducts = append(pokemonProducts, pokemonProduct)
	})

	c.OnScraped(func(r *colly.Response) {
		writeProductsToCSV(pokemonProducts)
	})

	// Start scraping on https://scrapeme.live/shop/page/1 through to https://scrapeme.live/shop/page/48
	// concurrently using goroutines and a WaitGroup
	var wg sync.WaitGroup

	numPages := getNumPages("https://scrapeme.live/shop/page/1")

	println("Number of pages: ", numPages)

	// Create a new progress bar
	bar := pb.StartNew(numPages)

	for i := 1; i <= numPages; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Visit("https://scrapeme.live/shop/page/" + strconv.Itoa(i))

			// Increment the progress bar
			bar.Increment()
		}(i)
	}

	wg.Wait()

	// Finish the progress bar
	bar.Finish()
}

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
