package backtestData

import (
	"errors"
	"fmt"
	"math"
	"os"
	"slices"
	"sort"
	"time"

	"github.com/BenHiramTaylor/strongbow-backtester/internal/utils"
	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog/log"
)

var (
	// DataEmptyError is an error for when a slice of data is empty.
	DataEmptyError = errors.New("slice is empty")
	// FilterEmptyError is an error for when we have successfully filtered a slice of data, but down to 0 results
	FilterEmptyError = errors.New("filtered data is empty")
	// EarliestTimeOutOfRange is an error for GetEarliestTime.
	// This is very specific on 100 years in the future because that is the start point we use in GetEarliestTime
	// to filter everything below that, if is returns 100 years in the future then it means that nothing is before that
	EarliestTimeOutOfRange = errors.New("earliest time found is one hundred years in the future")
	// SMALookbackInvalid is a custom error for when either of the SMA lookback amounts defined are less than or equal to 0.
	SMALookbackInvalid = errors.New("sma lookback amount is invalid, cannot be less than or equal to 0")
	// SMAValuesIntersect is an error for a very rare situation where the SMA values intersect with each-other
	SMAValuesIntersect = errors.New("sma values intersect with each other, do not trade")
	// NoBoundaryFound is a custom error for when we find no boundary in a search, usually if the boundary is empty.
	NoBoundaryFound = errors.New("no boundary found in filter")
)

// Boundary is a struct to store unbroken highs and lows in,
// this contains the price of the high/low and the index of the row.
type Boundary struct {
	// Time is the time of the candle that the boundary was defined at.
	Time time.Time

	// Value is the Close of the Row that forms this unbroken high/low
	Value float64

	// Broken is a boolean to show if the boundary is broken or not.
	Broken bool
}

// DeepCopy is a function to return a new Boundary pointer, this is due to updating this object
// in our logic and having it back-fill.
func (b *Boundary) DeepCopy() *Boundary {
	return &Boundary{
		Time:   b.Time,
		Value:  b.Value,
		Broken: b.Broken,
	}
}

// Boundaries is a slice of Boundary, used to track multiple unbroken price levels.
type Boundaries []*Boundary

// DeepCopy is a function that performs a deep copy of all boundaries in the slice.
func (b *Boundaries) DeepCopy() *Boundaries {
	// Create the return object
	var newBoundaries Boundaries

	for _, boundary := range *b {
		newBoundaries = append(newBoundaries, boundary.DeepCopy())
	}

	return &newBoundaries
}

// SortBoundaryByValue is a function that returns a copy of boundary sorted in either
// ascending or descending order by value
func (b *Boundaries) SortBoundaryByValue(ascending bool) *Boundaries {
	// Create a deepCopy of the whole slice of boundaries
	newBoundaries := b.DeepCopy()

	if ascending {
		// Sort the boundaries by value in ascending order.
		sort.Slice(*newBoundaries, func(i, j int) bool {
			return (*newBoundaries)[i].Value < (*newBoundaries)[j].Value
		})
	} else {
		// Sort the boundaries by value in descending order.
		sort.Slice(*newBoundaries, func(i, j int) bool {
			return (*newBoundaries)[i].Value > (*newBoundaries)[j].Value
		})
	}

	return newBoundaries
}

// GetSortedUnbrokenBoundary returns a copy of the boundaries, sorted in value order with broken = removed
func (b *Boundaries) GetSortedUnbrokenBoundary(ascending bool) (*Boundaries, error) {
	// Return custom error if no boundaries at beginning
	if len(*b) == 0 {
		return nil, NoBoundaryFound
	}

	// Create copy of slice
	var boundariesToReturn Boundaries

	// Remove the broken boundaries
	for _, boundary := range *b {
		// Skip Broken Boundaries
		if boundary.Broken {
			continue
		}

		boundariesToReturn = append(boundariesToReturn, boundary)
	}

	// Return custom error if no boundaries found after filtering
	if len(boundariesToReturn) == 0 {
		return nil, NoBoundaryFound
	}

	// Sort the boundaries
	sortedBoundaries := boundariesToReturn.SortBoundaryByValue(ascending)

	return sortedBoundaries, nil
}

