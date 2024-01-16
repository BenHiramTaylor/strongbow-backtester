package tradeLog

import (
	"github.com/BenHiramTaylor/strongbow-backtester/internal/tradeConfig"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// TestNewLog tests the NewLog constructor function
func TestNewLog(t *testing.T) {
	log := NewLog()
	if log == nil {
		t.Errorf("NewLog() should not return nil")
	}
}

// TestAddRow tests the AddRow method of the Log struct
func TestAddRow(t *testing.T) {
	// Setup
	tradeLog := NewLog()
	mockTrade := &tradeConfig.Trade{
		Instrument:       "AAPL",
		TakenAt:          time.Now(),
		Direction:        "LONG",
		EntryPrice:       100.0,
		StopPrice:        95.0,
		InitialStopPrice: 95.0,
		TargetPrice:      110.0,
		ClosedAtPrice:    105.0,
		ClosedAtTime:     time.Now().Add(24 * time.Hour),
	}

	// Execution
	updatedLog := AddRow(tradeLog, mockTrade)

	// Validation
	require.Len(t, *updatedLog, 1, "Log should have one entry")
	addedRow := (*updatedLog)[0]
	require.Equal(t, mockTrade.Instrument, addedRow.Instrument, "Instrument should match")
	// ... more assertions for each field
}

// TestTotalWins tests the TotalWins method of the Log struct
func TestTotalWins(t *testing.T) {
	// Setup
	tradeLog := &Log{
		&Row{Win: true},
		&Row{Win: false},
		&Row{Win: true},
	}

	// Execution
	wins := tradeLog.TotalWins()

	// Validation
	require.Equal(t, 2, wins, "There should be 2 winning trades")
}

// TestCalculateCumulativeProfit tests the CalculateCumulativeProfit method of the Log struct
func TestCalculateCumulativeProfit(t *testing.T) {
	// Setup
	tradeLog := &Log{
		&Row{Profit: 10},
		&Row{Profit: -1},
		&Row{Profit: 15},
	}

	// Execution
	cumulativeProfit := tradeLog.CalculateCumulativeProfit()

	// Validation
	expectedProfit := float32(1.25235) // Adjust calculation as per logic
	require.Equal(t, expectedProfit, cumulativeProfit, "Cumulative profit should match expected")
}

// TestCalculateProfitValues tests the CalculateProfitValues method of the Log struct
func TestCalculateProfitValue(t *testing.T) {
	// Setup
	startingBalance := 1000.0
	tradeLog := &Log{
		&Row{Profit: 10},
		&Row{Profit: 10},
	}

	// Execution
	finalBalance := int64(tradeLog.CalculateProfitValue(startingBalance))

	// Validation
	expectedBalance := int64(1210) // Adjust calculation as per logic
	require.Equal(t, expectedBalance, finalBalance, "Final balance should match expected")
}
