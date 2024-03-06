package tradeConfig

import (
	"errors"
	"fmt"
	"github.com/BenHiramTaylor/strongbow-backtester/internal/backtestData"
	"github.com/BenHiramTaylor/strongbow-backtester/internal/utils"
	"github.com/rs/zerolog/log"
	"time"
)

// Trade represents one single placed trade
// Later outputted in the TradeLog
type Trade struct {
	// Instrument is the instrument this trade was placed on
	Instrument string

	// TakenAt was the time this trade was taken at, usually on the close of a candle
	TakenAt time.Time

	// Direction is the direction of the trade, either LONG or SHORT
	Direction string

	// EntryPrice is the price value in which we entered the trade
	EntryPrice float64

	// StopPrice is the price value in which we had our stop set
	StopPrice float64

	// InitialStopPrice is the price value for our initial stop, this does not change
	InitialStopPrice float64

	// TargetPrice is the price value for our target
	TargetPrice float64

	// ClosedAtPrice is the price value we exited the trade at
	ClosedAtPrice float64

	// ClosedAtTime is the time value we exited the trade at
	ClosedAtTime time.Time
}

// String is a stringer method for Trade
func (t Trade) String() string {
	return fmt.Sprintf(
		"Instrument: %s TakenAt: %v Direction: %s EntryPrice: %f StopPrice: %f "+
			"InitialStopPrice: %f TargetPrice: %f ClosedAtPrice: %f ClosedAtTime: %v",
		t.Instrument,
		t.TakenAt,
		t.Direction,
		t.EntryPrice,
		t.StopPrice,
		t.InitialStopPrice,
		t.TargetPrice,
		t.ClosedAtPrice,
		t.ClosedAtTime,
	)
}

// Trades is a slice of Trade pointers.
type Trades []*Trade

// newTrade is a generic constructor function that takes all
// the fields to create a new Trade, defaulting the ClosedAt fields.
func newTrade(
	instrument string,
	takenAt time.Time,
	direction string,
	entryPrice float64,
	stopPrice float64,
	targetPrice float64,
) *Trade {

	return &Trade{
		Instrument:       instrument,
		TakenAt:          takenAt,
		Direction:        direction,
		EntryPrice:       entryPrice,
		StopPrice:        stopPrice,
		InitialStopPrice: stopPrice,
		TargetPrice:      targetPrice,
	}
}

// calculateRR is a wrapper around some logic for calculating Risk to Reward values.
func calculateRR(oneRisk, targetSize float64) float64 {
	// Check the RR
	var actualRR float64
	if oneRisk != 0 {
		actualRR = targetSize / oneRisk
	} else {
		// Handle the division by zero case
		// e.g., set actualRR to a default value or log an error
		actualRR = 0 // or some other value
		log.Error().Msg("Error: Division by zero detected.")
	}

	return actualRR
}