// GetSortedBrokenBoundary returns a copy of the boundaries, sorted in value order with broken = removed
func (b *Boundaries) GetSortedBrokenBoundary(ascending bool) (*Boundaries, error) {
	// Return custom error if no boundaries at beginning
	if len(*b) == 0 {
		return nil, NoBoundaryFound
	}

	// Create copy of slice
	var boundariesToReturn Boundaries

	// Remove the broken boundaries
	for _, boundary := range *b {
		// Skip Broken Boundaries
		if !boundary.Broken {
			continue
		}

		boundariesToReturn = append(boundariesToReturn, boundary)
	}

	// Return custom error if no boundaries found after filtering
	if len(boundariesToReturn) == 0 {
		return nil, NoBoundaryFound
	}

	// Sort the boundaries
	sortedBoundaries := boundariesToReturn.SortBoundaryByValue(ascending)

	return sortedBoundaries, nil
}

// filterOldBoundaries removes boundaries from the slice by oldest first if they exceed the amount allowed in the list.
func (b *Boundaries) filterOldBoundaries(maxBoundaries int) {
	count := len(*b)
	if count <= maxBoundaries {
		return // No need to filter if the count is within the limit.
	}

	// Sort the boundaries by time in ascending order.
	sort.Slice(*b, func(i, j int) bool {
		return (*b)[i].Time.Before((*b)[j].Time)
	})

	// Keep only the newest maxBoundaries number of boundaries.
	// This slices off the oldest boundaries which are at the start of the slice.
	*b = (*b)[count-maxBoundaries:]
}

// updateBoundaries updates the slice of Boundaries with a new value.
func (b *Boundaries) updateBoundaries(newValue float64, time time.Time) {
	// Create the new boundary pointer with a default broken value of false
	newBoundary := &Boundary{Time: time, Value: newValue, Broken: false}
	// Prepend that to the front of the slice
	*b = append(Boundaries{newBoundary}, *b...)
}

// filterBrokenBoundaries removes boundaries that are broken by the current price.
func (b *Boundaries) filterBrokenBoundaries(currentValue float64, isHigh bool, maxBoundaries int) {
	var newBoundaries Boundaries

	for _, boundary := range *b {
		// If it is a high and the current boundary is greater than a previous one then it is broken
		// If it is a low and the current boundary is lower than a previous one then it has been broken
		isBoundaryBroken := (isHigh && currentValue > boundary.Value) || (!isHigh && currentValue < boundary.Value)

		// If it is broken then create a new boundary with broken = true
		// Also check if the boundary isn't already broken, as we do not want to copy across if it is
		if isBoundaryBroken && !boundary.Broken {
			// Create a deep copy of the boundary and mark it as broken
			brokenBoundary := boundary.DeepCopy()
			brokenBoundary.Broken = true
			newBoundaries = append(newBoundaries, brokenBoundary)
		}

		// If it is not broken previously or this time then copy it across
		if !boundary.Broken && !isBoundaryBroken {
			newBoundaries = append(newBoundaries, boundary)
		}
	}

	// Replace the boundaries with the new filtered slice
	*b = newBoundaries

	// Apply filter to limit the number of boundaries.
	b.filterOldBoundaries(maxBoundaries)
}

// Row is a struct that represents one candle of trade data
type Row struct {
	// Time represents the time of the candles CLOSE.
	Time time.Time `csv:"Time"`

	// Open represents the open price of the candles' interval.
	Open float64 `csv:"Open"`

	// High represents the highest price of the candles' interval.
	High float64 `csv:"High"`

	// Low represents the lowest price of the candles' interval.
	Low float64 `csv:"Low"`

	// Close represents the close price of the candles' interval.
	Close float64 `csv:"Close"`

	// Volume represents the volume of transactions in the candles' interval.
	Volume int `csv:"Volume"`

	// LargeSMA is a value to be calculated later for a larger length simple rolling Moving Average
	LargeSMA float64 `csv:"LargeSMA,omitempty"`

	// SmallSMA is a value to be calculated later for a smaller length simple rolling Moving Average
	SmallSMA float64 `csv:"SmallSMA,omitempty"`

	// HighBoundaries is a slice of highs boundaries for the candle.
	HighBoundaries Boundaries `csv:"UnbrokenHigh,omitempty"`

	// LowBoundaries is a slice of low boundaries for the candle.
	LowBoundaries Boundaries `csv:"UnbrokenLow,omitempty"`
}

// String is a way of formatting the Row
// this is used because otherwise pointer address' are printed.
func (r Row) String() string {
	return fmt.Sprintf(
		"Time: %v, Open: %f, High: %f, Low: %f, Close: %f, Volume: %b, LargeSMA: %f, SmallSMA: %f, "+
			"HighBoundaries: %v, LowBoundaries: %v",
		r.Time,
		r.Open,
		r.High,
		r.Low,
		r.Close,
		r.Volume,
		r.LargeSMA,
		r.SmallSMA,
		r.HighBoundaries,
		r.LowBoundaries,
	)
}

