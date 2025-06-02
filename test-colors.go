package main

import (
	"fmt"
	"github.com/kjanat/go-chat-widget-dashboard/internal/models"
)

func main() {
	// Test the new brand colors format
	customer := &models.Customer{
		BrandColors: `primary: #007bff
secondary: #6c757d
background: #ffffff
text: #212529`,
	}

	// Test conversion to JSON
	jsonColors := customer.BrandColorsToJSON()
	fmt.Printf("Brand Colors as JSON: %s\n", jsonColors)

	// Test setting from JSON (backwards compatibility)
	customer2 := &models.Customer{}
	customer2.SetBrandColorsFromJSON(`{"primary": "#ff0000", "background": "#000000"}`)
	fmt.Printf("From JSON to simple format: %s\n", customer2.BrandColors)

	// Test with empty input
	customer3 := &models.Customer{}
	defaultJSON := customer3.BrandColorsToJSON()
	fmt.Printf("Default colors: %s\n", defaultJSON)
}
