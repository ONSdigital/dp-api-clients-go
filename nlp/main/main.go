package main

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin"
)

func main() {

	// Initialize the Berlin API client
	client := berlin.New("http://localhost:28900")

	// Create an Options struct and set a query parameter 'q'
	// you can also use url.Values directly into the Options
	options := berlin.OptInit()
	fmt.Println(options)

	options.Q("dentists in london with something")
	fmt.Println(options)

	// Get Berlin results using the created client and custom options
	results, err := client.GetBerlin(context.TODO(), options)
	fmt.Println(err)
	fmt.Println(results.Matches)
	fmt.Println(results.Matches[0].Scores.Score)
	fmt.Println(results.Matches[0].Scores.Offset)
	fmt.Println(results.Matches[0].Loc.Codes)
	fmt.Println(results.Matches[0].Loc.Encoding)
	fmt.Println(results.Matches[0].Loc.ID)
	fmt.Println(results.Query)
}
