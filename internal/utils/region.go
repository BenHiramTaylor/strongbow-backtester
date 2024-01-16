package utils

import (
	"fmt"
)

// region is a struct representing one of the regions
// this is private as it is static.
type region struct {
	RegionName  string
	MarketOpen  string
	MarketClose string
}

// String is here so that printing can be done correctly with pointers.
func (r region) String() string {
	return fmt.Sprintf(
		"RegionName: %s MarketOpen: %s MarketClose: %s",
		r.RegionName,
		r.MarketOpen,
		r.MarketClose,
	)
}

// standardConfiguration is a struct wrapper around a slice of Regions
type standardConfiguration struct{ Regions []*region }

var (
	// StandardConfiguration The MarketOpen times are 5 minutes ahead due to TradeStation using close times for its candles
	// and the backtester data using start times for their candles, so we essentially shift open > close
	// These times are in NYC Location time as the Algo works for the first hour of the market open for NYC.
	StandardConfiguration = standardConfiguration{
		Regions: []*region{
			{
				RegionName:  "New York",
				MarketOpen:  "02:00",
				MarketClose: "16:00",
			},
		},
	}
)
