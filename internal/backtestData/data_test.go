package backtestData

import (
	"github.com/BenHiramTaylor/strongbow-backtester/internal/utils"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

// Define a helper function to create Boundary instances
func createBoundary(timeStr string, value float64, broken bool) *Boundary {
	parsedTime, _ := time.Parse("2006-01-02", timeStr)
	return &Boundary{Time: parsedTime, Value: value, Broken: broken}
}

// TestFilterOldBoundaries tests the filterOldBoundaries function of the Boundaries type.
func TestFilterOldBoundaries(t *testing.T) {
	// Enable parallel execution of tests
	t.Parallel()

	// Define test cases
	tests := []struct {
		name           string
		boundaries     Boundaries
		maxBoundaries  int
		expectedResult Boundaries
	}{
		{
			name: "Filtering with more boundaries than max",
			boundaries: Boundaries{
				createBoundary("2024-01-01", 100, false),
				createBoundary("2024-01-03", 300, false),
				createBoundary("2024-01-02", 200, false),
				createBoundary("2024-01-04", 400, false),
			},
			maxBoundaries: 2,
			expectedResult: Boundaries{
				createBoundary("2024-01-03", 300, false),
				createBoundary("2024-01-04", 400, false),
			},
		},
		{
			name: "No filtering needed as count equals maxBoundaries",
			boundaries: Boundaries{
				createBoundary("2024-01-01", 100, false),
				createBoundary("2024-01-02", 200, false),
			},
			maxBoundaries: 2,
			expectedResult: Boundaries{
				createBoundary("2024-01-01", 100, false),
				createBoundary("2024-01-02", 200, false),
			},
		},
		{
			name:           "No boundaries to filter",
			boundaries:     Boundaries{},
			maxBoundaries:  2,
			expectedResult: Boundaries{},
		},
	}

	// Run each test case
	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Call filterOldBoundaries
			tc.boundaries.filterOldBoundaries(tc.maxBoundaries)

			// Check if the result matches the expected result using assert.Equal
			assert.EqualValues(t, tc.expectedResult, tc.boundaries, "Test %s failed", tc.name)
		})
	}
}

func TestUpdateBoundaries(t *testing.T) {
	// Define empty boundaries
	boundaries := &Boundaries{}

	// Define values to add
	outputTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	outputValue := 100.0
	outputBoundaries := &Boundaries{
		&Boundary{Time: outputTime, Value: outputValue},
	}

	// Execute the function
	boundaries.updateBoundaries(outputValue, outputTime)

	// Validate the output
	assert.Equal(t, outputBoundaries, boundaries)
}

