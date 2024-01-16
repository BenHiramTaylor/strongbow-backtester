package utils

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

// JsonDate is a struct specifically to implement custom Unmarshalling on read.
type JsonDate struct {
	time.Time
}

// InstrumentConfiguration is a struct representing a specific instrument configuration from config.json object
type InstrumentConfiguration struct {
	// MinimumRR is the minimum Risk to Reward value to place a trade on.
	MinimumRR float64 `json:"MinimumRR"`

	// StopSizeAddition is the number of ticks to add to the stop size.
	StopSizeAddition int `json:"StopSizeAddition"`

	// TrailingStopAmount is an integer for the amount of ticks to trail the stop by.
	TrailingStopAmount int `json:"TrailingStopAmount,omitempty"`

	// LargeSMALookbackAmount is the amount of candles to lookback and create a larger rolling moving average with.
	LargeSMALookbackAmount int `json:"LargeSMALookbackAmount"`

	// LargeSMALookbackAmount is the amount of candles to lookback and create a smaller rolling moving average with.
	SmallSMALookbackAmount int `json:"SmallSMALookbackAmount"`

	// UnbrokenBoundaryLeftBars is an integer for the amount of candles to look back on the left side to get
	// the unbroken highs and lows using Pivots.
	UnbrokenBoundaryLeftBars int `json:"UnbrokenBoundaryLeftBars"`

	// UnbrokenBoundaryRightBars is an integer for the amount of candles to look back on the right side to get
	// the unbroken highs and lows using Pivots.
	UnbrokenBoundaryRightBars int `json:"UnbrokenBoundaryRightBars"`

	// UnbrokenBoundaryMemoryLimit is the amount of candles to remember for unbroken highs/lows
	// for example: so if that value is 50 then any unbroken highs/lows are dropped until we are down to 50
	// starting with the oldest first.
	UnbrokenBoundaryMemoryLimit int `json:"UnbrokenBoundaryMemoryLimit"`

	// StochasticUpperBand is the upper band for the stochastic oscillator
	StochasticUpperBand float64 `json:"StochasticUpperBand"`

	// StochasticLowerBand is the lower band for the stochastic oscillator
	StochasticLowerBand float64 `json:"StochasticLowerBand"`

	// StochasticKPeriods is the number of periods to use for the stochastic oscillator
	StochasticKPeriods int `json:"StochasticKPeriods"`

	// StochasticDPeriods is the number of periods to use for the stochastic oscillator
	StochasticDPeriods int `json:"StochasticDPeriods"`

	// MoveToBreakEvenAt is a float64 representing a percentage of profit to move the stop to break even at.
	MoveToBreakEvenAt float64 `json:"MoveToBreakEvenAt,omitempty"`
}

// Configuration is a struct representing a read in config.json object
type Configuration struct {
	// The start date for back-testing (optional)
	BacktestStartDate JsonDate `json:"BacktestStartDate,omitempty"`

	// The end date for back-testing (optional)
	BacktestEndDate JsonDate `json:"BacktestEndDate,omitempty"`

	// Instruments to backtest on
	Instruments map[string]*InstrumentConfiguration `json:"instruments"`

	// StartingBalance is a number to use as a balance to applied simulated profit to. (optional defaults to 10k)
	StartingBalance float64 `json:"StartingBalance,omitempty"`

	// Mapping of asset names to tick values (optional)
	AssetTicks map[string]float64 `json:"AssetTicks,omitempty"`

	// WriteProcessedDataToFile is a boolean to write the processed data to a file (optional defaults to false)
	WriteProcessedDataToFile bool `json:"WriteProcessedDataToFile,omitempty"`
}

// newConfiguration This constructor is just an idiomatic wrapper to create default values for fields.
func newConfiguration() *Configuration {
	today := time.Now().Add(time.Hour * 24)
	return &Configuration{
		BacktestStartDate: JsonDate{
			Time: time.Date(
				2020,
				1,
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			),
		},
		BacktestEndDate: JsonDate{
			time.Date(
				today.Year(),
				today.Month(),
				today.Day(),
				0,
				0,
				0,
				0,
				time.UTC,
			),
		},
		StartingBalance: 10000.0,
	}
}

// Keys is a pointer method that returns a string slice of instrument names
func (c *Configuration) Keys() []string {
	names := make([]string, 0, len(c.Instruments))

	for instrumentName := range c.Instruments {
		names = append(names, instrumentName)
	}

	return names
}

// UnmarshalJSON Implements Unmarshal interface for JsonDate
// This is just so that we can use a custom time format
func (j *JsonDate) UnmarshalJSON(b []byte) error {
	var err error

	s := strings.Trim(string(b), "\"")
	j.Time, err = time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	return nil
}

// LoadConfiguration loads the config.json file from the file path
// There are some default variables for this if they are not present in config.json
func LoadConfiguration(filePath string) (*Configuration, error) {
	var cfg = newConfiguration()

	configurationFile, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn().Msg("config.json does not exist, using the default values")
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configurationFile, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
