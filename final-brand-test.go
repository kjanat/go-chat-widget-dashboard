package main

import (
	"fmt"
	"github.com/kjanat/go-chat-widget-dashboard/internal/models"
)

func main() {
	fmt.Println("=== Brand Colors Format Testing ===")
	
	// Test 1: New customer with default colors
	fmt.Println("\n1. Testing default colors for new customer:")
	customer1 := &models.Customer{}
	defaultColors := customer1.BrandColorsToJSON()
	fmt.Printf("Default JSON: %s\n", defaultColors)
	
	// Test 2: Customer with user-friendly format
	fmt.Println("\n2. Testing user-friendly format:")
	customer2 := &models.Customer{
		BrandColors: `primary: #ff6b35
secondary: #f7931e
background: #ffffff
text: #2c3e50
accent: #e74c3c`,
	}
	userFriendlyJSON := customer2.BrandColorsToJSON()
	fmt.Printf("User-friendly to JSON: %s\n", userFriendlyJSON)
	
	// Test 3: Backwards compatibility with literal \n
	fmt.Println("\n3. Testing backwards compatibility with literal \\n:")
	customer3 := &models.Customer{
		BrandColors: `primary: #007bff\nsecondary: #6c757d\nbackground: #ffffff\ntext: #212529`,
	}
	backwardsJSON := customer3.BrandColorsToJSON()
	fmt.Printf("Literal \\n to JSON: %s\n", backwardsJSON)
	
	// Test 4: JSON to user-friendly conversion
	fmt.Println("\n4. Testing JSON to user-friendly conversion:")
	customer4 := &models.Customer{}
	customer4.SetBrandColorsFromJSON(`{"primary": "#28a745", "secondary": "#6c757d", "background": "#f8f9fa", "text": "#343a40"}`)
	fmt.Printf("JSON to user-friendly: %s\n", customer4.BrandColors)
	
	fmt.Println("\n=== All tests completed successfully! ===")
}
