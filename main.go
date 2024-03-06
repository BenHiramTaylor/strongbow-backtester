package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/BenHiramTaylor/strongbow-backtester/internal/backtestData"
	"github.com/BenHiramTaylor/strongbow-backtester/internal/tradeConfig"
	"github.com/BenHiramTaylor/strongbow-backtester/internal/tradeLog"
	"github.com/BenHiramTaylor/strongbow-backtester/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"sync"
	"time"
)

func main() {
	// Initialise Logger values
	logFile, _ := os.OpenFile("back-tester.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	writers := []io.Writer{
		&zerolog.FilteredLevelWriter{
			Writer: zerolog.LevelWriterAdapter{Writer: zerolog.NewConsoleWriter()},
			Level:  zerolog.InfoLevel,
		},
		&zerolog.FilteredLevelWriter{
			Writer: zerolog.LevelWriterAdapter{Writer: logFile},
			Level:  zerolog.DebugLevel,
		},
	}
	writer := zerolog.MultiLevelWriter(writers...)
	log.Logger = zerolog.New(writer).Level(zerolog.DebugLevel).With().Logger()

	// Read in the config.json
	userConfiguration, err := utils.LoadConfiguration("config.json")
	if err != nil {
		// Was return, needs adding back at later date
		handleErrorAndExit(err)
	}

	// If any additional tick data is defined in config.json that is NOT in our tick definitions
	// we add it to ours, we do not overwrite what we have as ours is the primary trusted source
	// as agreed with the OMITTED team
	for instrument, tickSize := range userConfiguration.AssetTicks {
		_, exists := utils.AssetTicks[instrument]

		if !exists {
			utils.AssetTicks[instrument] = tickSize
		}
	}

	// Build the log file
	var logOfTrades = tradeLog.NewLog()

	log.Info().Msgf("Backtesting: %+v", userConfiguration.Keys())

	// Make a map of instrument name to trade data to store all the files in.
	var data = make(map[string]backtestData.Data)
	// Create a mutex and wait group to use for the concurrent reading of data
	var mutex = new(sync.Mutex)
	var wg = new(sync.WaitGroup)

	// Begin iteration of data files
	for instrumentName, instrumentConfig := range userConfiguration.Instruments {
		// Add one counter to the wait group
		wg.Add(1)
		// Assign instrument as a local variable to prevent loading issues with global scope
		localInstrumentName := instrumentName
		instrumentConfig := instrumentConfig
		// Create anon goroutine
		go func() {
			// Defer the wait group being decremented
			defer wg.Done()
			instrumentData, err := backtestData.Load(fmt.Sprintf("data/%s.csv", localInstrumentName))
			if err != nil {
				// Was return, needs adding back at later date
				// TODO ROB THIS NO LONGER WORKS AS ITS INSIDE OF A GOROUTINE
				handleErrorAndExit(err)
			}

			// Get the tick size from the mapping
			tickSize, ok := utils.AssetTicks[localInstrumentName]
			if !ok {
				log.Error().Str(
					"instrument",
					localInstrumentName,
				).Msg("Instrument not found in AssetTicks, add it to the JSON.")
				return
			}

			// Log message so user knows the time filtering
			log.Info().Str(
				"instrument",
				localInstrumentName,
			).Str(
				"startTime",
				userConfiguration.BacktestStartDate.Time.Format(time.RFC3339),
			).Str(
				"endTime",
				userConfiguration.BacktestEndDate.Time.Format(time.RFC3339),
			).Msg(
				"Inflating rows and filtering.",
			)

			// Calculate the long and short SMAs and populate the values to the data
			log.Info().Str("instrument", localInstrumentName).Msg("Calculating SMA values")
			err = instrumentData.CalculateSMA(instrumentConfig, tickSize)
			if err != nil {
				handleErrorAndExit(err)
			}
			log.Info().Str("instrument", localInstrumentName).Msg("Calculated SMA values")

			// Calculate the stochastic values and populate the values to the data
			log.Info().Str("instrument", localInstrumentName).Msg("Calculating Stochastic values")
			instrumentData.CalculateStochasticOscillator(
				instrumentConfig.StochasticKPeriods,
				instrumentConfig.StochasticDPeriods,
			)
			log.Info().Str("instrument", localInstrumentName).Msg("Calculated Stochastic values")

			// Calculate the unbroken highs and lows for each candle
			log.Info().Str("instrument", localInstrumentName).Msg("Calculating unbroken highs and lows")
			instrumentData.CalculateUnbrokenHighsLows(
				instrumentConfig.UnbrokenBoundaryLeftBars,
				instrumentConfig.UnbrokenBoundaryRightBars,
				instrumentConfig.UnbrokenBoundaryMemoryLimit,
				instrumentConfig.StochasticUpperBand,
				instrumentConfig.StochasticLowerBand,
			)

			log.Info().Str("instrument", localInstrumentName).Msg("Calculated Highs and Lows")

			// filterOldBoundaries times down to the start time and end time, default end time is tomorrow
			instrumentData, err = instrumentData.FilterByTimes(
				userConfiguration.BacktestStartDate.Time,
				userConfiguration.BacktestEndDate.Time,
			)
			if err != nil {
				// If the filtered slice is empty, skip it
				if errors.Is(err, backtestData.FilterEmptyError) {
					log.Warn().Str(
						"instrument",
						localInstrumentName,
					).Msg("Backtest data has been filtered to 0 rows. skipping")
				} else {
					log.Error().Str(
						"instrument",
						localInstrumentName,
					).Msgf(
						"Got unknown error on backtestData.FilterByTimes: %s",
						err.Error(),
					)
				}
				// TODO ROB THIS NO LONGER WORKS AS ITS INSIDE OF A GOROUTINE
				handleErrorAndExit(err)
			}

			log.Info().Str("instrument", localInstrumentName).Msg("Filtered by times.")

			if userConfiguration.WriteProcessedDataToFile {
				err = instrumentData.WriteToCSV(fmt.Sprintf("./%s.csv", localInstrumentName))
				if err != nil {
					handleErrorAndExit(err)
				}
			}

			// Lock the mutex, add the data to the map and then unlock the mutex to allow thread safety
			mutex.Lock()
			data[localInstrumentName] = instrumentData
			mutex.Unlock()
		}()
	}
	// Wait for all files to finish loading
	wg.Wait()

	if len(data) == 0 {
		// len(data) *is* zero.
		// Duped error message for now
		log.Error().Msgf("no backtester data was loaded. Please ensure that the data/ directory contains valid data files.")
		handleErrorAndExit(err)
	}

	for instrument, item := range data {
		var items = len(item)
		if items == 0 {
			log.Warn().Str(
				"instrument",
				instrument,
			).Msg(
				"instrument has no data rows, you are likely missing the instrument data file.",
			)
		} else {
			log.Info().Str(
				"instrument",
				instrument,
			).Msgf(
				"instrument has %d data rows",
				len(item),
			)
		}
	}

	// Begin iteration of instruments
	for instrument, historicalData := range data {
		log.Info().Str(
			"instrument",
			instrument,
		).Msg("Starting back testing")

		// Get the tick size from the mapping
		tickSize, ok := utils.AssetTicks[instrument]
		if !ok {
			log.Error().Str(
				"instrument",
				instrument,
			).Msg("Instrument not found in AssetTicks, add it to the JSON.")
			continue
		}

		// Get the instrument-specific configuration from userConfiguration
		instrumentConfig, configExists := userConfiguration.Instruments[instrument]
		if !configExists {
			log.Error().Str(
				"instrument",
				instrument,
			).Msg("Instrument configuration not found in userConfiguration.")
			continue
		}

		// Begin testing for each region/session
		for _, region := range utils.StandardConfiguration.Regions {
			log.Info().Str(
				"region",
				region.RegionName,
			).Msg("Starting back testing for region")

			// Subset the data into a slice of data's for each trade window.
			// Windows defined in ./internal/utils/region.go
			subsets, err := historicalData.SubsetDataForMultipleDays(
				region.MarketOpen,
				region.MarketClose,
			)
			if err != nil {
				log.Error().Str("error", err.Error()).Msg("Got error on generating subsets")
				continue
			}

			log.Info().Str(
				"instrument",
				instrument,
			).Msgf("Generated subsets")

			// Begin iteration of each subset
			for _, subset := range *subsets {
				tradeData := tradeConfig.GenerateTradesInWindow(
					subset,
					instrumentConfig,
					instrument,
					tickSize,
				)

				log.Debug().Msgf("%+v", tradeData)
				for _, trade := range tradeData {
					logOfTrades = tradeLog.AddRow(logOfTrades, trade)
				}
			}
		}
	}

	// Log outputs
	// Calculate ROI and Profit
	returnOnInvestment := logOfTrades.CalculateCumulativeProfit()
	totalProfit := (returnOnInvestment - 1) * 100

	// Calculate total wins and win percentage.
	totalTrades := len(*logOfTrades)
	totalWins := logOfTrades.TotalWins()
	winRate := 0.0
	profitValue := 0.0
	if totalWins > 0 {
		winRate = (float64(totalWins) / float64(totalTrades)) * 100
		profitValue = logOfTrades.CalculateProfitValue(userConfiguration.StartingBalance)
	}
	log.Info().Msgf("Cumulative profit percentage: %.2d%%", int64(totalProfit))
	log.Info().Msgf("Total RR value: %.2f", logOfTrades.SumTotalProfit())
	log.Info().Msgf("Cumulative profit value with a starting balance of %.2f:  %.2f",
		userConfiguration.StartingBalance,
		profitValue,
	)
	log.Info().Msgf(
		"Trades taken: %d with %d wins for a winrate of %.2f%%",
		totalTrades,
		totalWins,
		winRate,
	)

	// Write log to disk
	err = tradeLog.Write(logOfTrades)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	log.Info().Msg("Computering finito.")

	// waitForKeyPress waits for the user to press any key before continuing.
	waitForKeyPress()
}

// handleErrorAndExit logs the error and waits for a keypress before exiting.
func handleErrorAndExit(err error) {
	fmt.Println("An error occurred, press any key to exit...")
	log.Error().Msg(err.Error()) // Log the actual error after termbox init to ensure it's visible.
	waitForKeyPress()
	os.Exit(1) // Exit with a non-zero status code to indicate an error.
}

// waitForKeyPress waits for the user to press any key before continuing.
func waitForKeyPress() {
	fmt.Println("Press 'Enter' to continue...")
	_, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return
	}
}
