package data_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

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

func TestCalendar_GetEventsForDay_ExcludedDate_ReturnsNoExcludedEvents(t *testing.T) {

	//	Arrange
	testRRuleCalendar := `BEGIN:VCALENDAR
PRODID:-//Google Inc//Google Calendar 70.9054//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-CALNAME:Family calendar
X-WR-TIMEZONE:America/New_York
X-WR-CALDESC:Esparza family events
BEGIN:VTIMEZONE
TZID:America/Grand_Turk
X-LIC-LOCATION:America/Grand_Turk
BEGIN:STANDARD
TZOFFSETFROM:-0400
TZOFFSETTO:-0500
TZNAME:EST
DTSTART:19701101T020000
RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU
END:STANDARD
BEGIN:DAYLIGHT
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
DTSTART:19700308T020000
RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU
END:DAYLIGHT
END:VTIMEZONE
BEGIN:VTIMEZONE
TZID:America/New_York
X-LIC-LOCATION:America/New_York
BEGIN:DAYLIGHT
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
DTSTART:19700308T020000
RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU
END:DAYLIGHT
BEGIN:STANDARD
TZOFFSETFROM:-0400
TZOFFSETTO:-0500
TZNAME:EST
DTSTART:19701101T020000
RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU
END:STANDARD
END:VTIMEZONE
BEGIN:VTIMEZONE
TZID:America/Phoenix
X-LIC-LOCATION:America/Phoenix
BEGIN:STANDARD
TZOFFSETFROM:-0700
TZOFFSETTO:-0700
TZNAME:MST
DTSTART:19700101T000000
END:STANDARD
END:VTIMEZONE
BEGIN:VEVENT
DTSTART;TZID=America/New_York:20201220T173000
DTEND;TZID=America/New_York:20201220T183000
EXDATE;TZID=America/New_York:20210425T173000
EXDATE;TZID=America/New_York:20211024T173000
EXDATE;TZID=America/New_York:20211031T173000
EXDATE;TZID=America/New_York:20211121T173000
EXDATE;TZID=America/New_York:20211128T173000
EXDATE;TZID=America/New_York:20220220T173000
RRULE:FREQ=WEEKLY
DTSTAMP:20220220T161319Z
UID:05F28281-077F-4059-971E-40E43F8AB3B5
URL:https://us02web.zoom.us/j/9288411040?pwd=ZVZFWVNGUWc4UHVzaHRKK010dGwrdz
	09
CREATED:20201220T170112Z
DESCRIPTION:
LAST-MODIFIED:20220205T221812Z
LOCATION:
SEQUENCE:0
STATUS:CONFIRMED
SUMMARY:Esparza family conference call 📲
TRANSP:OPAQUE
BEGIN:VALARM
ACTION:NONE
TRIGGER;VALUE=DATE-TIME:19760401T005545Z
END:VALARM
END:VEVENT
BEGIN:VEVENT
DTSTART;VALUE=DATE:20220218
DTEND;VALUE=DATE:20220222
DTSTAMP:20220220T161319Z
UID:6952A06D-6C4A-46D2-83DC-427A0FC5F53B
CREATED:20211021T173647Z
DESCRIPTION:
LAST-MODIFIED:20211021T173647Z
LOCATION:
SEQUENCE:0
STATUS:CONFIRMED
SUMMARY:Natalie’s Dress Shopping
TRANSP:OPAQUE
END:VEVENT
END:VCALENDAR
`
	timezone := "America/New_York"
	location, err := time.LoadLocation(timezone)
	if err != nil {
		t.Errorf("Error setting timezone: %v", err)
	}

	startTime := time.Date(2022, 2, 20, 0, 0, 0, 0, location)
	endTime := time.Date(2022, 2, 20, 23, 59, 59, 0, location)

	//	Act
	events, err := data.GetEventsForDay(strings.NewReader(testRRuleCalendar), startTime, endTime, location)
	if err != nil {
		t.Errorf("Error getting events for day: %v", err)
	}

	//	Assert
	//	We should see:
	//	- Natalie's Dress Shopping
	// We should not see:
	//	- Esparza family conference call (the date given has been excluded)
	// t.Logf("%+v", events)
	if len(events) != 1 {
		t.Errorf("We should have gotten exactly 1 event returned.  Instead, we got: %v", events)
	}

}
