package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DefaultBrandColors defines the default color palette for customers and widgets.
var DefaultBrandColors = map[string]string{
	"primary":    "#007bff",
	"secondary":  "#6c757d",
	"background": "#ffffff",
	"text":       "#212529",
}

var defaultBrandColorsOrder = []string{"primary", "secondary", "background", "text"}

// DefaultBrandColorsJSON returns the default colors as a JSON string.
func DefaultBrandColorsJSON() string {
	bytes, _ := json.Marshal(DefaultBrandColors)
	return string(bytes)
}

// DefaultBrandColorsSimple returns the default colors in key-value newline format.
func DefaultBrandColorsSimple() string {
	var lines []string
	for _, key := range defaultBrandColorsOrder {
		lines = append(lines, fmt.Sprintf("%s: %s", key, DefaultBrandColors[key]))
	}
	return strings.Join(lines, "\n")
}