func TestFilterBrokenBoundaries(t *testing.T) {
	t.Parallel()

	timeForTests := time.Now()

	testCases := []struct {
		name               string
		initialBoundaries  Boundaries
		currentValue       float64
		isHigh             bool
		maxBoundaries      int
		expectedBoundaries Boundaries
	}{
		{
			name: "NoBoundariesBroken",
			initialBoundaries: Boundaries{
				{Time: timeForTests.Add(-time.Hour), Value: 101.0},
				{Time: timeForTests, Value: 100.0},
			},
			currentValue:  99.0,
			isHigh:        true,
			maxBoundaries: 2,
			expectedBoundaries: Boundaries{
				{Time: timeForTests.Add(-time.Hour), Value: 101.0},
				{Time: timeForTests, Value: 100.0},
			},
		},
		{
			name: "AllBoundariesBroken",
			initialBoundaries: Boundaries{
				{Time: timeForTests.Add(-time.Hour), Value: 101.0},
				{Time: timeForTests, Value: 100.0},
			},
			currentValue:  102.0,
			isHigh:        true,
			maxBoundaries: 2,
			expectedBoundaries: Boundaries{
				{Time: timeForTests.Add(-time.Hour), Value: 101.0, Broken: true},
				{Time: timeForTests, Value: 100.0, Broken: true},
			},
		},
		{
			name: "SomeBoundariesBroken",
			initialBoundaries: Boundaries{
				{Time: timeForTests.Add(-2 * time.Hour), Value: 99.0},
				{Time: timeForTests.Add(-time.Hour), Value: 100.0},
				{Time: timeForTests, Value: 101.0},
			},
			currentValue:  100.5,
			isHigh:        true,
			maxBoundaries: 3,
			expectedBoundaries: Boundaries{
				{Time: timeForTests.Add(-2 * time.Hour), Value: 99.0, Broken: true},
				{Time: timeForTests.Add(-time.Hour), Value: 100.0, Broken: true},
				{Time: timeForTests, Value: 101.0, Broken: false},
			},
		},
		{
			name: "OldBrokenBoundariesRemoved",
			initialBoundaries: Boundaries{
				{Time: timeForTests.Add(-2 * time.Hour), Value: 99.0, Broken: true}, // Will be removed
				{Time: timeForTests.Add(-time.Hour), Value: 100.0, Broken: true},    // Will be removed
				{Time: timeForTests, Value: 101.0, Broken: false},                   // Will remain
			},
			currentValue:  100.5,
			isHigh:        true,
			maxBoundaries: 3,
			expectedBoundaries: Boundaries{
				{Time: timeForTests, Value: 101.0, Broken: false},
			},
		},
		// Additional test cases can be added as needed.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.initialBoundaries.filterBrokenBoundaries(tc.currentValue, tc.isHigh, tc.maxBoundaries)
			assert.EqualValues(t, tc.expectedBoundaries, tc.initialBoundaries)
		})
	}
}

// TestIsGreenCandle is a series of tests for the TestIsGreenCandle function
func TestIsGreenCandle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		row     Row
		isGreen bool
	}{
		{
			name:    "GreenCandle",
			row:     Row{Open: 100.0, Close: 105.0},
			isGreen: true,
		},
		{
			name:    "RedCandle",
			row:     Row{Open: 105.0, Close: 100.0},
			isGreen: false,
		},
		// Add more test cases for edge cases like open equals close.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.row.IsGreenCandle()
			assert.Equal(t, tc.isGreen, result)
		})
	}
}

// TestIsRedCandle is a series of tests for the IsRedCandle function
func TestIsRedCandle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		row   Row
		isRed bool
	}{
		{
			name:  "RedCandle",
			row:   Row{Open: 105.0, Close: 100.0},
			isRed: true,
		},
		{
			name:  "GreenCandle",
			row:   Row{Open: 100.0, Close: 105.0},
			isRed: false,
		},
		// Add more test cases including edge cases.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.row.IsRedCandle()
			assert.Equal(t, tc.isRed, result)
		})
	}
}

// TestTradeDirection is a function for testing if a trade is in an expected direction based on price
func TestTradeDirection(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		row         Row
		expectedDir string
		expectedErr error
	}{
		{
			name:        "LongTrade",
			row:         Row{SmallSMA: 100.0, LargeSMA: 105.0},
			expectedDir: utils.TradeDirection.SHORT,
			expectedErr: nil,
		},
		{
			name:        "ShortTrade",
			row:         Row{SmallSMA: 105.0, LargeSMA: 100.0},
			expectedDir: utils.TradeDirection.LONG,
			expectedErr: nil,
		},
		{
			name:        "ErrorDueToIntersection",
			row:         Row{SmallSMA: 105.0, LargeSMA: 105.0},
			expectedDir: "",
			expectedErr: SMAValuesIntersect,
		},
		// Add more test cases including one for SMAValuesIntersect.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := tc.row.TradeDirection()
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedDir, dir)
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
		})
	}
}