// GenerateTradesInWindow takes a backtest data trade window, and applies multiple calculations
// and checks to generate a pointer to a Trade object, or an error.
func GenerateTradesInWindow(
	tradeWindow backtestData.Data,
	instrumentConfig *utils.InstrumentConfiguration, // Change the parameter to InstrumentConfiguration
	instrument string,
	tickSize float64,
) Trades {
	// Create variables
	var (
		inTrade = false
		trades  Trades
	)

	// Begin iteration of the trade window data
	for _, tradeRow := range tradeWindow {
		if inTrade {
			// Get the last trade of the trades slice and get its ClosedAtTime
			// Then check if the current time is equal to or after it and if so then we are out of that window
			lastTradeClosedAtTime := trades[len(trades)-1].ClosedAtTime
			if tradeRow.Time.Equal(lastTradeClosedAtTime) || tradeRow.Time.After(lastTradeClosedAtTime) {
				log.Info().Msgf("Trade closed at %v", tradeRow.Time)
				inTrade = false
			}

			log.Debug().Msgf("We are in a trade, not considering this row %v.", tradeRow.Time)
			continue
		}
		log.Debug().Str("rowTime", tradeRow.Time.Format("15:04")).Msg("Looking for trade at interval")

		// Get the tradeDirection based on the SMA values
		tradeDirection, err := tradeRow.TradeDirection()
		if err != nil {
			if errors.Is(err, backtestData.SMAValuesIntersect) {
				// Do nothing as this is not a valid time to trade.
				continue
			}
		}

		// Using the trade direction and levels, check if this is a valid candle to trade on.
		// If it is not a valid entry then skip
		if !tradeRow.IsValidEntry(tradeDirection) {
			log.Debug().Str("rowTime", tradeRow.Time.Format("15:04")).Msg("Not a valid entry.")
			continue
		}

		var (
			oneRisk    float64
			targetSize float64
			stopPrice  float64
			actualRR   float64
			target     *backtestData.Boundary
		)

		// Calculate Stop and risk sizes
		switch tradeDirection {
		case utils.TradeDirection.SHORT:
			// Get the stop by adding two ticks onto the high of the candle
			stopPrice = tradeRow.High
			stopPrice += tickSize * float64(instrumentConfig.StopSizeAddition)
			stopPrice = utils.RoundToDecimalLength(stopPrice, tickSize)

			// Get the first target, sort by descending as we want the highest low
			sortedBoundaries, err := tradeRow.LowBoundaries.GetSortedUnbrokenBoundary(false)
			if err != nil {
				if errors.Is(err, backtestData.NoBoundaryFound) {
					log.Debug().Msgf("No unbroken boundaries found in LowBoundaries for %v", tradeRow.Time)
					continue
				} else {
					log.Error().Msg(err.Error())
					return nil
				}
			}

			// Get the first target using the index
			target = (*sortedBoundaries)[0]

			// Get risk and target values
			oneRisk = stopPrice - tradeRow.Close
			targetSize = tradeRow.Close - target.Value

			// Calculate RR and check it against the minimum
			actualRR = calculateRR(oneRisk, targetSize)

		case utils.TradeDirection.LONG:
			// Get the stop by subtracting two ticks onto the low of the candle
			stopPrice = tradeRow.Low
			stopPrice -= tickSize * float64(instrumentConfig.StopSizeAddition)
			stopPrice = utils.RoundToDecimalLength(stopPrice, tickSize)

			// Get the first target sort by ascending as we want the lowest high
			sortedBoundaries, err := tradeRow.HighBoundaries.GetSortedUnbrokenBoundary(true)
			if err != nil {
				if errors.Is(err, backtestData.NoBoundaryFound) {
					log.Debug().Msgf("No unbroken boundaries found in HighBoundaries for %v", tradeRow.Time)
					continue
				} else {
					log.Error().Msg(err.Error())
					return nil
				}
			}

			// Get the first target using the index
			target = (*sortedBoundaries)[0]

			// Get risk and target values
			oneRisk = tradeRow.Close - stopPrice
			targetSize = target.Value - tradeRow.Close

			// Calculate RR and check it against the minimum
			actualRR = calculateRR(oneRisk, targetSize)

		default:
			// Return invalid error if not SHORT OR LONG
			log.Error().Msgf("Got invalid direction %s", tradeDirection)
			continue
		}

		// Skip this trade if RR is not met
		if actualRR < instrumentConfig.MinimumRR {
			log.Info().Msgf(
				"not taking trade at %v direction %s as it does not meet the minimum RR "+
					"specified by the user. Minimum: %f, Trade: %f Entry: %f Stop: %f Target: %f",
				tradeRow.Time,
				tradeDirection,
				instrumentConfig.MinimumRR,
				actualRR,
				tradeRow.Close,
				stopPrice,
				target.Value,
			)
			continue
		}

		// Info log that we are taking the trade
		log.Info().Msgf(
			"%v Taking trade at %f with a "+
				"target of %f and a stop of %f as it meets the users minimum RR of %f "+
				"with an RR of %f",
			tradeRow.Time,
			tradeRow.Close,
			target.Value,
			stopPrice,
			instrumentConfig.MinimumRR,
			actualRR,
		)

		// Create the new trade
		trade := newTrade(
			instrument,
			tradeRow.Time,
			tradeDirection,
			tradeRow.Close,
			stopPrice,
			target.Value,
		)

		// Set the flag that we are in a trade
		inTrade = true

		// Validate the trade
		trade.ValidateTradeWithWindow(tradeWindow, instrumentConfig, tickSize)

		// Add the trade to the results
		trades = append(trades, trade)
	}
	// If we have iterated through and not returned, then return with custom NoTradeFound error
	return trades
}

