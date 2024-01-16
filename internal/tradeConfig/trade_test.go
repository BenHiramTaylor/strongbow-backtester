package tradeConfig

import (
	"github.com/BenHiramTaylor/strongbow-backtester/internal/backtestData"
	"github.com/BenHiramTaylor/strongbow-backtester/internal/utils"
	"reflect"
	"testing"
	"time"
)

// TestTradeString tests the Stringer method of the Trade object
func TestTradeString(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		args *Trade
		want string
	}{
		{
			name: "basic test",
			args: &Trade{
				Instrument:       "INSTR",
				TakenAt:          now,
				Direction:        "LONG",
				EntryPrice:       100.0,
				StopPrice:        95.0,
				InitialStopPrice: 95.0,
				TargetPrice:      110.0,
			},
			want: "Instrument: INSTR TakenAt: " + now.String() + " Direction: LONG EntryPrice: 100.000000 StopPrice: 95.000000 InitialStopPrice: 95.000000 TargetPrice: 110.000000 ClosedAtPrice: 0.000000 ClosedAtTime: 0001-01-01 00:00:00 +0000 UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.String() != tt.want {
				t.Errorf("Trade.String() = %v, want %v", tt.args.String(), tt.want)
			}
		})
	}
}

// TestNewTrade tests the newTrade constructor function
func TestNewTrade(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		args *Trade
		want *Trade
	}{
		{
			name: "basic test",
			args: &Trade{
				Instrument:       "INSTR",
				TakenAt:          now,
				Direction:        "LONG",
				EntryPrice:       100.0,
				StopPrice:        95.0,
				InitialStopPrice: 95.0,
				TargetPrice:      110.0,
			},
			want: newTrade(
				"INSTR",
				now,
				"LONG",
				100.0,
				95.0,
				110.0,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.args, tt.want) {
				t.Errorf("newTrade() = %v, want %v", tt.args, tt.want)
			}
		})
	}
}