// IsGreenCandle determines if the Row represents a green candle (close > open).
func (r Row) IsGreenCandle() bool {
	return r.Close > r.Open
}

// IsRedCandle determines if the Row represents a red candle (close < open).
func (r Row) IsRedCandle() bool {
	return r.Close < r.Open
}

// TradeDirection is a getter function for figuring out the trade direction based on the SMA
// This function returns either "SHORT" OR "LONG" based on the definition of utils.TradeDirection
// In the rare event that the SMA values are exactly the same on an intersection then it returns a
// custom SMAValuesIntersect error
func (r Row) TradeDirection() (string, error) {
	switch {
	case r.SmallSMA == r.LargeSMA:
		return "", SMAValuesIntersect
	case r.SmallSMA < r.LargeSMA:
		return utils.TradeDirection.SHORT, nil
	default:
		return utils.TradeDirection.LONG, nil
	}
}

// IsValidEntry returns a boolean for if a candle is valid for a trade.
// In the event of a short a valid trade entry candle would wick above a previous high, but close below
// In the event of a long a valid trade would wick below the previous low but close above.
func (r Row) IsValidEntry(tradeDirection string) bool {
	// If the direction is SHORT
	if tradeDirection == utils.TradeDirection.SHORT {
		// Get the most recent broken high to check if we can enter, sort by ascending as we want the lowest high first
		filteredBoundaries, err := r.HighBoundaries.GetSortedBrokenBoundary(true)
		if err != nil {
			log.Debug().Msg("No broken high boundary found")
			return false
		}

		// Get the first filtered boundary
		mostRecentHighBoundary := (*filteredBoundaries)[0]

		// if it has wicked above but closed below the most recent high.
		if r.High > mostRecentHighBoundary.Value && r.Close < mostRecentHighBoundary.Value {
			return true
		}
	} else if tradeDirection == utils.TradeDirection.LONG {
		// Get the most recent broken low to check if we can enter, sort by descending as we want the highest low first
		filteredBoundaries, err := r.LowBoundaries.GetSortedBrokenBoundary(false)
		if err != nil {
			log.Debug().Msg("No broken low boundary found")
			return false
		}

		// Get the first filtered boundary
		mostRecentLowBoundary := (*filteredBoundaries)[0]

		// if it has wicked below but closed above the most recent low .
		if r.Low < mostRecentLowBoundary.Value && r.Close > mostRecentLowBoundary.Value {
			return true
		}
	}
	return false
}

// Data is a slice of Row memory pointers
type Data []*Row

// isPivotHigh checks if the Row at index i in Data is a pivot high.
func (d *Data) isPivotHigh(i, leftBars, rightBars int) bool {
	if i < leftBars || i+rightBars >= len(*d) {
		return false
	}

	currentHigh := (*d)[i].High
	for j := 1; j <= leftBars; j++ {
		if (*d)[i-j].High >= currentHigh {
			return false
		}
	}
	for j := 1; j <= rightBars; j++ {
		if (*d)[i+j].High >= currentHigh {
			return false
		}
	}
	return true
}

// isPivotLow checks if the Row at index i in Data is a pivot low.
func (d *Data) isPivotLow(i, leftBars, rightBars int) bool {
	if i < leftBars || i+rightBars >= len(*d) {
		return false
	}

	currentLow := (*d)[i].Low
	for j := 1; j <= leftBars; j++ {
		if (*d)[i-j].Low <= currentLow {
			return false
		}
	}
	for j := 1; j <= rightBars; j++ {
		if (*d)[i+j].Low <= currentLow {
			return false
		}
	}
	return true
}

// WriteToCSV writes the Data to a CSV file.
func (d *Data) WriteToCSV(filename string) error {
	// Create a new file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use gocsv to marshal data to CSV
	err = gocsv.MarshalFile(d, file)
	if err != nil {
		return err
	}

	// CSV writer is flushed automatically by the defer statement
	return nil
}

// Load reads the CSV file from the given file path and returns the data as a slice of DataRow.
func Load(filePath string) (Data, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			// Handle the error, but don't return it
			fmt.Printf("Error closing data file: %v\n", err)
		}
	}(file)

	var data Data

	if err := gocsv.UnmarshalFile(file, &data); err != nil { // Load data from file
		return nil, err
	}

	return data, nil
}

