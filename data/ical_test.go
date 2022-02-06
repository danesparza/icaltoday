package data_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/danesparza/icaltoday/data"
)

func TestCalendar_GetTodaysEvents_ReturnsValidData(t *testing.T) {
	//	Arrange
	service := data.CalService{}

	test_calendar_url := os.Getenv("ICALTODAY_CALENDAR_URL")
	if test_calendar_url == "" {
		t.Fatal("The environment variable ICALTODAY_CALENDAR_URL is blank and should be set to the calendar url to test")
	}

	test_timezone := os.Getenv("ICALTODAY_TIMEZONE")
	if test_calendar_url == "" {
		t.Fatal("The environment variable ICALTODAY_TIMEZONE is blank and should be set to the timezone to test (like America/New_York)")
	}

	url := test_calendar_url
	timezone := test_timezone

	ctx := context.Background()
	ctx, seg := xray.BeginSegment(ctx, "unit-test")
	defer seg.Close(nil)

	//	Act
	response, err := service.GetTodaysEvents(ctx, url, timezone)

	//	Assert
	if err != nil {
		t.Errorf("Error calling GetTodaysEvents: %v", err)
	}

	t.Logf("Returned object: %+v", response)

}