// TestIsValidEntry is a series of tests for if a candle is a valid entry based on its previous boundaries.
func TestIsValidEntry(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		row            Row
		tradeDirection string
		isValid        bool
	}{
		{
			name: "ValidShortEntry",
			row: Row{
				High:           110.0,
				Close:          95.0,
				HighBoundaries: Boundaries{{Time: time.Now(), Value: 100.0, Broken: true}},
				LowBoundaries:  Boundaries{{Time: time.Now(), Value: 0}},
			},
			tradeDirection: utils.TradeDirection.SHORT,
			isValid:        true,
		},
		{
			name: "InvalidShortEntry",
			row: Row{
				High:           95.0,
				Close:          90.0,
				HighBoundaries: Boundaries{{Time: time.Now(), Value: 100.0}},
				LowBoundaries:  Boundaries{{Time: time.Now(), Value: 0}}, // Assuming irrelevant in this test case.
			},
			tradeDirection: utils.TradeDirection.SHORT,
			isValid:        false,
		},
		{
			name: "ValidLongEntry",
			row: Row{
				Low:            95.0,
				Close:          110.0,
				HighBoundaries: Boundaries{{Time: time.Now(), Value: 0}},
				LowBoundaries:  Boundaries{{Time: time.Now(), Value: 100.0, Broken: true}},
			},
			tradeDirection: utils.TradeDirection.LONG,
			isValid:        true,
		},
		{
			name: "InvalidLongEntry",
			row: Row{
				High:           95.0,
				Close:          90.0,
				HighBoundaries: Boundaries{{Time: time.Now(), Value: 0}},
				LowBoundaries:  Boundaries{{Time: time.Now(), Value: 100.0}},
			},
			tradeDirection: utils.TradeDirection.LONG,
			isValid:        false,
		},
		{
			name: "NoHighBoundaryFound",
			row: Row{
				High:           95.0,
				Close:          90.0,
				HighBoundaries: Boundaries{{Time: time.Now(), Value: 100.0}},
				LowBoundaries:  Boundaries{{Time: time.Now(), Value: 100.0, Broken: true}},
			},
			tradeDirection: "",
			isValid:        false,
		},
		{
			name: "NoLowBoundaryFound",
			row: Row{
				High:           95.0,
				Close:          90.0,
				HighBoundaries: Boundaries{{Time: time.Now(), Value: 100.0, Broken: true}},
				LowBoundaries:  Boundaries{{Time: time.Now(), Value: 100.0}},
			},
			tradeDirection: "",
			isValid:        false,
		},
		{
			name: "EverythingValidButNoSetup",
			row: Row{
				High:           95.0,
				Close:          90.0,
				HighBoundaries: Boundaries{{Time: time.Now(), Value: 100.0, Broken: true}},
				LowBoundaries:  Boundaries{{Time: time.Now(), Value: 100.0, Broken: true}},
			},
			tradeDirection: utils.TradeDirection.SHORT,
			isValid:        false,
		},
		// Add more test cases for long entries, edge cases, and no boundaries.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.row.IsValidEntry(tc.tradeDirection) // or LONG based on the case
			assert.Equal(t, tc.isValid, result)
		})
	}
}

// TestRowString tests printing the string of a row
func TestRowString(t *testing.T) {
	sampleRow := Row{
		Time:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Open:     1,
		Close:    2,
		High:     3,
		Low:      4,
		Volume:   5,
		LargeSMA: 6,
		SmallSMA: 7,
	}

	output := sampleRow.String()
	expectedOutput := "Time: 2020-01-01 00:00:00 +0000 UTC, Open: 1.000000, High: 3.000000, Low: 4.000000, " +
		"Close: 2.000000, Volume: 101, LargeSMA: 6.000000, SmallSMA: 7.000000, HighBoundaries: [], LowBoundaries: []"

	if output != expectedOutput {
		t.Errorf("did not get expected output of row string, got: %s want: %s", output, expectedOutput)
	}
}

