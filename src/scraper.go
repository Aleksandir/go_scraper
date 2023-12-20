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
	})

	// Start scraping on https://scrapeme.live/shop/page/1 through to https://scrapeme.live/shop/page/48
	// concurrently using goroutines and a WaitGroup
	var wg sync.WaitGroup
	numPages := 48

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
