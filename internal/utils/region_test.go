package utils

import (
	"fmt"
	"testing"
)

func TestRegion_String(t *testing.T) {
	r := region{
		RegionName:  "London",
		MarketOpen:  "07:00",
		MarketClose: "13:00",
	}
	stringOutput := fmt.Sprintln(r)
	expectedOutput := fmt.Sprintln("RegionName: London MarketOpen: 07:00 MarketClose: 13:00")

	if stringOutput != expectedOutput {
		t.Errorf(
			"regionString does not match expected output. \n want: %s \n got: %s",
			expectedOutput,
			stringOutput,
		)
	}
}