// TestIsPivotHigh tests the isPivotHigh function of the Data struct.
func TestIsPivotHigh(t *testing.T) {
	// Enable parallel execution of tests
	t.Parallel()

	// Define a helper function to create test Rows with only the High value
	createRowHigh := func(high float64) *Row {
		return &Row{High: high}
	}

	// Define test cases for isPivotHigh
	tests := []struct {
		name      string
		data      Data
		index     int
		leftBars  int
		rightBars int
		expected  bool
	}{
		{
			name: "Valid pivot high",
			data: Data{
				createRowHigh(100),
				createRowHigh(200),
				createRowHigh(300), // Pivot high
				createRowHigh(200),
				createRowHigh(100),
			},
			index:     2,
			leftBars:  2,
			rightBars: 2,
			expected:  true,
		},
		{
			name: "Invalid pivot high - equal high in leftBars",
			data: Data{
				createRowHigh(100),
				createRowHigh(300), // Same high in leftBars
				createRowHigh(300), // Potential pivot high
				createRowHigh(200),
				createRowHigh(100),
			},
			index:     2,
			leftBars:  2,
			rightBars: 2,
			expected:  false,
		},
		{
			name: "Invalid pivot high - equal high in rightBars",
			data: Data{
				createRowHigh(100),
				createRowHigh(200),
				createRowHigh(300), // Potential pivot high
				createRowHigh(200),
				createRowHigh(300), // Same high in rightBars
			},
			index:     2,
			leftBars:  2,
			rightBars: 2,
			expected:  false,
		},
		{
			name: "Not enough candles either side",
			data: Data{
				createRowHigh(300),
			},
			index:     2,
			leftBars:  2,
			rightBars: 2,
			expected:  false,
		},
		// Add more test cases for edge cases and other scenarios
	}

	// Run each test case for isPivotHigh
	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.data.isPivotHigh(tc.index, tc.leftBars, tc.rightBars)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestIsPivotLow tests the isPivotLow function of the Data struct.
func TestIsPivotLow(t *testing.T) {
	// Enable parallel execution of tests
	t.Parallel()

	// Define a helper function to create test Rows with only the Low value
	createRowLow := func(low float64) *Row {
		return &Row{Low: low}
	}

	// Define test cases for isPivotLow
	tests := []struct {
		name      string
		data      Data
		index     int
		leftBars  int
		rightBars int
		expected  bool
	}{
		{
			name: "Valid pivot low",
			data: Data{
				createRowLow(300),
				createRowLow(200),
				createRowLow(100), // Pivot low
				createRowLow(200),
				createRowLow(300),
			},
			index:     2,
			leftBars:  2,
			rightBars: 2,
			expected:  true,
		},
		{
			name: "Invalid pivot low - equal low in leftBars",
			data: Data{
				createRowLow(300),
				createRowLow(100), // Same low in leftBars
				createRowLow(100), // Potential pivot low
				createRowLow(200),
				createRowLow(300),
			},
			index:     2,
			leftBars:  2,
			rightBars: 2,
			expected:  false,
		},
		{
			name: "Invalid pivot low - equal low in rightBars",
			data: Data{
				createRowLow(300),
				createRowLow(200),
				createRowLow(100), // Potential pivot low
				createRowLow(200),
				createRowLow(100), // Same low in rightBars
			},
			index:     2,
			leftBars:  2,
			rightBars: 2,
			expected:  false,
		},
		{
			name: "Not enough candles either side",
			data: Data{
				createRowLow(300),
			},
			index:     2,
			leftBars:  2,
			rightBars: 2,
			expected:  false,
		},
		// Add more test cases for edge cases and other scenarios
	}

	// Run each test case for isPivotLow
	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.data.isPivotLow(tc.index, tc.leftBars, tc.rightBars)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestCalculateSMA tests the CalculateSMA function of the Data struct.
// It covers various scenarios including edge cases and typical usage.
func TestCalculateSMA(t *testing.T) {
	// Enable parallel execution of tests
	t.Parallel()

	// Define a helper function to create test Rows
	createRow := func(timeStr string, close float64) *Row {
		parsedTime, _ := time.Parse("2006-01-02", timeStr)
		return &Row{Time: parsedTime, Close: close}
	}

	// Create the valid outputs
	expectedValidSMAOutput := Data{
		createRow("2024-01-01", 100),
		createRow("2024-01-02", 200),
		createRow("2024-01-03", 300),
	}

	// Set the expected SMA outputs
	expectedValidSMAOutput[0].SmallSMA = 100
	expectedValidSMAOutput[0].LargeSMA = 100
	expectedValidSMAOutput[1].SmallSMA = 150
	expectedValidSMAOutput[1].LargeSMA = 150
	expectedValidSMAOutput[2].SmallSMA = 250
	expectedValidSMAOutput[2].LargeSMA = 200

	// Test cases
	tests := []struct {
		name        string
		data        Data
		config      *utils.InstrumentConfiguration
		tickSize    float64
		expectedSMA Data
		expectError bool
	}{
		{
			name: "Valid small and large SMA calculation",
			data: Data{
				createRow("2024-01-01", 100),
				createRow("2024-01-02", 200),
				createRow("2024-01-03", 300),
				// ... add more rows as needed
			},
			config: &utils.InstrumentConfiguration{
				LargeSMALookbackAmount: 3,
				SmallSMALookbackAmount: 2,
			},
			tickSize:    0.01,
			expectedSMA: expectedValidSMAOutput,
			expectError: false,
		},
		{
			name: "Invalid lookback period",
			data: Data{
				createRow("2020-01-01", 100),
			},
			config: &utils.InstrumentConfiguration{
				LargeSMALookbackAmount: -1, // Invalid lookback period
				SmallSMALookbackAmount: 2,
			},
			tickSize:    0.01,
			expectedSMA: Data{
				// ...
			},
			expectError: true,
		},
		{
			name: "Empty Data",
			data: Data{
				// Data empty
			},
			config: &utils.InstrumentConfiguration{
				LargeSMALookbackAmount: 1, // Invalid lookback period
				SmallSMALookbackAmount: 2,
			},
			tickSize:    0.01,
			expectedSMA: Data{
				// ...
			},
			expectError: true,
		},
		// Add more test cases for edge cases and other scenarios
	}

	// Run each test case
	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Perform the CalculateSMA operation
			err := tc.data.CalculateSMA(tc.config, tc.tickSize)

			// Check if an error was expected
			if (err != nil) != tc.expectError {
				t.Fatalf("Expected error: %v, got: %v", tc.expectError, err)
			}

			// If no error, validate the calculated SMA values
			if !tc.expectError {
				if !reflect.DeepEqual(tc.data, tc.expectedSMA) {
					t.Errorf("Expected SMA: %v, got: %v", tc.expectedSMA, tc.data)
				}
			}
		})
	}
}