// CalculateSMA calculates the Simple Moving Averages for each Row in Data.
func (d *Data) CalculateSMA(config *utils.InstrumentConfiguration, tickSize float64) error {
	// Check if the lookback amounts are positive and non-zero and check if the data is not empty
	switch {
	case config.LargeSMALookbackAmount <= 0 || config.SmallSMALookbackAmount <= 0:
		return SMALookbackInvalid // Invalid lookback period, do nothing
	case len(*d) == 0:
		return DataEmptyError // Data is empty
	}

	// Iterate through each row in Data
	for i, row := range *d {
		// Calculate Large SMA
		largeLookback := min(i+1, config.LargeSMALookbackAmount)
		var sumLarge float64 // Sum of close prices for Large SMA calculation
		for j := i - largeLookback + 1; j <= i; j++ {
			sumLarge += (*d)[j].Close // Summing the Close price of the correct rows
		}
		largeSMA := sumLarge / float64(largeLookback)                 // Create the average
		row.LargeSMA = utils.RoundToDecimalLength(largeSMA, tickSize) // Assign it to the row

		// Calculate Small SMA
		smallLookback := min(i+1, config.SmallSMALookbackAmount)
		var sumSmall float64 // Sum of close prices for Small SMA calculation
		for j := i - smallLookback + 1; j <= i; j++ {
			sumSmall += (*d)[j].Close // Summing the Close price of the correct rows
		}
		smallSMA := sumSmall / float64(smallLookback)                 // Create the average
		row.SmallSMA = utils.RoundToDecimalLength(smallSMA, tickSize) // Assign it to the row
	}
	return nil
}

// CalculateUnbrokenHighsLows updates each Row in the Data slice with unbroken highs and lows.
// It uses lookback to determine the range for finding unbroken highs and lows,
// and memoryAmount to define how far back to keep track of these values.
func (d *Data) CalculateUnbrokenHighsLows(leftBars, rightBars, maxBoundaries int) {
	for i, row := range *d {
		// Check if this candle is a high or low pivot
		highPivot := d.isPivotHigh(i, leftBars, rightBars)
		lowPivot := d.isPivotLow(i, leftBars, rightBars)

		// Copy the boundaries from the previous candle if it is not the first row
		if i > 0 {
			previousRow := (*d)[i-1]
			row.HighBoundaries = slices.Clone(previousRow.HighBoundaries)
			row.LowBoundaries = slices.Clone(previousRow.LowBoundaries)
		}

		// If we are a highPivot then add it to the list of boundaries
		if highPivot {
			row.HighBoundaries.updateBoundaries(row.High, row.Time)
		}

		// If we are a lowPivot then add it to the list of boundaries
		if lowPivot {
			row.LowBoundaries.updateBoundaries(row.Low, row.Time)
		}

		// filterOldBoundaries broken boundaries for both highs and lows
		row.HighBoundaries.filterBrokenBoundaries(row.High, true, maxBoundaries)
		row.LowBoundaries.filterBrokenBoundaries(row.Low, false, maxBoundaries)
	}
}

// FilterByTimes takes a start and end time.Time object and
// returns a subset of data if it is between that time. Returns an error if no data is found.
func (d *Data) FilterByTimes(start, end time.Time) (Data, error) {
	var filteredData Data
	for _, row := range *d {
		if (row.Time.Equal(start) || row.Time.After(start)) && (row.Time.Equal(end) || row.Time.Before(end)) {
			filteredData = append(filteredData, row)
		}
	}
	if len(filteredData) == 0 {
		return filteredData, fmt.Errorf("no data found between %v and %v: %w", start, end, FilterEmptyError)
	}
	return filteredData, nil
}

// GetEarliestTime returns the earliest time from a slice of Data
// This is usually used before creating a new data slice to add an hour to a subset.
func (d *Data) GetEarliestTime() (time.Time, error) {
	if len(*d) == 0 {
		return time.Time{}, fmt.Errorf("error on GetEarliestTime for TradeData: %w", DataEmptyError)
	}

	// Start with a time value exactly one hundred years in the future as the minimum
	startTime := time.Now().UTC().Add(time.Hour * 24 * 365 * 100)
	earliestTime := startTime

	// Iterate over the slice and update the earliest time
	for _, row := range *d {
		if row.Time.Before(earliestTime) {
			earliestTime = row.Time
		}
	}

	if startTime.Equal(earliestTime) {
		return time.Time{}, fmt.Errorf("got error on GetEarliestTime %w", EarliestTimeOutOfRange)
	}

	return earliestTime, nil
}

