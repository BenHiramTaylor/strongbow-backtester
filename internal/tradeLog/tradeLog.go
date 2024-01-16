package tradeLog

import (
	"fmt"
	"github.com/BenHiramTaylor/strongbow-backtester/internal/tradeConfig"
	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

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

// Log is a slice of Row pointers, representing the TradeLog
type Log []*Row

// NewLog creates an empty Log object and returns it
func NewLog() *Log {
	var tradeLog Log
	return &tradeLog
}

// AddRow is a function that adds a row to a trade log, with some basic formatting on timestamps
// And some extra column calculations as per request from the Q7 team
func AddRow(l *Log, trade *tradeConfig.Trade) *Log {
	row := &Row{
		Instrument:       trade.Instrument,
		TakenAt:          trade.TakenAt,
		Direction:        trade.Direction,
		EntryPrice:       trade.EntryPrice,
		StopPrice:        trade.StopPrice,
		InitialStopPrice: trade.InitialStopPrice,
		TargetPrice:      trade.TargetPrice,
		ClosedAtPrice:    trade.ClosedAtPrice,
		ClosedAtTime:     trade.ClosedAtTime,
		Profit:           0,
	}

	// Split taken at date and time as per request from Q7 team
	row.TakenAtDate = trade.TakenAt.Format("2006-01-02")
	row.TakenAtTime = trade.TakenAt.Format("15:04:05")

	// Check if trade is win or not
	switch row.Direction {
	case "LONG":
		if row.ClosedAtPrice > row.EntryPrice {
			row.Win = true
		} else {
			row.Win = false
		}
	case "SHORT":
		if row.ClosedAtPrice < row.EntryPrice {
			row.Win = true
		} else {
			row.Win = false
		}
	default:
		// This should never happen
		log.Error().Msgf(
			"output trade does not have position LONG/SHORT, got: %s",
			row.Direction,
		)
	}

	if row.EntryPrice == trade.InitialStopPrice {
		// Avoid division by zero if InitiatedAt is equal to InitialStop.
		row.Profit = 0
	} else {
		// TODO add scaling by R (risk ratio - usually 1%, but sometimes not!)
		row.Profit = float32((trade.ClosedAtPrice - trade.EntryPrice) / (trade.EntryPrice - trade.InitialStopPrice))
	}

	newLog := append(*l, row)
	return &newLog
}

// TotalWins returns the total amount of winning trades in a trade log.
func (l *Log) TotalWins() int {
	var wins int

	for _, row := range *l {
		if row.Win {
			wins += 1
		}
	}

	return wins
}

// CalculateCumulativeProfit TODO FILL THIS IN ROB.
func (l *Log) CalculateCumulativeProfit() float32 {
	var cumulativeProfit float32 = 1.0 // Start with a base multiplier of 1.

	for _, row := range *l {
		cumulativeProfit *= 1 + (row.Profit / 100)
	}

	return cumulativeProfit
}

// SumTotalProfit is a function that returns the sum of the profit column.
func (l *Log) SumTotalProfit() float32 {
	var totalProfit float32

	for _, row := range *l {
		totalProfit += row.Profit
	}

	return totalProfit
}

// Write takes a Log pointer as a parameter and writes it to a CSV on disk
func Write(l *Log) error {
	// Ensure the directory exists
	dir := "backtesting_results"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			return err
		} // Create the directory with read/write permissions
	}

	// Create or overwrite the CSV file in the directory
	filePath := filepath.Join(dir, "results-"+time.Now().UTC().Format("2006-01-02-15_04_05")+".csv")

	file, err := os.Create(filePath)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			// Handle the error, but don't return it
			fmt.Printf("Error closing tradelog file: %v\n", err)
		}
	}(file)

	// Marshal the log into the CSV file
	err = gocsv.MarshalFile(l, file)
	if err != nil {
		return err
	}

	return nil
}

// CalculateProfitValue Simulates true cumulative profit by taking a starting balance and applying the trades profit
// in timestamp order, then returning the final value after all trades.
func (l *Log) CalculateProfitValue(startingBalance float64) float64 {
	// Return the starting balance if no trades in the log
	if len(*l) == 0 {
		return startingBalance
	}

	// Sort the trades by ClosedAtTime in ascending order
	sort.Slice(*l, func(i, j int) bool {
		return (*l)[i].ClosedAtTime.Before((*l)[j].ClosedAtTime)
	})

	balance := startingBalance
	for _, trade := range *l {
		// Calculate the profit/loss for each trade and update the balance
		balance *= float64(1 + trade.Profit/100)
	}

	return balance
}