func TestFilterByTimes(t *testing.T) {
	layout := "2006-01-02T15:04:05"
	start, _ := time.Parse(layout, "2023-10-01T00:00:00")
	end, _ := time.Parse(layout, "2023-10-01T02:00:00")

	data := Data{
		&Row{Time: start.Add(-time.Hour), Open: 1, High: 2, Low: 1, Close: 1.5, Volume: 100},
		&Row{Time: start, Open: 1.5, High: 2.5, Low: 1, Close: 2, Volume: 200},
		&Row{Time: start.Add(time.Hour), Open: 2, High: 3, Low: 1.5, Close: 2.5, Volume: 300},
		&Row{Time: end, Open: 2.5, High: 3.5, Low: 2, Close: 3, Volume: 400},
		&Row{Time: end.Add(time.Hour), Open: 3, High: 4, Low: 2.5, Close: 3.5, Volume: 500},
	}

	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		wantLen   int
		wantErr   bool
	}{
		{"Test valid range", start, end, 3, false},
		{"Test invalid range", end, start, 0, true},
		{"Test no data in range", end.Add(time.Hour * 5), end.Add(2 * time.Hour), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := data.FilterByTimes(tt.startTime, tt.endTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterByTimes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("FilterByTimes() got = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestGetEarliestTime(t *testing.T) {
	tests := []struct {
		name      string
		data      Data
		expected  time.Time
		expectErr bool
	}{
		{
			name:      "empty_data",
			data:      Data{},
			expected:  time.Time{},
			expectErr: true,
		},
		{
			name: "single_entry",
			data: Data{
				&Row{Time: time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC)},
			},
			expected:  time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC),
			expectErr: false,
		},
		{
			name: "in_order_times",
			data: Data{
				&Row{Time: time.Date(2023, 10, 9, 0, 0, 0, 0, time.UTC)},
				&Row{Time: time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC)},
			},
			expected:  time.Date(2023, 10, 9, 0, 0, 0, 0, time.UTC),
			expectErr: false,
		},
		{
			name: "out_of_order_times",
			data: Data{
				&Row{Time: time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC)},
				&Row{Time: time.Date(2023, 10, 9, 0, 0, 0, 0, time.UTC)},
			},
			expected:  time.Date(2023, 10, 9, 0, 0, 0, 0, time.UTC),
			expectErr: false,
		},
		{
			name: "edge_near_3000",
			data: Data{
				&Row{Time: time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC)},
				&Row{Time: time.Date(2998, 12, 31, 0, 0, 0, 0, time.UTC)},
			},
			expected:  time.Date(2998, 12, 31, 0, 0, 0, 0, time.UTC),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			earliestTime, err := tt.data.GetEarliestTime()
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("got unexpected error: %v", err)
				return
			}

			if !earliestTime.Equal(tt.expected) {
				t.Errorf("got %v, want %v", earliestTime, tt.expected)
			}
		})
	}
}