// SubsetDataForMultipleDays returns a slice of Data blocks
// one for each trading window time that is not a weekend
func (d *Data) SubsetDataForMultipleDays(startTime, endTime string) (*[]Data, error) {
	var result []Data

	// Parse times to figure out spansMidnight
	tStart, err := time.Parse("15:04", startTime)
	if err != nil {
		return nil, err
	}
	tEnd, err := time.Parse("15:04", endTime)
	if err != nil {
		return nil, err
	}

	// Create a bool flag for if the time spans midnight
	spansMidnight := tStart.After(tEnd)

	// Set some default variables
	var subset Data
	isNextDay := false

	for _, row := range *d {
		// Skip weekends and unwanted weekdays
		if row.Time.Weekday() == time.Saturday || row.Time.Weekday() == time.Sunday {
			continue
		}

		// Alias this for repeatable use
		refDate := row.Time

		// Check if we are in the second of a two-day subset span
		// and if the boolean for spans midnight is true
		if isNextDay && spansMidnight {
			refDate = refDate.Add(-24 * time.Hour)
		}

		// Generate the start and end times to use for subsetting
		fullStartTime := time.Date(
			refDate.Year(),
			refDate.Month(),
			refDate.Day(),
			tStart.Hour(),
			tStart.Minute(),
			0,
			0,
			refDate.Location(),
		)
		fullEndTime := time.Date(
			refDate.Year(),
			refDate.Month(),
			refDate.Day(),
			tEnd.Hour(),
			tEnd.Minute(),
			0,
			0,
			refDate.Location(),
		)

		// Add another 24 hours if the start time is after the end time
		if fullStartTime.After(fullEndTime) {
			fullEndTime = fullEndTime.Add(24 * time.Hour)
		}

		// Check if the current time is within the last 5-minute interval of the day
		isLastIntervalOfDay := row.Time.Hour() == 23 && row.Time.Minute() >= 55

		// THIS IS THE COMPLEX CONDITIONAL LOGIC FOR THE MULTIPLE CASES OF ADDING A ROW TO A SUBSET
		// If spansMidnight is true (ADR) follow to that branch
		if spansMidnight {
			// If it is the first day of the two-day subset, and the time is after the start
			// (greater than 23:30 on the first day) then add it to the subset
			if !isNextDay && (row.Time.After(fullStartTime) || row.Time.Equal(fullStartTime)) {
				subset = append(subset, row)
				// If it is the second day of a two-day subset, and the time is either before or equal to the end time
				// (less than or equal to 06:00) then add it to the subset
			} else if isNextDay && (row.Time.Before(fullEndTime) || row.Time.Equal(fullEndTime)) {
				subset = append(subset, row)
			}

			// If it is the last interval of the day, and there is no nextDay
			// (we are on the second day) then reset some variables and continue
			if isLastIntervalOfDay && !isNextDay {
				isNextDay = true
				continue // We'll start collecting the next day's data in the next iteration
			}
			// If the time is after the startTime, and before or equal to the end time then add it to the subset
			// This is the normal path, ie after 08:00 and before or equal to 10:00
		} else if (row.Time.After(fullStartTime) || row.Time.Equal(fullStartTime)) && (row.Time.Before(fullEndTime) || row.Time.Equal(fullEndTime)) {
			subset = append(subset, row)
		}

		// If the time is equal to the final time and the subset has values then perform some checks
		if row.Time.Equal(fullEndTime) && subset != nil {
			// If the first and last times in a subset are not as expected
			// (not 08:00 and 10:00 or 23:00 and 06:00) then wipe the subset as incomplete and move on
			// We do this as depending on the users data, could be incomplete, this is expected to only happen
			// at the beginning of the data file and at the end, if either are mid-window
			if !subset[0].Time.Equal(fullStartTime) || !subset[len(subset)-1].Time.Equal(fullEndTime) {
				log.Debug().Msgf(
					"Skipping subset as incomplete, got a subset length of %s-%s",
					subset[0].Time,
					subset[len(subset)-1].Time,
				)
				subset = nil
				continue
			}

			// If the above skipping criteria are not met
			// then add the completed subset to the results and reset some variables for the next iteration
			result = append(result, subset)
			subset = nil
			isNextDay = false
		}
	}

	return &result, nil
}

// GetMaxHigh returns the highest High float64 in a data struct
func (d *Data) GetMaxHigh() float64 {
	high := -math.MaxFloat64

	for _, row := range *d {
		if row.High > high {
			high = row.High
		}
	}

	return high
}

// GetMinLow returns the Lowest Low float64 in a data struct
func (d *Data) GetMinLow() float64 {
	low := math.MaxFloat64

	for _, row := range *d {
		if row.Low < low {
			low = row.Low
		}
	}

	return low
}
