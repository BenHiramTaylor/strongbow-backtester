package utils

// AssetTicks is a mapping of Instrument -> TickSize - measured in index points
var AssetTicks = map[string]float64{
	// S&P 500
	"ES":  0.25,
	"MES": 0.25,
	// Nasdaq
	"NQ":  0.25,
	"MNQ": 0.25,
	// Euro
	"EC":  0.00005,
	"M6E": 0.0001,
	// Crude oil
	"CL":  0.01,
	"MCL": 0.01,
	// Gold
	"GC":  0.1,
	"MGC": 0.1,
	// Yen
	"6J":  0.0000005,
	"M6J": 0.01,
	// Pound
	"BP":  0.0001,
	"M6B": 0.0001,
	// Australian Dollar
	"AD":  0.00005,
	"M6A": 0.0001,
}