// mockOneDayRDRInputData generates a fixed day of input data
func mockOneDayRDRInputData() *Data {
	// Starting time
	startTime := time.Date(2020, 6, 5, 13, 30, 0, 0, time.UTC)
	endTime := time.Date(2020, 6, 5, 21, 0, 0, 0, time.UTC)

	var data Data

	// Let's generate data for 2 days from the starting time, at 5-minute intervals
	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
		data = append(data, &Row{
			Time:   t,
			Open:   1.0,
			High:   1.1,
			Low:    0.9,
			Close:  1.05,
			Volume: 100,
		})
	}

	return &data
}

// mockTwoDayRDRInputData generates two fixed days of input data
func mockTwoDayRDRInputData() *Data {
	// Starting time
	startTime := time.Date(2020, 6, 4, 13, 30, 0, 0, time.UTC)
	// End time is the day after
	endTime := time.Date(2020, 6, 5, 21, 0, 0, 0, time.UTC)

	var data Data

	// Let's generate data for 2 days from the starting time, at 5-minute intervals
	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
		data = append(data, &Row{
			Time:   t,
			Open:   1.0,
			High:   1.1,
			Low:    0.9,
			Close:  1.05,
			Volume: 100,
		})
	}

	return &data
}

// mockIncompleteInputData generates two hours of input data
func mockIncompleteInputData() *Data {
	// Starting time
	startTime := time.Date(2020, 6, 5, 13, 35, 0, 0, time.UTC)
	endTime := time.Date(2020, 6, 5, 15, 0, 0, 0, time.UTC)

	var data Data

	// Let's generate data for 2 days from the starting time, at 5-minute intervals
	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
		data = append(data, &Row{
			Time:   t,
			Open:   1.0,
			High:   1.1,
			Low:    0.9,
			Close:  1.05,
			Volume: 100,
		})
	}

	return &data
}

// mockOneDayADRInputData generates a static input for an ADR session
func mockOneDayADRInputData() *Data {
	var data Data
	// Starting and end times
	startTime := time.Date(2020, 6, 1, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2020, 6, 2, 9, 0, 0, 0, time.UTC)

	// Let's generate data to fill the start times at 5-minute intervals
	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
		data = append(data, &Row{
			Time:   t,
			Open:   1.0,
			High:   1.1,
			Low:    0.9,
			Close:  1.05,
			Volume: 100,
		})
	}

	return &data
}

