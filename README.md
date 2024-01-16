# q7-backtester-go

This is the backtesting repository for the Q7 bot. 

TLDR: this takes a playbook & data and returns you results on the efficiency of your strategy.

# Inputs needed
## Playbook

The playbook is the agreed format for inputting setups with the Q7 team.

The playbook file needs to be in the same directory as the executable, and it needs to be named `Playbook.csv`

There are many columns in the playbook, as shown from this struct:

```go
// Row is a struct representing one processed row of a Playbook.
type Row struct {
// ID is the PlaybookID.
ID string

// Instrument is the instrument symbol we are testing.
Instrument string

// Day is the day of the setup.
Day string[]()

// Session is the session the setup is in, RDR,ODR,ADR.
Session string

// ConfirmationTime is the time.Time object containing
// when to look for a confirmation candle for this setup.
ConfirmationTime time.Time

// LongShort is a string containing either LONG or SHORT for the direction of the trade.
LongShort string

// RetStart is a legacy column.
RetStart string

// RetEnd is the time.Time object for when to actually place the trade if all other criteria are met.
RetEnd time.Time

// R1Low is a legacy column.
R1Low float64

// R1High is a legacy column.
R1High float64

// X1Start is a float representing the amount of standard deviations for a setup to place the target at.
X1Start float64

// MedianRetracement is a legacy column.
MedianRetracement float64

// MedianExtension is a legacy column.
MedianExtension float64

// MinR is a float representing the minimum required Risk/Return value to take a trade.
MinR float64

// PlusStop is a float representing the minimum amount of ticks to add to the stop.
PlusStop float64

// BreakEvenAt is a float representing the amount of standard deviations
// for progressing the trades stop to break even.
BreakEvenAt float64

// MinimumStopSize is a float representing the minimum amount of ticks the stop needs to be,
// if the stop is below that, then it is set to this value.
MinimumStopSize float64
}
```

We do not actually care about/use the following columns, they are simply legacy:

- RetStart
- R1Low
- R1High
- MedianRetracement
- MedianExtension

## Data

The data files are downloaded from the Q7 discord.

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
// Configuration is a struct representing a read in config.json object
type Configuration struct {
// The start date for backtesting (optional)
BacktestStartDate JsonDate `json:",omitempty"`

// The end date for backtesting (optional)
BacktestEndDate JsonDate `json:",omitempty"`

// Whether wick confirmation is enabled (optional)
WickConfirmation bool `json:",omitempty"`

// Maximum concurrent trades allowed (optional)
MaxConcurrentTrades int32 `json:",omitempty"`

// Whether to skip false days (optional)
SkipFalseDays bool `json:",omitempty"`

// Mapping of asset names to tick values (optional)
AssetTicks map[string]float64 `json:",omitempty"`
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
// PlaybookID is the ID of the playbook row used for the trade.
PlaybookID string `csv:"PlaybookID"`

// Instrument is the instrument symbol we traded.
Instrument string `csv:"Instrument"`

// TakenAt is a string representation of the timestamp in which we entered the trade.
TakenAt string `csv:"TakenAt"`

// Position is the direction of the trade, either LONG or SHORT.
Position string `csv:"Position"`

// InitiatedAtPrice is the price value in which we entered the trade.
InitiatedAtPrice float64 `csv:"InitiatedAtPrice"`

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
ClosedAtTime string `csv:"ClosedAtTime"`

// ConfirmedAtPrice is the price value at which the confirmation candle was marked as confirmed.
ConfirmedAtPrice float64 `csv:"ConfirmedAtPrice"`

// ConfirmedAtTime is a string representation of the timestamp in which the confirmation candle closed.
ConfirmedAtTime string `csv:"ConfirmedAtTime"`

// MoveToBreakEvenAt is the price value at which the trade is configured to move the stop to break even.
MoveToBreakEvenAt float64 `csv:"MoveToBreakEvenAt"`

// IDRHigh is the price value for the IDRHigh bounds.
IDRHigh float64 `csv:"IDRHigh"`

// IDRLow is the price value for the IDRLow bounds.
IDRLow float64 `csv:"IDRLow"`

// DRHigh is the price value for the DRHigh bounds.
DRHigh float64 `csv:"DRHigh"`

// DRLow is the price value for the DRLow bounds.
DRLow float64 `csv:"DRLow"`

// Win is a boolean column for if the trade was a winner or not.
Win bool `csv:"Win"`

// TakenAtDate is a string representation for the DATE part only of the TakenAt field.
TakenAtDate string `csv:"TakenAtDate"`

// TakenAtDate is a string representation for the TIME part only of the TakenAt field.
TakenAtTime string `csv:"TakenAtTime"`
}
```
