package data_test

import (
	"log"
	"testing"
	"time"

	"github.com/danesparza/icaltoday/data"
)

func TestTZ_GetTime_ValidForNewYork(t *testing.T) {
	//	Arrange
	timezone := "America/New_York"

	//	Act
	if err := data.SetTimezone(timezone); err != nil {
		log.Fatal(err) // most likely timezone not loaded in Docker OS
	}
	timeResponse := data.GetTime(time.Now())
	response := timeResponse.String()

	//	Assert
	t.Logf("Returned time for %v: %+v", timezone, response)

}

func TestTZ_GetTime_ValidForLosAngeles(t *testing.T) {
	//	Arrange
	timezone := "America/Los_Angeles"

	//	Act
	if err := data.SetTimezone(timezone); err != nil {
		log.Fatal(err) // most likely timezone not loaded in Docker OS
	}
	timeResponse := data.GetTime(time.Now())
	response := timeResponse.String()

	//	Assert
	t.Logf("Returned time for %v: %+v", timezone, response)

}