// mockTwoDayADRInputData generates a static input for an ADR session
func mockTwoDayADRInputData() *Data {
	var data Data
	// Starting and end times
	startTime := time.Date(2020, 6, 1, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2020, 6, 3, 9, 0, 0, 0, time.UTC)

	// Let's generate data to fill the start times at 5-minute intervals
	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
		data = append(data, &Row{
			Time:   t,
			Open:   1.0,
			High:   1.1,
			Low:    0.9,
			Close:  1.05,
			Volume: 100,
		})
	}

	return &data
}

// getADRData generates a static expected response for an ADR session
func getADRData() *[]Data {
	var data Data
	var result []Data
	// Starting and end times
	startTime := time.Date(2020, 6, 1, 23, 35, 0, 0, time.UTC)
	endTime := time.Date(2020, 6, 2, 6, 0, 0, 0, time.UTC)

	// Let's generate data to fill the start times at 5-minute intervals
	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
		data = append(data, &Row{
			Time:   t,
			Open:   1.0,
			High:   1.1,
			Low:    0.9,
			Close:  1.05,
			Volume: 100,
		})
	}
	// Add one subsets result to the results
	result = append(result, data)

	return &result
}

// getADRData generates a 2x static expected response for an ADR session
func getDoubleADRData() *[]Data {
	var result []Data
	var startTime time.Time
	var endTime time.Time
	// Starting and end times
	startTime = time.Date(2020, 6, 1, 23, 35, 0, 0, time.UTC)
	endTime = time.Date(2020, 6, 2, 6, 0, 0, 0, time.UTC)

	for i := 0; i < 2; i++ {
		var data Data
		// Let's generate data to fill the start times at 5-minute intervals
		for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
			data = append(data, &Row{
				Time:   t,
				Open:   1.0,
				High:   1.1,
				Low:    0.9,
				Close:  1.05,
				Volume: 100,
			})
		}
		// Add one subsets result to the results
		result = append(result, data)
		// Increment the window by a day
		startTime = startTime.Add(24 * time.Hour)
		endTime = endTime.Add(24 * time.Hour)
	}
	return &result
}

// getSingleRDRData generates a static expected response for an RDR session
func getSingleRDRData() *[]Data {
	var result []Data
	var data Data
	// Starting time
	startTime := time.Date(2020, 6, 5, 13, 35, 0, 0, time.UTC)
	endTime := time.Date(2020, 6, 5, 20, 0, 0, 0, time.UTC)

	// Let's generate data to fill the start times at 5-minute intervals
	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
		data = append(data, &Row{
			Time:   t,
			Open:   1.0,
			High:   1.1,
			Low:    0.9,
			Close:  1.05,
			Volume: 100,
		})
	}
	// Add one subsets result to the results
	result = append(result, data)
	return &result
}

// getDoubleRDRData generates a 2x static expected response for an RDR session
func getDoubleRDRData() *[]Data {
	var result []Data
	var startTime time.Time
	var endTime time.Time
	// Starting time
	startTime = time.Date(2020, 6, 4, 13, 35, 0, 0, time.UTC)
	endTime = time.Date(2020, 6, 4, 20, 0, 0, 0, time.UTC)

	for i := 0; i < 2; i++ {
		var data Data

		// Let's generate data to fill the start times at 5-minute intervals
		for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(5 * time.Minute) {
			data = append(data, &Row{
				Time:   t,
				Open:   1.0,
				High:   1.1,
				Low:    0.9,
				Close:  1.05,
				Volume: 100,
			})
		}

		// Add one subsets result to the results
		result = append(result, data)
		// Increment the window by a day
		startTime = startTime.Add(24 * time.Hour)
		endTime = endTime.Add(24 * time.Hour)
	}
	return &result
}

