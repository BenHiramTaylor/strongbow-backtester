# strongbow-backtester

This is the backtesting repository for the Strongbow bot.

TLDR: This takes a config file defining some parameters per futures Asset, and simulates trading it live over the stored
historical data to determine its profitability.

# Inputs needed

## Data

This is data that has been historically downloaded from TradeStation and is stored in .csv format.

This is required to be inside of a `data` folder in the same directory as the executable.

These csv's are in the structure of:

```go
// Row is a struct that represents one candle of trade data
type Row struct {
// Time represents the time of the candles CLOSE.
Time time.Time

// Open represents the open price of the candles' interval.
Open float64

// High represents the highest price of the candles' interval.
High float64

// Low represents the lowest price of the candles' interval.
Low float64

// Close represents the close price of the candles' interval.
Close float64

// Volume represents the volume of transactions in the candles' interval.
Volume int
}
```

**The csv files must be named the same as the instrument we wish to test,
so if we wish to test ES then we need ES.csv in the data folder**

## config.json

The config.json file is not required to be on disk, if the file is not found then the default values are used

### Configuration Struct

The `Configuration` struct represents the configuration settings loaded from a `config.json` object.

```go
// InstrumentConfiguration is a struct representing a specific instrument configuration from config.json object
type InstrumentConfiguration struct {
// MinimumRR is the minimum Risk to Reward value to place a trade on.
MinimumRR float64 `json:"MinimumRR"`

// StopSizeAddition is the number of ticks to add to the stop size.
StopSizeAddition int `json:"StopSizeAddition"`

// TrailingStop is an integer for the amount of ticks to trail the stop by.
TrailingStop bool `json:"TrailingStop"`

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
```

# Outputs

## Backtesting results

The backtester outputs a CSV file with the write time as the name in the following format: `"2006-01-02T15:04:05.csv"`

These files are loaded into a folder called `backtesting_results` which is in the same directory as the executable, if
this folder is missing it will be created.

The results CSV has the following structure:

```go
// Row is a struct representing one row of the TradeLog output CSV
type Row struct {
// Instrument is the instrument symbol we traded.
Instrument string `csv:"Instrument"`

// TakenAt is a string representation of the timestamp in which we entered the trade.
TakenAt time.Time `csv:"TakenAt"`

// Direction is the direction of the trade, either LONG or SHORT.
Direction string `csv:"Direction"`

// EntryPrice is the price value in which we entered the trade.
EntryPrice float64 `csv:"EntryPrice"`

// StopPrice is the price value in which we had our stop set
// this could be different from InitialStopPrice if MoveToBreakEvenAt is configured and hit.
StopPrice float64 `csv:"StopPrice"`

// InitialStopPrice is the price value for our initial stop, this does not change.
InitialStopPrice float64 `csv:"InitialStopPrice"`

// TargetPrice is the price value for our target.
TargetPrice float64 `csv:"TargetPrice"`

// ClosedAtPrice is the price value we exited the trade at.
ClosedAtPrice float64 `csv:"ClosedAtPrice"`

// ClosedAtTime is a string representation of the timestamp in which we exited the trade.
ClosedAtTime time.Time `csv:"ClosedAtTime"`

// TakenAtDate is a string representation for the DATE part only of the TakenAt field.
TakenAtDate string `csv:"TakenAtDate"`

// TakenAtDate is a string representation for the TIME part only of the TakenAt field.
TakenAtTime string `csv:"TakenAtTime"`

// Win is a boolean column for if the trade was a winner or not.
Win bool `csv:"Win"`

// ProfitPercentage is a float representing the total gain or loss, 1.0 would be no change
Profit float32 `csv:"Profit"`
}
```
