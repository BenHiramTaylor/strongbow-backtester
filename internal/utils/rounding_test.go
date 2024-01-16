package utils

import (
	"testing"
)

func TestRoundToDecimalLength(t *testing.T) {
	tests := []struct {
		numberA, numberB, want float64
	}{
		{1234.567, 0.005, 1234.565},
		{1234.567, 0.01, 1234.57},
		{1234.567, 0.1, 1234.6},
		{0.01234, 0.005, 0.01},
		{0.01234, 0.001, 0.012},
	}

	for _, tt := range tests {
		got := RoundToDecimalLength(tt.numberA, tt.numberB)
		if got != tt.want {
			t.Errorf("RoundToDecimalLength(%f, %f) = %f; want %f", tt.numberA, tt.numberB, got, tt.want)
		}
	}
}

func TestGetDecimalPlaces(t *testing.T) {
	tests := []struct {
		num  float64
		want int
	}{
		{0.005, 3},
		{0.01, 2},
		{0.1, 1},
		{0.01234, 5},
		{0.00001, 5},
	}

	for _, tt := range tests {
		got := getDecimalPlaces(tt.num)
		if got != tt.want {
			t.Errorf("getDecimalPlaces(%f) = %d; want %d", tt.num, got, tt.want)
		}
	}
}