// TestSubsetDataForMultipleDays is a complex test, it tests for the amount of subsets returned.
// It also tests for the values of those subsets due to the complexity of the Subset function.
func TestSubsetDataForMultipleDays(t *testing.T) {
	tests := []struct {
		name       string
		startTime  string
		endTime    string
		want       int // number of subsets
		wantValue  *[]Data
		sampleData *Data
		expectErr  bool
	}{
		{
			"Test RDR one day one subset",
			"13:35",
			"20:00",
			1,
			getSingleRDRData(),
			mockOneDayRDRInputData(),
			false,
		},
		{
			"Test RDR two days two subsets",
			"13:35",
			"20:00",
			2,
			getDoubleRDRData(),
			mockTwoDayRDRInputData(),
			false,
		},
		{
			"Test within range with wrong dayNames",
			"08:00",
			"10:00",
			0,
			nil,
			mockOneDayRDRInputData(),
			false,
		},
		{
			"Test ADR cross days one subset",
			"23:35",
			"06:00",
			1,
			getADRData(),
			mockOneDayADRInputData(),
			false,
		},
		{
			"Test ADR cross days two subsets",
			"23:35",
			"06:00",
			2,
			getDoubleADRData(),
			mockTwoDayADRInputData(),
			false,
		},
		{
			"invalid start time",
			"29000",
			"06:00",
			0,
			getSingleRDRData(),
			mockOneDayRDRInputData(),
			true,
		},
		{
			"invalid end time",
			"21:00",
			"1234",
			0,
			getSingleRDRData(),
			mockOneDayRDRInputData(),
			true,
		},
		{
			"Test incomplete subset",
			"12:00",
			"15:00",
			0,
			getSingleRDRData(),
			mockIncompleteInputData(),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subsets, err := tt.sampleData.SubsetDataForMultipleDays(tt.startTime, tt.endTime)
			// Handle errors
			if err != nil {
				// If we do not expect an error
				if !tt.expectErr {
					t.Errorf("got error when one was not expected, %s", err.Error())
				}
				return
			}

			// Check if the length of the output is the length of the expected data
			if len(*subsets) != tt.want {
				t.Errorf("got %v subsets, want %v subsets", len(*subsets), tt.want)
			}
			// check how manny subsets are wanted
			if tt.want > 0 {
				// If we want subsets, check they match the expected outputs
				if !reflect.DeepEqual(subsets, tt.wantValue) {
					t.Errorf("got %+v subsets \n want %+v subsets", subsets, tt.wantValue)
				}
			}
		})
	}
}

func TestGetMaxHigh(t *testing.T) {
	tests := []struct {
		name string
		data Data
		want float64
	}{
		{
			name: "single row",
			data: Data{
				&Row{Time: time.Now(), Open: 1.0, High: 2.0, Low: 0.5, Close: 1.5, Volume: 100},
			},
			want: 2.0,
		},
		{
			name: "multiple rows",
			data: Data{
				&Row{Time: time.Now(), Open: 1.0, High: 2.0, Low: 0.5, Close: 1.5, Volume: 100},
				&Row{Time: time.Now(), Open: 2.0, High: 4.0, Low: 1.5, Close: 3.0, Volume: 200},
				&Row{Time: time.Now(), Open: 3.0, High: 3.5, Low: 2.5, Close: 3.0, Volume: 300},
			},
			want: 4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.GetMaxHigh(); got != tt.want {
				t.Errorf("GetMaxHigh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMinLow(t *testing.T) {
	tests := []struct {
		name string
		data Data
		want float64
	}{
		{
			name: "single row",
			data: Data{
				&Row{Time: time.Now(), Open: 1.0, High: 2.0, Low: 0.5, Close: 1.5, Volume: 100},
			},
			want: 0.5,
		},
		{
			name: "multiple rows",
			data: Data{
				&Row{Time: time.Now(), Open: 1.0, High: 2.0, Low: 0.5, Close: 1.5, Volume: 100},
				&Row{Time: time.Now(), Open: 2.0, High: 4.0, Low: 1.5, Close: 3.0, Volume: 200},
				&Row{Time: time.Now(), Open: 3.0, High: 3.5, Low: 2.5, Close: 3.0, Volume: 300},
			},
			want: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.GetMinLow(); got != tt.want {
				t.Errorf("GetMinLow() = %v, want %v", got, tt.want)
			}
		})
	}
}
