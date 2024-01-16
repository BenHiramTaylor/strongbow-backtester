package utils

import (
	"fmt"
	"math"
	"strings"
)

// RoundToDecimalLength takes two numbers, the first number is the one
// that you wish to round, the second number is the increment you wish to round to
// This has the output roundToDecimalLength(1234.567, 0.005) == 1234.565
func RoundToDecimalLength(numberA, numberB float64) float64 {
	// Calculate the multiplier based on numberB's decimal places
	decimalPlaces := getDecimalPlaces(numberB)
	multiplier := math.Pow(10, float64(decimalPlaces))

	// Normalize numberA and numberB to integers based on the number of decimal places in numberB
	normalizedA := numberA * multiplier
	normalizedB := numberB * multiplier

	// Round numberA to the nearest multiple of numberB
	roundedA := math.Round(normalizedA/normalizedB) * normalizedB

	// Return the rounded number in its original scale
	return roundedA / multiplier
}

// getDecimalPlaces is a logic encapsulation to get the amount of decimal points in a given number
func getDecimalPlaces(num float64) int {
	parts := fmt.Sprintf("%f", num)                 // Convert the number to a string
	decimalPart := strings.Split(parts, ".")[1]     // Split by the dot and take the decimal part
	return len(strings.TrimRight(decimalPart, "0")) // Trim the trailing zeros and get the length
}