// TestValidateTradeWithWindow Contains test cases for the TestValidateTradeWithWindow function
// TODO MAYBE IMPROVE THESE TEST CASES, THEY WORK BUT ARE MINIMAL FOR THE DEVELOPER
func TestValidateTradeWithWindow(t *testing.T) {
	// Define a common time layout
	layout := "2006-01-02T15:04:05.000Z"
	time1, _ := time.Parse(layout, "2023-10-20T09:00:00.000Z")
	time2, _ := time.Parse(layout, "2023-10-20T10:00:00.000Z")
	time3, _ := time.Parse(layout, "2023-10-20T11:00:00.000Z")

	tests := []struct {
		name                  string
		trade                 *Trade
		data                  backtestData.Data
		instrumentConfig      *utils.InstrumentConfiguration
		tickSize              float64
		expectedClosedAtPrice float64
		expectedClosedAtTime  time.Time
		expectedStopPrice     float64
	}{
		{
			name: "Test LONG position hits STOP",
			trade: &Trade{
				TakenAt:     time2,
				Direction:   "LONG",
				EntryPrice:  100,
				StopPrice:   95,
				TargetPrice: 110,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, Low: 96},
				&backtestData.Row{Time: time2, Low: 94},
				&backtestData.Row{Time: time3, Low: 91},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{},
			tickSize:              0.5,
			expectedClosedAtPrice: 95,
			expectedClosedAtTime:  time3,
			expectedStopPrice:     95,
		},
		{
			name: "Test LONG position hits TARGET",
			trade: &Trade{
				TakenAt:     time2,
				Direction:   "LONG",
				EntryPrice:  100,
				StopPrice:   95,
				TargetPrice: 110,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, Low: 96, High: 100},
				&backtestData.Row{Time: time2, Low: 96, High: 101},
				&backtestData.Row{Time: time3, Low: 96, High: 150},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{},
			tickSize:              0.5,
			expectedClosedAtPrice: 110,
			expectedClosedAtTime:  time3,
			expectedStopPrice:     95,
		},
		{
			name: "Test SHORT position hits TARGET",
			trade: &Trade{
				TakenAt:     time2,
				Direction:   "SHORT",
				EntryPrice:  110,
				StopPrice:   115,
				TargetPrice: 105,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, High: 111, Low: 109},
				&backtestData.Row{Time: time2, High: 112, Low: 104},
				&backtestData.Row{Time: time3, High: 113, Low: 100},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{},
			tickSize:              0.5,
			expectedClosedAtPrice: 105,
			expectedClosedAtTime:  time3,
			expectedStopPrice:     115,
		},
		{
			name: "Test SHORT position hits STOP",
			trade: &Trade{
				TakenAt:     time2,
				Direction:   "SHORT",
				EntryPrice:  110,
				StopPrice:   115,
				TargetPrice: 105,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, High: 111, Low: 109},
				&backtestData.Row{Time: time2, High: 112, Low: 104},
				&backtestData.Row{Time: time3, High: 120, Low: 100},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{},
			tickSize:              0.5,
			expectedClosedAtPrice: 115,
			expectedClosedAtTime:  time3,
			expectedStopPrice:     115,
		},
		{
			name: "Test position closes at the end",
			trade: &Trade{
				TakenAt:     time2,
				Direction:   "LONG",
				EntryPrice:  110,
				StopPrice:   105,
				TargetPrice: 115,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, High: 111, Low: 109, Close: 110},
				&backtestData.Row{Time: time2, High: 112, Low: 104, Close: 110},
				&backtestData.Row{Time: time3, High: 112, Low: 107, Close: 111},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{},
			tickSize:              0.5,
			expectedClosedAtPrice: 111,
			expectedClosedAtTime:  time3,
			expectedStopPrice:     105,
		},
		{
			name: "Test SHORT position trailing STOP",
			trade: &Trade{
				TakenAt:     time1,
				Direction:   "SHORT",
				EntryPrice:  100,
				StopPrice:   300,
				TargetPrice: 50,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, High: 111, Low: 90, Open: 105, Close: 100},
				&backtestData.Row{Time: time2, High: 105, Low: 80, Open: 100, Close: 90},
				&backtestData.Row{Time: time3, High: 90, Low: 60, Open: 90, Close: 75},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{TrailingStopAmount: 50},
			tickSize:              0.5,
			expectedClosedAtPrice: 75,
			expectedClosedAtTime:  time3,
			expectedStopPrice:     85, // This is the Low of the last candle (60) + (50 ticks at 0.5 tick size for 25)
		},
		{
			name: "Test LONG position trailing STOP",
			trade: &Trade{
				TakenAt:     time1,
				Direction:   "LONG",
				EntryPrice:  100,
				StopPrice:   50,
				TargetPrice: 300,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, High: 111, Low: 90, Open: 105, Close: 100},
				&backtestData.Row{Time: time2, High: 105, Low: 80, Open: 100, Close: 90},
				&backtestData.Row{Time: time3, High: 90, Low: 60, Open: 90, Close: 75},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{TrailingStopAmount: 50},
			tickSize:              0.5,
			expectedClosedAtPrice: 80,
			expectedClosedAtTime:  time3,
			expectedStopPrice:     80, // This is the low of the second candle as the trailing stop on the first candle would be 50 ticks away from the high
		},
		{
			name: "Test LONG position move to BE at 50%",
			trade: &Trade{
				TakenAt:     time1,
				Direction:   "LONG",
				EntryPrice:  100,
				StopPrice:   50,
				TargetPrice: 300,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, High: 111, Low: 90, Open: 105, Close: 100},
				&backtestData.Row{Time: time2, High: 215, Low: 110, Open: 120, Close: 150},
				&backtestData.Row{Time: time3, High: 250, Low: 130, Open: 180, Close: 150},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{MoveToBreakEvenAt: 50},
			tickSize:              0.5,
			expectedClosedAtPrice: 150, // Close at end of session
			expectedClosedAtTime:  time3,
			expectedStopPrice:     100, // This is the entry price
		},
		{
			name: "Test SHORT position move to BE at 50%",
			trade: &Trade{
				TakenAt:     time1,
				Direction:   "SHORT",
				EntryPrice:  200,
				StopPrice:   300,
				TargetPrice: 100,
			},
			data: backtestData.Data{
				&backtestData.Row{Time: time1, High: 111, Low: 90, Open: 105, Close: 100},
				&backtestData.Row{Time: time2, High: 180, Low: 160, Open: 120, Close: 150},
				&backtestData.Row{Time: time3, High: 180, Low: 130, Open: 180, Close: 150},
			},
			instrumentConfig:      &utils.InstrumentConfiguration{MoveToBreakEvenAt: 50},
			tickSize:              0.5,
			expectedClosedAtPrice: 150, // Close at end of session
			expectedClosedAtTime:  time3,
			expectedStopPrice:     200, // This is the entry price
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.trade.ValidateTradeWithWindow(tt.data, tt.instrumentConfig, tt.tickSize)

			if tt.trade.ClosedAtPrice != tt.expectedClosedAtPrice {
				t.Errorf("expected ClosedAtPrice: %v, got: %v", tt.expectedClosedAtPrice, tt.trade.ClosedAtPrice)
			}
			if !tt.trade.ClosedAtTime.Equal(tt.expectedClosedAtTime) {
				t.Errorf("expected ClosedAtTime: %v, got: %v", tt.expectedClosedAtTime, tt.trade.ClosedAtTime)
			}
			if tt.trade.StopPrice != tt.expectedStopPrice {
				t.Errorf("expected StopPrice: %v, got %v", tt.expectedStopPrice, tt.trade.StopPrice)
			}
		})
	}
}