// ValidateTradeWithWindow iterates over a trade window and a trade configuration object.
// It determines if the trade hits the Stop/Target or expires at the end of the session.
func (t *Trade) ValidateTradeWithWindow(
	tradeWindow backtestData.Data,
	instrumentConfig *utils.InstrumentConfiguration,
	tickSize float64,
) {
	// Iterate over rows in the trade window
	for _, row := range tradeWindow {
		// Skip rows that occur before or at the time the trade was taken
		if row.Time.Before(t.TakenAt) || row.Time.Equal(t.TakenAt) {
			continue
		}

		// Determine the trade outcome based on the trade direction and price conditions
		switch {
		// If the trade is a LONG and the current row's low is less than or equal to the stop
		case t.Direction == utils.TradeDirection.LONG && row.Low <= t.StopPrice:
			// Stop condition met for a LONG trade
			log.Debug().Msgf("Stop condition met for a LONG trade at %v as Stop: %f Low: %f",
				row.Time,
				t.StopPrice,
				row.Low,
			)

			t.ClosedAtPrice = t.StopPrice
			t.ClosedAtTime = row.Time
			return // Exiting the loop as the trade is closed

		// If the trade is a SHORT and the current row's high is greater than or equal to the stop
		case t.Direction == utils.TradeDirection.SHORT && row.High >= t.StopPrice:
			// Stop condition met for a SHORT trade
			log.Debug().Msgf("Stop condition met for a SHORT trade at %v as Stop: %f High: %f",
				row.Time,
				t.StopPrice,
				row.High,
			)

			t.ClosedAtPrice = t.StopPrice
			t.ClosedAtTime = row.Time
			return // Exiting the loop as the trade is closed

		// If the trade is a LONG and the current row's high is greater than or equal to the target
		case t.Direction == utils.TradeDirection.LONG && row.High >= t.TargetPrice:
			// Target condition met for a LONG trade
			log.Debug().Msgf("Target condition met for a LONG trade at %v as Target: %f High: %f",
				row.Time,
				t.TargetPrice,
				row.High,
			)

			t.ClosedAtPrice = t.TargetPrice
			t.ClosedAtTime = row.Time
			return // Exiting the loop as the trade is closed

		// If the trade is a SHORT and the current row's low is less than or equal to the target
		case t.Direction == utils.TradeDirection.SHORT && row.Low <= t.TargetPrice:
			// Target condition met for a SHORT trade
			log.Debug().Msgf("Target condition met for a SHORT trade at %v as Target: %f Low: %f",
				row.Time,
				t.TargetPrice,
				row.Low,
			)
			t.ClosedAtPrice = t.TargetPrice
			t.ClosedAtTime = row.Time
			return // Exiting the loop as the trade is closed

		// If the TrailingStop is boolean
		case instrumentConfig.TrailingStop:
			// Dynamic Trailing Stop Logic
			if t.Direction == utils.TradeDirection.LONG {
				// Calculate the potential new stop price
				if row.High > t.EntryPrice {
					potentialNewStop := t.StopPrice + (row.High - t.EntryPrice)
					potentialNewStop = utils.RoundToDecimalLength(potentialNewStop, tickSize)
					// Update the stop price if the potential new stop is greater than the current stop price
					if potentialNewStop > t.StopPrice {
						t.StopPrice = potentialNewStop
					}
				}
			} else if t.Direction == utils.TradeDirection.SHORT {
				// Calculate the potential new stop price
				if row.Low < t.EntryPrice {
					potentialNewStop := t.StopPrice - (t.EntryPrice - row.Low)
					potentialNewStop = utils.RoundToDecimalLength(potentialNewStop, tickSize)
					// Update the stop price if the potential new stop is less than the current stop price
					if potentialNewStop < t.StopPrice {
						t.StopPrice = potentialNewStop
					}
				}
			}

			log.Debug().Msgf("Dynamic trailing stop adjusted to %f", t.StopPrice)

		// If the MoveToBreakEvenAt is set to a value greater than 0 and the stop price is not equal to the entry price (we have not changed it yet)
		case instrumentConfig.MoveToBreakEvenAt > 0 && t.StopPrice != t.EntryPrice:
			// Calculate the profit target to move to break even, based on the percentage specified in the configuration
			profitTarget := t.EntryPrice * (1 + instrumentConfig.MoveToBreakEvenAt/100)

			// Determine if the profit target is reached for LONG or SHORT trades
			profitTargetReached := false
			if t.Direction == utils.TradeDirection.LONG {
				// For a LONG trade, the profit target is reached when the high price exceeds the profit target
				profitTargetReached = row.High >= profitTarget
			} else if t.Direction == utils.TradeDirection.SHORT {
				// For a SHORT trade, the profit target is reached when the low price is less than or equal to the profit target
				profitTargetReached = row.Low <= profitTarget
			}

			// If the profit target is reached, move the stop price to the entry price
			if profitTargetReached {
				t.StopPrice = t.EntryPrice
				log.Debug().Msgf("Moved stop to break-even (entry price) at %f", t.StopPrice)
			}
		}
	}

	// If the trade does not hit the stop or target by the end of the window,
	// close the trade at the final price in the window.
	if t.ClosedAtPrice == 0 {
		lastRow := tradeWindow[len(tradeWindow)-1]
		log.Debug().Msgf("Trade did not hit stop or target, closing at %v with value of %f",
			lastRow.Time,
			lastRow.Close,
		)

		t.ClosedAtPrice = lastRow.Close
		t.ClosedAtTime = lastRow.Time
	}
}
